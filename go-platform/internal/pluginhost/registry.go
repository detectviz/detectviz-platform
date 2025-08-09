package pluginhost

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	contractspb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"go.uber.org/zap"
)

// Handler 為每個 tool_id 的處理器介面
type Handler interface {
	Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error)
}

// 可選：具備釋放資源能力的處理器
// 若插件內部持有連線池/檔案/背景 goroutine，請實作 Close() 以便 Registry 在卸載或替換時釋放資源。
type ClosableHandler interface {
	Handler
	Close() error
}

// Registry 維護 tool_id → Handler 的映射（並發安全）
type Registry struct {
	rwmu     sync.RWMutex
	byToolID map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{byToolID: map[string]Handler{}}
}

// ErrHandlerExists 代表在嚴格模式下該 toolID 已存在
var ErrHandlerExists = errors.New("handler already registered")

// RegisterStrict 嚴格註冊：若 toolID 已存在則直接回報錯誤，不覆蓋、也不關閉舊 handler。
// 適用於不允許同名覆蓋的場景，可避免無意間的替換造成語義混亂。
func (r *Registry) RegisterStrict(toolID string, h Handler) error {
	r.rwmu.Lock()
	if _, ok := r.byToolID[toolID]; ok {
		r.rwmu.Unlock()
		return fmt.Errorf("%w: tool_id=%s", ErrHandlerExists, toolID)
	}
	r.byToolID[toolID] = h
	r.rwmu.Unlock()
	zap.L().Info("Handler registered (strict)", zap.String("tool_id", toolID))
	return nil
}

// Register 登錄或覆蓋指定 toolID 的處理器（寫鎖）
// - 若同一 toolID 已存在，會嘗試關閉舊的 handler 以避免資源洩漏，然後再覆蓋。
func (r *Registry) Register(toolID string, h Handler) {
	r.rwmu.Lock()
	if old, ok := r.byToolID[toolID]; ok && old != h {
		if err := closeIfPossible(old); err != nil {
			zap.L().Warn("Failed to close previous handler on re-register", zap.String("tool_id", toolID), zap.Error(err))
		}
	}
	r.byToolID[toolID] = h
	r.rwmu.Unlock()
	zap.L().Info("Handler registered", zap.String("tool_id", toolID))
}

// RegisterOrReplace 與 Register 相同語意；提供更語義化的名稱以利閱讀。
func (r *Registry) RegisterOrReplace(toolID string, h Handler) {
	r.Register(toolID, h)
}

// Lookup 讀取指定 toolID 的處理器（讀鎖）
func (r *Registry) Lookup(toolID string) (Handler, bool) {
	r.rwmu.RLock()
	h, ok := r.byToolID[toolID]
	r.rwmu.RUnlock()
	return h, ok
}

// Unregister 解除登錄，並嘗試釋放資源（寫鎖）
// 回傳是否存在且已移除。
func (r *Registry) Unregister(toolID string) bool {
	r.rwmu.Lock()
	old, ok := r.byToolID[toolID]
	if ok {
		delete(r.byToolID, toolID)
	}
	r.rwmu.Unlock()

	if ok {
		if err := closeIfPossible(old); err != nil {
			zap.L().Warn("Failed to close handler on unregister", zap.String("tool_id", toolID), zap.Error(err))
		} else {
			zap.L().Info("Handler unregistered", zap.String("tool_id", toolID))
		}
	}
	return ok
}

// List 取得目前已註冊的所有 toolID 快照（讀鎖）
func (r *Registry) List() []string {
	r.rwmu.RLock()
	ids := make([]string, 0, len(r.byToolID))
	for id := range r.byToolID {
		ids = append(ids, id)
	}
	r.rwmu.RUnlock()
	return ids
}

// Size 目前註冊數量（讀鎖）
func (r *Registry) Size() int {
	r.rwmu.RLock()
	n := len(r.byToolID)
	r.rwmu.RUnlock()
	return n
}

// Shutdown 關閉所有可關閉的 handler 並清空註冊表（寫鎖 + 批次釋放）
func (r *Registry) Shutdown() error {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	var errs error
	for id, h := range r.byToolID {
		if err := closeIfPossible(h); err != nil {
			errs = errors.Join(errs, fmt.Errorf("tool_id=%s: %w", id, err))
		}
		delete(r.byToolID, id)
	}
	if errs != nil {
		zap.L().Warn("Registry shutdown with errors", zap.Error(errs))
	} else {
		zap.L().Info("Registry shutdown complete")
	}
	return errs
}

// closeIfPossible 嘗試呼叫 Close()；支援 ClosableHandler 或 io.Closer
func closeIfPossible(h Handler) error {
	switch x := h.(type) {
	case ClosableHandler:
		return x.Close()
	case io.Closer:
		return x.Close()
	default:
		return nil
	}
}
