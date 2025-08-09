package http_request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	// 與 contracts/module.card 規範一致：tools 命名空間
	ToolID   = "detectviz.tools.http_request"
	PluginID = "detectviz.plugins.http_request_gateway"

	defaultTimeout      = 5 * time.Second
	defaultMaxResponseB = 1 << 20 // 1 MiB
)

// Plugin 提供以 HTTP 為主的工具執行能力。
// 支援參數（payload 中）：
// - method: GET/POST/PUT/PATCH/DELETE（預設 GET）
// - url: 目標 URL（必填）
// - headers: map[string]string
// - query: map[string]string（會附加到 URL 上）
// - body: string（原樣寫入）
// - json: object（自動 JSON encode，Content-Type: application/json）
// - form: map[string]string（application/x-www-form-urlencoded）
// - timeout_ms: int（覆蓋預設）
// - max_response_bytes: int（覆蓋預設）
// - allow_insecure_tls: bool（預留，未來支援自訂 Transport）
// 回應：status, headers(map[string]string), body(string)

type Plugin struct {
	client *http.Client
	// TODO: 支援 mTLS/自訂 Transport、egress allowlist、retry/backoff
}

func New() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: defaultTimeout, Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}
}

func (p *Plugin) Close() error {
	// 若使用自訂 Transport，可在此加入釋放邏輯
	if tr, ok := p.client.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
	return nil
}

func (p *Plugin) Invoke(ctx context.Context, req *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error) {
	pl := req.GetPayload().AsMap()

	method := strings.ToUpper(getString(pl, "method", "GET"))
	if method == "" {
		method = "GET"
	}
	urlStr := getString(pl, "url", "")
	if urlStr == "" {
		return wrapErr(errors.New("missing required field: url")), nil
	}

	// 處理 query 參數
	if qmap, ok := getMapStringString(pl, "query"); ok {
		u, err := url.Parse(urlStr)
		if err != nil {
			return wrapErr(err), nil
		}
		q := u.Query()
		for k, v := range qmap {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		urlStr = u.String()
	}

	// 構建 body 與 Content-Type
	var body io.Reader
	contentType := ""
	if obj, ok := pl["json"]; ok && obj != nil {
		bs, err := json.Marshal(obj)
		if err != nil {
			return wrapErr(err), nil
		}
		body = bytes.NewReader(bs)
		contentType = "application/json"
	} else if formMap, ok := getMapStringString(pl, "form"); ok {
		vals := url.Values{}
		for k, v := range formMap {
			vals.Set(k, v)
		}
		body = strings.NewReader(vals.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else if b, ok := pl["body"].(string); ok && b != "" {
		body = strings.NewReader(b)
	}

	// Timeout（可覆蓋）
	ctx, cancel := withOptionalTimeout(ctx, pl)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return wrapErr(err), nil
	}
	// 設定 headers（支援 map[string]string / map[string]interface{}）
	if hmap, ok := getMapStringString(pl, "headers"); ok {
		for k, v := range hmap {
			if v != "" {
				httpReq.Header.Set(k, v)
			}
		}
	}
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	start := time.Now()
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return wrapErr(err), nil
	}
	defer resp.Body.Close()

	// 限制回應大小，避免 OOM
	maxB := int64(getInt(pl, "max_response_bytes", int(defaultMaxResponseB)))
	lr := io.LimitReader(resp.Body, maxB)
	bs, rerr := io.ReadAll(lr)
	if rerr != nil {
		return wrapErr(rerr), nil
	}

	// headers 正規化為 map[string]string（以逗號合併）
	hdr := make(map[string]interface{}, len(resp.Header))
	for k, vs := range resp.Header {
		hdr[k] = strings.Join(vs, ",")
	}

	// 非 2xx 視為錯誤碼，但回傳內容仍帶回
	status := &statuspb.Status{Code: 0, Message: "OK"}
	if resp.StatusCode >= 400 {
		status = &statuspb.Status{Code: 2, Message: resp.Status} // UNKNOWN
	}

	out := map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": hdr,
		"body":    string(bs),
	}
	s, _ := structpb.NewStruct(out)

	return &pb.ToolInvokeReply{
		Result: s,
		Status: status,
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

func withOptionalTimeout(ctx context.Context, pl map[string]interface{}) (context.Context, context.CancelFunc) {
	timeoutMs := getInt(pl, "timeout_ms", int(defaultTimeout/time.Millisecond))
	if timeoutMs <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
}

func getString(m map[string]interface{}, key, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getInt(m map[string]interface{}, key string, def int) int {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case int64:
			return int(t)
		case json.Number:
			if i, err := t.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return def
}

func getMapStringString(m map[string]interface{}, key string) (map[string]string, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return nil, false
	}
	switch mm := v.(type) {
	case map[string]string:
		return mm, true
	case map[string]interface{}:
		out := make(map[string]string, len(mm))
		for k, val := range mm {
			if s, ok := val.(string); ok {
				out[k] = s
			}
		}
		return out, true
	default:
		return nil, false
	}
}
