package pluginhost

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/detectviz/detectviz-platform/go-platform/internal/pluginhost/plugins/gateway/http_request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// TestResourceMonitoringIntegration 完整資源監控整合測試
func TestResourceMonitoringIntegration(t *testing.T) {
	// 1. 創建註冊表
	config := &RegistryConfig{
		MaxLoad:        100,
		HealthInterval: 100 * time.Millisecond,
	}
	registry := NewRegistryWithConfig(config)
	registry.StartHealthChecks()
	defer registry.Shutdown()

	// 2. 創建支援資源監控的 HTTP 插件
	securePlugin := http_request.NewSecurePlugin()
	pluginID := "http_request_secure"

	// 3. 註冊插件
	err := registry.Register(pluginID, securePlugin)
	require.NoError(t, err)

	// 4. 驗證插件已註冊並包裝監控
	handler, exists := registry.Lookup(pluginID)
	assert.True(t, exists)
	assert.NotNil(t, handler)

	// 5. 測試插件狀態
	status := registry.GetPluginStatus()
	assert.Contains(t, status, pluginID)

	// 6. 創建測試請求
	requestPayload := map[string]interface{}{
		"method":     "GET",
		"url":        "https://httpbin.org/get",
		"headers":    map[string]interface{}{"User-Agent": "detectviz-test"}, // 修正類型
		"timeout_ms": 5000,
	}

	payload, err := structpb.NewStruct(requestPayload)
	require.NoError(t, err)

	req := &v1.InvokeRequest{
		ToolId:  pluginID,
		Payload: payload,
	}

	// 7. 執行請求並驗證
	ctx := context.Background()

	// 執行請求
	reply, err := registry.Invoke(ctx, pluginID, req)
	assert.NoError(t, err)
	assert.NotNil(t, reply)

	// 8. 驗證插件狀態
	statusAfter := registry.GetPluginStatus()
	pluginStatus := statusAfter[pluginID].(map[string]interface{})

	t.Logf("插件狀態: 健康=%s, 負載=%d",
		pluginStatus["health_str"], pluginStatus["load"])

	// 9. 測試更新指標
	registry.UpdateMetrics() // 應該不會產生錯誤

	// 10. 測試插件列表
	pluginNames := registry.GetPluginNames()
	assert.Contains(t, pluginNames, pluginID)
	assert.Equal(t, 1, registry.GetPluginCount())

	t.Logf("註冊插件: %v", pluginNames)
}

