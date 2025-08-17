package metrics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/testutil"
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

func TestPrometheusProvider_QueryTimeout(t *testing.T) {
	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			// This function now needs to respect the context's deadline.
			select {
			case <-time.After(100 * time.Millisecond):
				// Simulate work that takes too long
				return model.Matrix{}, nil, nil
			case <-ctx.Done():
				// The context timed out, as expected
				return nil, nil, ctx.Err()
			}
		},
	}

	provider, err := newTestProvider(&PrometheusConfig{}, mockAPI)
	assert.NoError(t, err)

	// Create a context with a timeout that is shorter than the mock's sleep time.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// The provider's Query method passes the context to executeQuery, which passes it to
	// executePrometheusQuery, which passes it to the client.QueryRange call.
	// The mock now simulates a real API by respecting the context.
	_, err = provider.Query(ctx, MetricQuery{Metric: "test_metric"})

	assert.Error(t, err, "Expected an error due to timeout")
	// The provider wraps the context error, so we check for the wrapped error.
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestPrometheusProvider_CacheConcurrency(t *testing.T) {
	// This test is most effective when run with the -race flag
	t.Parallel()

	var callCount int32
	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			atomic.AddInt32(&callCount, 1)
			return model.Matrix{}, nil, nil
		},
	}

	config := &PrometheusConfig{
		Cache: struct {
			Enabled bool          `yaml:"enabled" json:"enabled"`
			TTL     time.Duration `yaml:"ttl" json:"ttl"`
			MaxSize int           `yaml:"max_size" json:"max_size"`
		}{
			Enabled: true,
			TTL:     1 * time.Minute,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	numGoroutines := 50
	queriesPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				query := MetricQuery{Metric: fmt.Sprintf("metric_%d", j%5)}
				_, err := provider.Query(context.Background(), query)
				assert.NoError(t, err)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(5), atomic.LoadInt32(&callCount), "API should only be called for the 5 unique queries")
}

func TestPrometheusProvider_Metrics(t *testing.T) {
	// Since tests can run in parallel, we check increments rather than absolute values.
	// Resetting is not feasible for counters.
	providerCircuitBreakerState.Set(0) // Reset gauge for this test

	initialSuccess := testutil.ToFloat64(providerQueriesTotal.WithLabelValues("success"))
	initialCacheHit := testutil.ToFloat64(providerQueriesTotal.WithLabelValues("cache_hit"))
	initialCacheMiss := testutil.ToFloat64(providerCacheMissesTotal)
	initialError := testutil.ToFloat64(providerQueriesTotal.WithLabelValues("error"))
	initialRejected := testutil.ToFloat64(providerQueriesTotal.WithLabelValues("circuit_breaker_rejected"))

	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			if strings.Contains(query, "error_metric") {
				return nil, nil, errors.New("query failed")
			}
			return &model.Matrix{}, nil, nil
		},
	}

	config := &PrometheusConfig{
		Cache: struct {
			Enabled bool          `yaml:"enabled" json:"enabled"`
			TTL     time.Duration `yaml:"ttl" json:"ttl"`
			MaxSize int           `yaml:"max_size" json:"max_size"`
		}{Enabled: true, TTL: 1 * time.Minute},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 1, // Open after 1 failure
			RecoveryTimeout:  10 * time.Second,
		},
	}

	provider, err := newTestProvider(config, mockAPI)
	assert.NoError(t, err)

	// 1. Successful query (cache miss)
	provider.Query(context.Background(), MetricQuery{Metric: "success_metric"})
	assert.Equal(t, initialSuccess+1, testutil.ToFloat64(providerQueriesTotal.WithLabelValues("success")))
	assert.Equal(t, initialCacheMiss+1, testutil.ToFloat64(providerCacheMissesTotal))

	// 2. Successful query (cache hit)
	provider.Query(context.Background(), MetricQuery{Metric: "success_metric"})
	assert.Equal(t, initialCacheHit+1, testutil.ToFloat64(providerQueriesTotal.WithLabelValues("cache_hit")))

	// 3. Error query, which should open the circuit breaker
	provider.Query(context.Background(), MetricQuery{Metric: "error_metric"})
	assert.Equal(t, initialError+1, testutil.ToFloat64(providerQueriesTotal.WithLabelValues("error")))
	assert.Equal(t, float64(1), testutil.ToFloat64(providerCircuitBreakerState), "Circuit breaker should be open")

	// 4. Circuit breaker rejected query
	provider.Query(context.Background(), MetricQuery{Metric: "another_metric"})
	assert.Equal(t, initialRejected+1, testutil.ToFloat64(providerQueriesTotal.WithLabelValues("circuit_breaker_rejected")))
}

