package pluginhost

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	pb "github.com/detectviz/detectviz-platform/contracts/gen/go/detectviz/contracts/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// MockHandler 模擬處理器用於測試
type MockHandler struct {
	invokeDuration time.Duration
	shouldError    bool
	callCount      int64
	mu             sync.Mutex
}

func (m *MockHandler) Invoke(ctx context.Context, req *pb.ToolInvokeRequest) (*pb.ToolInvokeReply, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	
	if m.invokeDuration > 0 {
		time.Sleep(m.invokeDuration)
	}
	
	if m.shouldError {
		return &pb.ToolInvokeReply{
			Status: &statuspb.Status{Code: 2, Message: "模擬錯誤"},
		}, nil
	}
	
	result, _ := structpb.NewStruct(map[string]interface{}{
		"success": true,
		"data":    "test response",
	})
	
	return &pb.ToolInvokeReply{
		Result: result,
		Status: &statuspb.Status{Code: 0, Message: "OK"},
	}, nil
}

func (m *MockHandler) Close() error {
	return nil
}

func (m *MockHandler) GetCallCount() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

// TestResourceMonitor_BasicOperations 測試基本的監控操作
func TestResourceMonitor_BasicOperations(t *testing.T) {
	monitor, err := NewResourceMonitor(100 * time.Millisecond)
	require.NoError(t, err)
	defer monitor.Stop()
	
	pluginID := "test_plugin"
	
	// 測試註冊插件
	monitor.RegisterPlugin(pluginID)
	
	// 驗證插件已註冊
	metrics := monitor.GetPluginMetrics(pluginID)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.ActiveRequests)
	
	// 測試記錄請求
	monitor.RecordRequest(pluginID)
	metrics = monitor.GetPluginMetrics(pluginID)
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.ActiveRequests)
	
	// 測試記錄請求完成
	monitor.RecordRequestComplete(pluginID, 100, false)
	metrics = monitor.GetPluginMetrics(pluginID)
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.ActiveRequests)
	assert.Equal(t, float64(100), metrics.AvgResponseTimeMs)
	assert.Equal(t, float64(0), metrics.ErrorRate)
	
	// 測試錯誤請求
	monitor.RecordRequest(pluginID)
	monitor.RecordRequestComplete(pluginID, 200, true)
	metrics = monitor.GetPluginMetrics(pluginID)
	assert.Equal(t, int64(2), metrics.TotalRequests)
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, float64(50), metrics.ErrorRate) // 1/2 = 50%
	
	// 測試資源使用更新
	monitor.UpdateResourceUsage(pluginID, 1024*1024, 10, 5)
	metrics = monitor.GetPluginMetrics(pluginID)
	assert.Equal(t, int64(1024*1024), metrics.MemoryUsageBytes)
	assert.Equal(t, int32(10), metrics.GoroutineCount)
	assert.Equal(t, int32(5), metrics.ConnectionCount)
	
	// 測試取消註冊
	monitor.UnregisterPlugin(pluginID)
	metrics = monitor.GetPluginMetrics(pluginID)
	assert.Nil(t, metrics)
}

// TestResourceMonitor_ConcurrentAccess 測試併發存取安全性
func TestResourceMonitor_ConcurrentAccess(t *testing.T) {
	monitor, err := NewResourceMonitor(50 * time.Millisecond)
	require.NoError(t, err)
	defer monitor.Stop()
	
	pluginCount := 10
	requestCount := 100
	
	// 註冊多個插件
	for i := 0; i < pluginCount; i++ {
		monitor.RegisterPlugin(fmt.Sprintf("plugin_%d", i))
	}
	
	var wg sync.WaitGroup
	
	// 併發記錄請求
	for i := 0; i < pluginCount; i++ {
		pluginID := fmt.Sprintf("plugin_%d", i)
		
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			
			for j := 0; j < requestCount; j++ {
				monitor.RecordRequest(pid)
				time.Sleep(1 * time.Millisecond)
				monitor.RecordRequestComplete(pid, int64(j), j%10 == 0) // 10% 錯誤率
			}
		}(pluginID)
	}
	
	// 併發更新資源使用
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < requestCount; i++ {
			for j := 0; j < pluginCount; j++ {
				pluginID := fmt.Sprintf("plugin_%d", j)
				monitor.UpdateResourceUsage(pluginID, int64(i*1024), int32(i), int32(i/2))
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()
	
	wg.Wait()
	
	// 驗證結果
	allMetrics := monitor.GetAllMetrics()
	assert.Equal(t, pluginCount, len(allMetrics))
	
	for i := 0; i < pluginCount; i++ {
		pluginID := fmt.Sprintf("plugin_%d", i)
		metrics := allMetrics[pluginID]
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(requestCount), metrics.TotalRequests)
		assert.Equal(t, int64(0), metrics.ActiveRequests) // 所有請求都已完成
		assert.Equal(t, int64(requestCount/10), metrics.TotalErrors) // 10% 錯誤率
		assert.Equal(t, float64(10), metrics.ErrorRate) // 10%
	}
}

