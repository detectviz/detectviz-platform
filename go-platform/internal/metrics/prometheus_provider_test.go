package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockPrometheusAPI is a mock implementation of the Prometheus API for testing.
type mockPrometheusAPI struct {
	v1.API
	queryRangeFunc func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error)
}

func (m *mockPrometheusAPI) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	if m.queryRangeFunc != nil {
		return m.queryRangeFunc(ctx, query, r, opts...)
	}
	return nil, nil, errors.New("mock not implemented")
}

// newTestProvider creates a PrometheusProvider with a mock API for testing.
func newTestProvider(config *PrometheusConfig, mockAPI v1.API) (*PrometheusProvider, error) {
	if config == nil {
		config = &PrometheusConfig{}
	}
	// Use NewPrometheusProvider to get default values, but then override the client.
	provider, err := NewPrometheusProvider(config, zap.NewNop())
	if err != nil {
		return nil, err
	}
	provider.client = mockAPI
	return provider, nil
}

func TestPrometheusProvider_CircuitBreaker_StaysClosed(t *testing.T) {
	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return model.Matrix{}, nil, nil // Successful response
		},
	}

	config := &PrometheusConfig{
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 3,
			RecoveryTimeout:  1 * time.Second,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err := provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
		assert.NoError(t, err)
		assert.Equal(t, StateClosed, provider.cbState)
	}
}

func TestPrometheusProvider_CircuitBreaker_OpensOnFailure(t *testing.T) {
	failureError := errors.New("prometheus is down")
	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			return nil, nil, failureError
		},
	}

	config := &PrometheusConfig{
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 3,
			RecoveryTimeout:  10 * time.Second,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	// Fail 3 times to open the circuit breaker
	for i := 0; i < 3; i++ {
		_, err := provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
		assert.Error(t, err)
	}
	assert.Equal(t, StateOpen, provider.cbState, "Circuit breaker should be open after 3 failures")

	// The 4th call should fail immediately with a circuit breaker error
	_, err = provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestPrometheusProvider_CircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	failureError := errors.New("prometheus is down")
	var queryCount int
	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			queryCount++
			return nil, nil, failureError
		},
	}

	config := &PrometheusConfig{
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	// Fail 2 times to open it
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Equal(t, StateOpen, provider.cbState)
	assert.Equal(t, 2, queryCount)

	// This call should be rejected immediately
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Equal(t, 2, queryCount)

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// The next call is the half-open attempt. It should be executed.
	_, err = provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Error(t, err) // It will still fail because the mock API is failing.
	assert.Equal(t, 3, queryCount, "Query should be attempted after recovery timeout")

	// After a failure in the half-open state, the breaker should re-open.
	assert.Equal(t, StateOpen, provider.cbState, "Circuit breaker should re-open after a failed half-open attempt")
}

func TestPrometheusProvider_CircuitBreaker_ClosesOnSuccess(t *testing.T) {
	callCount := 0
	failureError := errors.New("prometheus is down")

	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			callCount++
			if callCount <= 2 { // Fail first two times
				return nil, nil, failureError
			}
			return model.Matrix{}, nil, nil // Succeed on the third call (half-open attempt)
		},
	}

	config := &PrometheusConfig{
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	// Fail twice to open
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Equal(t, StateOpen, provider.cbState)

	// Wait for recovery
	time.Sleep(150 * time.Millisecond)

	// This call should be in half-open state and succeed
	_, err = provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.NoError(t, err, "Query should succeed in half-open state")
	assert.Equal(t, StateClosed, provider.cbState, "Circuit breaker should be closed after success in half-open")

	// Subsequent calls should also succeed
	_, err = provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, provider.cbState)
}

func TestPrometheusProvider_CircuitBreaker_ReOpensOnFailure(t *testing.T) {
	callCount := 0
	failureError := errors.New("prometheus is down")

	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			callCount++
			// Always fail
			return nil, nil, failureError
		},
	}

	config := &PrometheusConfig{
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	// Fail twice to open
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Equal(t, StateOpen, provider.cbState)

	// Wait for recovery
	time.Sleep(150 * time.Millisecond)

	// This call is the half-open attempt, it will fail
	_, err = provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, provider.cbState, "Circuit breaker should be open again after failure in half-open")

	// It should still be open and reject the next call
	_, err = provider.Query(context.Background(), MetricQuery{Metric: "test_metric"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}
