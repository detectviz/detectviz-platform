package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	v1 "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/stretchr/testify/assert"
)

// Mock handler implementations for testing
type mockHandler struct {
	id       string
	invoked  bool
	closed   bool
	closeErr error
}

func (m *mockHandler) Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error) {
	m.invoked = true
	return &v1.InvokeResponse{}, nil
}

func (m *mockHandler) Close() error {
	m.closed = true
	return m.closeErr
}

// Mock handler without close method (升級為 ClosableHandler)
type basicMockHandler struct {
	id      string
	invoked bool
	closed  bool
}

func (m *basicMockHandler) Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error) {
	m.invoked = true
	return &v1.InvokeResponse{}, nil
}

func (m *basicMockHandler) Close() error {
	m.closed = true
	return nil
}

func TestRegistry_NewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.GetPluginCount())
	assert.Empty(t, r.GetPluginNames())
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}

	// First registration should succeed
	err := r.Register("tool1", handler1)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.GetPluginCount())

	// Duplicate registration should fail
	err = r.Register("tool1", handler2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "插件已註冊")
	assert.Equal(t, 1, r.GetPluginCount())

	// Verify original handler is still registered
	h, ok := r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler1, h)
}

func TestRegistry_RegisterOrReplace(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}

	// Initial registration
	err := r.RegisterOrReplace("tool1", handler1)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.GetPluginCount())

	// Replacement should close old handler and register new one
	err = r.RegisterOrReplace("tool1", handler2)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.GetPluginCount())
	assert.True(t, handler1.closed, "Old handler should be closed")

	// Verify new handler is registered
	h, ok := r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler2, h)
}

func TestRegistry_RegisterOrReplace_CloseError(t *testing.T) {
	r := NewRegistry()
	closeError := errors.New("close failed")
	handler1 := &mockHandler{id: "handler1", closeErr: closeError}
	handler2 := &mockHandler{id: "handler2"}

	err := r.RegisterOrReplace("tool1", handler1)
	assert.NoError(t, err)
	// This should log a warning but still succeed
	err = r.RegisterOrReplace("tool1", handler2)
	assert.NoError(t, err)

	assert.True(t, handler1.closed)
	h, ok := r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler2, h)
}

func TestRegistry_RegisterOrReplace_BasicHandler(t *testing.T) {
	r := NewRegistry()
	handler1 := &basicMockHandler{id: "handler1"} // Now implements Close method
	handler2 := &mockHandler{id: "handler2"}

	err := r.RegisterOrReplace("tool1", handler1)
	assert.NoError(t, err)
	// This should work
	err = r.RegisterOrReplace("tool1", handler2)
	assert.NoError(t, err)
	assert.True(t, handler1.closed)

	h, ok := r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler2, h)
}

func TestRegistry_Lookup(t *testing.T) {
	r := NewRegistry()
	handler := &mockHandler{id: "handler"}

	// Lookup non-existent handler
	h, ok := r.Lookup("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, h)

	// Register and lookup
	err := r.RegisterOrReplace("tool1", handler)
	assert.NoError(t, err)
	h, ok = r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler, h)
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	handler := &mockHandler{id: "handler"}

	// Unregister non-existent handler
	err := r.Unregister("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "插件不存在")

	// Register, then unregister
	err = r.RegisterOrReplace("tool1", handler)
	assert.NoError(t, err)
	err = r.Unregister("tool1")
	assert.NoError(t, err)
	assert.True(t, handler.closed)
	assert.Equal(t, 0, r.GetPluginCount())

	// Verify handler is gone
	h, ok := r.Lookup("tool1")
	assert.False(t, ok)
	assert.Nil(t, h)
}

func TestRegistry_GetPluginNames(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}

	// Empty registry
	names := r.GetPluginNames()
	assert.Empty(t, names)

	// Add handlers
	err := r.RegisterOrReplace("tool1", handler1)
	assert.NoError(t, err)
	err = r.RegisterOrReplace("tool2", handler2)
	assert.NoError(t, err)

	names = r.GetPluginNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "tool1")
	assert.Contains(t, names, "tool2")
}

func TestRegistry_Shutdown(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}
	handler3 := &mockHandler{id: "handler3", closeErr: errors.New("close error")}

	err := r.RegisterOrReplace("tool1", handler1)
	assert.NoError(t, err)
	err = r.RegisterOrReplace("tool2", handler2)
	assert.NoError(t, err)
	err = r.RegisterOrReplace("tool3", handler3)
	assert.NoError(t, err)

	r.Shutdown() // Shutdown doesn't return error in new implementation

	// All handlers should be closed
	assert.True(t, handler1.closed)
	assert.True(t, handler2.closed)
	assert.True(t, handler3.closed)
	assert.Equal(t, 0, r.GetPluginCount())
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent register/lookup operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				toolID := fmt.Sprintf("tool_%d_%d", id, j)
				handler := &mockHandler{id: fmt.Sprintf("handler_%d_%d", id, j)}

				// 註冊處理器
				err := r.RegisterOrReplace(toolID, handler)
				assert.NoError(t, err)

				// 立即查詢應該成功
				h, ok := r.Lookup(toolID)
				assert.True(t, ok, fmt.Sprintf("Should find handler for %s", toolID))
				assert.NotNil(t, h, fmt.Sprintf("Handler %s should not be nil", toolID))

				// 偶爾取消註冊（但不是剛註冊的）
				if j > 0 && j%10 == 0 {
					oldToolID := fmt.Sprintf("tool_%d_%d", id, j-5) // 取消註冊較舊的
					_ = r.Unregister(oldToolID)                     // 忽略錯誤，因為可能已經被取消註冊
				}
			}
		}(i)
	}

	wg.Wait()
	// Should not crash or deadlock
}

func TestRegistry_GracefulShutdownIntegration(t *testing.T) {
	// Simulate a realistic shutdown scenario
	r := NewRegistry()

	// Register multiple handlers with different characteristics
	handlers := make([]*mockHandler, 5)
	for i := 0; i < 5; i++ {
		handlers[i] = &mockHandler{id: "handler" + string(rune('1'+i))}
		err := r.RegisterOrReplace("tool"+string(rune('1'+i)), handlers[i])
		assert.NoError(t, err)
	}

	assert.Equal(t, 5, r.GetPluginCount())

	// Simulate graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		r.Shutdown()
		done <- true
	}()

	select {
	case <-done:
		assert.Equal(t, 0, r.GetPluginCount())

		// Verify all handlers were closed
		for _, h := range handlers {
			assert.True(t, h.closed, "Handler %s should be closed", h.id)
		}

	case <-ctx.Done():
		t.Fatal("Shutdown took too long")
	}
}

// Benchmark tests
func BenchmarkRegistry_Register(b *testing.B) {
	r := NewRegistry()
	handlers := make([]*basicMockHandler, b.N)
	for i := 0; i < b.N; i++ {
		handlers[i] = &basicMockHandler{id: "handler"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.RegisterOrReplace("tool", handlers[i])
	}
}

func BenchmarkRegistry_Lookup(b *testing.B) {
	r := NewRegistry()
	handler := &basicMockHandler{id: "handler"}
	_ = r.RegisterOrReplace("tool", handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Lookup("tool")
	}
}

func BenchmarkRegistry_ConcurrentLookup(b *testing.B) {
	r := NewRegistry()
	handler := &basicMockHandler{id: "handler"}
	_ = r.RegisterOrReplace("tool", handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.Lookup("tool")
		}
	})
}