// TestMonitoredHandler 測試監控處理器
func TestMonitoredHandler(t *testing.T) {
	monitor, err := NewResourceMonitor(100 * time.Millisecond)
	require.NoError(t, err)
	defer monitor.Stop()
	
	mockHandler := &MockHandler{
		invokeDuration: 50 * time.Millisecond,
		shouldError:    false,
	}
	
	pluginID := "test_monitored_plugin"
	monitoredHandler := NewMonitoredHandler(pluginID, mockHandler, monitor)
	defer monitoredHandler.Close()
	
	// 創建測試請求
	req := &pb.ToolInvokeRequest{
		ToolId: "test_tool",
		Payload: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"test": structpb.NewStringValue("data"),
			},
		},
	}
	
	ctx := context.Background()
	
	// 執行請求
	reply, err := monitoredHandler.Invoke(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, int32(0), reply.Status.Code)
	
	// 驗證監控指標
	metrics := monitor.GetPluginMetrics(pluginID)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.ActiveRequests)
	assert.Equal(t, int64(0), metrics.TotalErrors)
	assert.True(t, metrics.AvgResponseTimeMs >= 50) // 至少 50ms
	
	// 驗證處理器被調用
	assert.Equal(t, int64(1), mockHandler.GetCallCount())
}

// TestMonitoredHandler_ErrorHandling 測試監控處理器的錯誤處理
func TestMonitoredHandler_ErrorHandling(t *testing.T) {
	monitor, err := NewResourceMonitor(100 * time.Millisecond)
	require.NoError(t, err)
	defer monitor.Stop()
	
	mockHandler := &MockHandler{
		invokeDuration: 25 * time.Millisecond,
		shouldError:    true, // 強制錯誤
	}
	
	pluginID := "test_error_plugin"
	monitoredHandler := NewMonitoredHandler(pluginID, mockHandler, monitor)
	defer monitoredHandler.Close()
	
	req := &pb.ToolInvokeRequest{
		ToolId: "test_tool",
		Payload: &structpb.Struct{},
	}
	
	ctx := context.Background()
	
	// 執行請求（預期錯誤）
	reply, err := monitoredHandler.Invoke(ctx, req)
	assert.NoError(t, err) // gRPC 層面沒有錯誤
	assert.NotNil(t, reply)
	assert.Equal(t, int32(2), reply.Status.Code) // 業務層錯誤
	
	// 驗證錯誤被正確記錄
	metrics := monitor.GetPluginMetrics(pluginID)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.TotalRequests)
	assert.Equal(t, int64(0), metrics.ActiveRequests)
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, float64(100), metrics.ErrorRate) // 100% 錯誤率
}

// TestResourceMonitor_HealthStatus 測試健康狀態監控
func TestResourceMonitor_HealthStatus(t *testing.T) {
	monitor, err := NewResourceMonitor(100 * time.Millisecond)
	require.NoError(t, err)
	defer monitor.Stop()
	
	// 註冊多個插件
	pluginIDs := []string{"plugin_1", "plugin_2", "plugin_3"}
	for _, pid := range pluginIDs {
		monitor.RegisterPlugin(pid)
		
		// 模擬一些請求活動
		monitor.RecordRequest(pid)
		monitor.RecordRequestComplete(pid, 100, false)
		monitor.RecordRequest(pid)
		monitor.RecordRequestComplete(pid, 200, true) // 一個錯誤
	}
	
	// 獲取健康狀態
	healthStatus := monitor.GetHealthStatus()
	assert.NotNil(t, healthStatus)
	
	// 驗證健康狀態數據
	assert.Equal(t, len(pluginIDs), healthStatus["total_plugins"])
	assert.Equal(t, int64(0), healthStatus["total_active_requests"]) // 所有請求已完成
	assert.Equal(t, int64(6), healthStatus["total_requests"])        // 每個插件2個請求，共6個
	assert.Equal(t, int64(3), healthStatus["total_errors"])          // 每個插件1個錯誤，共3個
	assert.Equal(t, float64(50), healthStatus["overall_error_rate"]) // 3/6 = 50%
	assert.Contains(t, healthStatus, "monitoring_interval")
	assert.Contains(t, healthStatus, "last_collection_time")
}

// BenchmarkResourceMonitor_RecordOperations 基準測試監控記錄操作
func BenchmarkResourceMonitor_RecordOperations(b *testing.B) {
	monitor, err := NewResourceMonitor(time.Second)
	require.NoError(b, err)
	defer monitor.Stop()
	
	pluginID := "benchmark_plugin"
	monitor.RegisterPlugin(pluginID)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			monitor.RecordRequest(pluginID)
			monitor.RecordRequestComplete(pluginID, 100, false)
		}
	})
}

// BenchmarkResourceMonitor_ConcurrentAccess 基準測試併發存取
func BenchmarkResourceMonitor_ConcurrentAccess(b *testing.B) {
	monitor, err := NewResourceMonitor(time.Second)
	require.NoError(b, err)
	defer monitor.Stop()
	
	// 註冊多個插件
	pluginCount := 10
	for i := 0; i < pluginCount; i++ {
		monitor.RegisterPlugin(fmt.Sprintf("plugin_%d", i))
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		pluginID := fmt.Sprintf("plugin_%d", time.Now().UnixNano()%int64(pluginCount))
		for pb.Next() {
			monitor.RecordRequest(pluginID)
			monitor.UpdateResourceUsage(pluginID, 1024, 10, 5)
			monitor.RecordRequestComplete(pluginID, 100, false)
		}
	})
}