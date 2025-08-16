package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	contractspb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/stretchr/testify/assert"
)

// Mock handler implementations for testing
type mockHandler struct {
	id       string
	invoked  bool
	closed   bool
	closeErr error
}

func (m *mockHandler) Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error) {
	m.invoked = true
	return &contractspb.ToolInvokeReply{}, nil
}

func (m *mockHandler) Close() error {
	m.closed = true
	return m.closeErr
}

// Mock handler without close method
type basicMockHandler struct {
	id      string
	invoked bool
}

func (m *basicMockHandler) Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error) {
	m.invoked = true
	return &contractspb.ToolInvokeReply{}, nil
}

func TestRegistry_NewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Size())
	assert.Empty(t, r.List())
}

func TestRegistry_RegisterStrict(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}

	// First registration should succeed
	err := r.RegisterStrict("tool1", handler1)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.Size())

	// Duplicate registration should fail
	err = r.RegisterStrict("tool1", handler2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrHandlerExists)
	assert.Equal(t, 1, r.Size())

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
	r.RegisterOrReplace("tool1", handler1)
	assert.Equal(t, 1, r.Size())

	// Replacement should close old handler and register new one
	r.RegisterOrReplace("tool1", handler2)
	assert.Equal(t, 1, r.Size())
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

	r.RegisterOrReplace("tool1", handler1)
	// This should log a warning but still succeed
	r.RegisterOrReplace("tool1", handler2)

	assert.True(t, handler1.closed)
	h, ok := r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler2, h)
}

func TestRegistry_RegisterOrReplace_BasicHandler(t *testing.T) {
	r := NewRegistry()
	handler1 := &basicMockHandler{id: "handler1"} // No Close method
	handler2 := &mockHandler{id: "handler2"}

	r.RegisterOrReplace("tool1", handler1)
	// This should work even though handler1 doesn't implement Close
	r.RegisterOrReplace("tool1", handler2)

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
	r.RegisterOrReplace("tool1", handler)
	h, ok = r.Lookup("tool1")
	assert.True(t, ok)
	assert.Same(t, handler, h)
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	handler := &mockHandler{id: "handler"}

	// Unregister non-existent handler
	removed := r.Unregister("nonexistent")
	assert.False(t, removed)

	// Register, then unregister
	r.RegisterOrReplace("tool1", handler)
	removed = r.Unregister("tool1")
	assert.True(t, removed)
	assert.True(t, handler.closed)
	assert.Equal(t, 0, r.Size())

	// Verify handler is gone
	h, ok := r.Lookup("tool1")
	assert.False(t, ok)
	assert.Nil(t, h)
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}

	// Empty registry
	list := r.List()
	assert.Empty(t, list)

	// Add handlers
	r.RegisterOrReplace("tool1", handler1)
	r.RegisterOrReplace("tool2", handler2)

	list = r.List()
	assert.Len(t, list, 2)
	assert.Contains(t, list, "tool1")
	assert.Contains(t, list, "tool2")
}

func TestRegistry_Shutdown(t *testing.T) {
	r := NewRegistry()
	handler1 := &mockHandler{id: "handler1"}
	handler2 := &mockHandler{id: "handler2"}
	handler3 := &mockHandler{id: "handler3", closeErr: errors.New("close error")}

	r.RegisterOrReplace("tool1", handler1)
	r.RegisterOrReplace("tool2", handler2)
	r.RegisterOrReplace("tool3", handler3)

	err := r.Shutdown()

	// Should return error from handler3, but still close all handlers
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close error")
	assert.Contains(t, err.Error(), "tool3")

	assert.True(t, handler1.closed)
	assert.True(t, handler2.closed)
	assert.True(t, handler3.closed)
	assert.Equal(t, 0, r.Size())
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
				r.RegisterOrReplace(toolID, handler)

				// 立即查詢應該成功
				h, ok := r.Lookup(toolID)
				assert.True(t, ok, fmt.Sprintf("Should find handler for %s", toolID))
				assert.NotNil(t, h, fmt.Sprintf("Handler %s should not be nil", toolID))

				// 偶爾取消註冊（但不是剛註冊的）
				if j > 0 && j%10 == 0 {
					oldToolID := fmt.Sprintf("tool_%d_%d", id, j-5) // 取消註冊較舊的
					r.Unregister(oldToolID)
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
		r.RegisterOrReplace("tool"+string(rune('1'+i)), handlers[i])
	}

	assert.Equal(t, 5, r.Size())

	// Simulate graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.Shutdown()
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
		assert.Equal(t, 0, r.Size())

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
	handlers := make([]Handler, b.N)
	for i := 0; i < b.N; i++ {
		handlers[i] = &basicMockHandler{id: "handler"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RegisterOrReplace("tool", handlers[i])
	}
}

func BenchmarkRegistry_Lookup(b *testing.B) {
	r := NewRegistry()
	handler := &basicMockHandler{id: "handler"}
	r.RegisterOrReplace("tool", handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Lookup("tool")
	}
}

func BenchmarkRegistry_ConcurrentLookup(b *testing.B) {
	r := NewRegistry()
	handler := &basicMockHandler{id: "handler"}
	r.RegisterOrReplace("tool", handler)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.Lookup("tool")
		}
	})
}