func TestPrometheusProvider_BatchQuery_MergesQueries(t *testing.T) {
	var executedPromQLs []string
	mockAPI := &mockPrometheusAPI{
		queryRangeFunc: func(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
			executedPromQLs = append(executedPromQLs, query)

			// Return a result that can be de-multiplexed
			matrix := model.Matrix{}
			if strings.Contains(query, "metric1") {
				matrix = append(matrix, &model.SampleStream{
					Metric: model.Metric{"__name__": "metric1", "label1": "value1"},
					Values: []model.SamplePair{{Timestamp: 0, Value: 1}},
				})
			}
			if strings.Contains(query, "metric2") {
				matrix = append(matrix, &model.SampleStream{
					Metric: model.Metric{"__name__": "metric2", "label1": "value1"},
					Values: []model.SamplePair{{Timestamp: 0, Value: 2}},
				})
			}
			if strings.Contains(query, "metric3") {
				matrix = append(matrix, &model.SampleStream{
					Metric: model.Metric{"__name__": "metric3", "label2": "value2"},
					Values: []model.SamplePair{{Timestamp: 0, Value: 3}},
				})
			}
			return matrix, nil, nil
		},
	}

	provider, err := newTestProvider(&PrometheusConfig{}, mockAPI)
	assert.NoError(t, err)

	now := time.Now()
	queries := []MetricQuery{
		// Group 1 (mergeable)
		{Metric: "metric1", Labels: map[string]string{"label1": "value1"}, TimeRange: TimeRange{Start: now, End: now.Add(time.Hour)}, Step: time.Minute},
		{Metric: "metric2", Labels: map[string]string{"label1": "value1"}, TimeRange: TimeRange{Start: now, End: now.Add(time.Hour)}, Step: time.Minute},
		// Group 2 (not mergeable with group 1)
		{Metric: "metric3", Labels: map[string]string{"label2": "value2"}, TimeRange: TimeRange{Start: now, End: now.Add(time.Hour)}, Step: time.Minute},
	}

	results, err := provider.BatchQuery(context.Background(), queries)
	assert.NoError(t, err)

	// Assertions
	assert.Len(t, results, 3, "Should have 3 results")
	assert.Len(t, executedPromQLs, 2, "Should have executed 2 queries (one merged, one single)")

	// Check that the correct queries were executed (order not guaranteed)
	expectedMergedQL1 := `{__name__=~"metric1|metric2",label1="value1"}`
	expectedMergedQL2 := `{__name__=~"metric2|metric1",label1="value1"}`
	expectedSingleQL := `metric3{label2="value2"}`

	foundMerged := false
	foundSingle := false
	for _, ql := range executedPromQLs {
		if ql == expectedMergedQL1 || ql == expectedMergedQL2 {
			foundMerged = true
		}
		if ql == expectedSingleQL {
			foundSingle = true
		}
	}
	assert.True(t, foundMerged, "A merged query should have been executed")
	assert.True(t, foundSingle, "A single query should have been executed")

	// Check results are correctly de-multiplexed
	assert.NotNil(t, results[0])
	assert.Equal(t, "metric1", results[0].Query.Metric)
	assert.Len(t, results[0].Series, 1)
	assert.Equal(t, float64(1), results[0].Series[0].Values[0].Value)

	assert.NotNil(t, results[1])
	assert.Equal(t, "metric2", results[1].Query.Metric)
	assert.Len(t, results[1].Series, 1)
	assert.Equal(t, float64(2), results[1].Series[0].Values[0].Value)

	assert.NotNil(t, results[2])
	assert.Equal(t, "metric3", results[2].Query.Metric)
	assert.Len(t, results[2].Series, 1)
	assert.Equal(t, float64(3), results[2].Series[0].Values[0].Value)
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
