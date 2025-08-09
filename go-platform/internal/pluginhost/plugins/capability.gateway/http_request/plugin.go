package http_request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const (
	ToolID   = "detectviz.tools.http_request"
	PluginID = "detectviz.plugins.http_request_gateway"
)

type Plugin struct {
	client *http.Client
	// TODO: egress allowlist、mTLS transport、retry/backoff
}

func New() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *Plugin) Invoke(ctx context.Context, req *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error) {
	pl := req.GetPayload().AsMap()
	method, _ := pl["method"].(string)
	url, _ := pl["url"].(string)

	var body io.Reader
	if b, ok := pl["body"].(string); ok && b != "" {
		body = bytes.NewBufferString(b)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return wrapErr(err), nil
	}
	if hs, ok := pl["headers"].(map[string]interface{}); ok {
		for k, v := range hs {
			if sv, _ := v.(string); sv != "" {
				httpReq.Header.Set(k, sv)
			}
		}
	}

	start := time.Now()
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return wrapErr(err), nil
	}
	defer resp.Body.Close()
	bs, _ := io.ReadAll(resp.Body)

	out := map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": resp.Header,
		"body":    string(bs),
	}
	s, _ := structpb.NewStruct(out)

	return &pb.ToolInvokeReply{
		Result: s,
		Status: &statuspb.Status{Code: 0, Message: "OK"},
		ExecMeta: &pb.ToolExecutionMeta{
			Attempt:    1,
			DurationMs: uint64(time.Since(start) / time.Millisecond),
			PluginId:   PluginID,
		},
	}, nil
}

func wrapErr(err error) *pb.ToolInvokeReply {
	return &pb.ToolInvokeReply{
		Result: nil,
		Status: &statuspb.Status{Code: 2, Message: err.Error()}, // UNKNOWN
		ExecMeta: &pb.ToolExecutionMeta{
			Attempt:  1,
			PluginId: PluginID,
		},
	}
}