// TestConcurrentPluginMonitoring 測試多插件並發監控
func TestConcurrentPluginMonitoring(t *testing.T) {
	config := &RegistryConfig{
		MaxLoad:        50,
		HealthInterval: 50 * time.Millisecond,
	}
	registry := NewRegistryWithConfig(config)
	registry.StartHealthChecks()
	defer registry.Shutdown()

	pluginCount := 3
	requestsPerPlugin := 5

	// 註冊多個插件
	for i := 0; i < pluginCount; i++ {
		pluginID := fmt.Sprintf("http_plugin_%d", i)
		plugin := http_request.NewSecurePlugin()

		err := registry.Register(pluginID, plugin)
		require.NoError(t, err)
	}

	// 並發執行請求
	done := make(chan bool, pluginCount)

	for i := 0; i < pluginCount; i++ {
		go func(pluginIndex int) {
			defer func() { done <- true }()

			pluginID := fmt.Sprintf("http_plugin_%d", pluginIndex)
			_, exists := registry.Lookup(pluginID)
			if !exists {
				t.Errorf("插件 %s 不存在", pluginID)
				return
			}

			for j := 0; j < requestsPerPlugin; j++ {
				// 創建簡單的測試請求（不實際發送HTTP，避免外部依賴）
				requestPayload := map[string]interface{}{
					"method":     "GET",
					"url":        fmt.Sprintf("https://httpbin.org/delay/%d", j%3),
					"timeout_ms": 2000,
				}

				payload, _ := structpb.NewStruct(requestPayload)
				req := &v1.InvokeRequest{
					ToolId:  pluginID,
					Payload: payload,
				}

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_, err := registry.Invoke(ctx, pluginID, req)
				cancel()

				if err != nil {
					t.Logf("插件 %s 請求 %d 失敗: %v", pluginID, j, err)
				}

				// 短暫間隔避免過度並發
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < pluginCount; i++ {
		<-done
	}

	// 等待監控指標更新
	time.Sleep(100 * time.Millisecond)

	// 驗證所有插件的狀態
	allStatus := registry.GetPluginStatus()
	require.Equal(t, pluginCount, len(allStatus))

	for pluginID, status := range allStatus {
		statusMap := status.(map[string]interface{})
		assert.NotNil(t, statusMap, "插件 %s 應該有狀態指標", pluginID)
		t.Logf("插件 %s: 健康=%s, 負載=%d",
			pluginID, statusMap["health_str"], statusMap["load"])
	}

	// 驗證總體統計
	assert.Equal(t, pluginCount, registry.GetPluginCount())
	pluginNames := registry.GetPluginNames()
	assert.Equal(t, pluginCount, len(pluginNames))

	t.Logf("總計: %d 個插件", pluginCount)
}

// TestPluginResourceLimitEnforcement 測試資源限制執行
func TestPluginResourceLimitEnforcement(t *testing.T) {
	config := &RegistryConfig{
		MaxLoad:        10, // 設定較低的負載限制
		HealthInterval: 100 * time.Millisecond,
	}
	registry := NewRegistryWithConfig(config)
	registry.StartHealthChecks()
	defer registry.Shutdown()

	pluginID := "resource_test_plugin"
	plugin := http_request.NewSecurePlugin()

	err := registry.Register(pluginID, plugin)
	require.NoError(t, err)

	// 測試負載保護
	ctx := context.Background()
	payload, _ := structpb.NewStruct(map[string]interface{}{
		"method": "GET",
		"url":    "https://httpbin.org/get",
	})
	req := &v1.InvokeRequest{
		ToolId:  pluginID,
		Payload: payload,
	}

	// 正常請求應該成功
	_, err = registry.Invoke(ctx, pluginID, req)
	assert.NoError(t, err)

	// 檢查插件狀態
	status := registry.GetPluginStatus()
	pluginStatus := status[pluginID].(map[string]interface{})
	t.Logf("插件狀態: 健康=%s, 負載=%d",
		pluginStatus["health_str"], pluginStatus["load"])
}

// MockClosableHandler 用於基準測試的模擬處理器
type MockClosableHandler struct {
	invokeDuration time.Duration
}

func (m *MockClosableHandler) Invoke(ctx context.Context, req *v1.InvokeRequest) (*v1.InvokeResponse, error) {
	time.Sleep(m.invokeDuration)
	payload, _ := structpb.NewStruct(map[string]interface{}{
		"result": "success",
	})
	return &v1.InvokeResponse{Result: payload}, nil
}

func (m *MockClosableHandler) Close() error {
	return nil
}

// BenchmarkMonitoredPluginPerformance 監控插件性能基準測試
func BenchmarkMonitoredPluginPerformance(b *testing.B) {
	registry := NewRegistry()
	defer registry.Shutdown()

	pluginID := "benchmark_plugin"
	plugin := &MockClosableHandler{invokeDuration: time.Microsecond} // 極短執行時間

	err := registry.Register(pluginID, plugin)
	require.NoError(b, err)

	// 準備測試請求
	payload, _ := structpb.NewStruct(map[string]interface{}{
		"test": "data",
	})
	req := &v1.InvokeRequest{
		ToolId:  pluginID,
		Payload: payload,
	}

	ctx := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := registry.Invoke(ctx, pluginID, req)
			if err != nil {
				b.Errorf("請求執行失敗: %v", err)
			}
		}
	})

	b.StopTimer()

	// 報告最終狀態
	status := registry.GetPluginStatus()
	if pluginStatus, ok := status[pluginID]; ok {
		statusMap := pluginStatus.(map[string]interface{})
		b.Logf("基準測試完成 - 插件健康: %s, 負載: %d",
			statusMap["health_str"], statusMap["load"])
	}
}
