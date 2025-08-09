package pluginhost

import (
	"context"

	contractspb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
)

// Handler 為每個 tool_id 的處理器介面
type Handler interface {
	Invoke(ctx context.Context, req *contractspb.ToolInvokeRequest) (*contractspb.ToolInvokeReply, error)
}

// Registry 維護 tool_id → Handler 的映射
type Registry struct {
	byToolID map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{byToolID: map[string]Handler{}}
}

func (r *Registry) Register(toolID string, h Handler) {
	r.byToolID[toolID] = h
}

func (r *Registry) Lookup(toolID string) (Handler, bool) {
	h, ok := r.byToolID[toolID]
	return h, ok
}
