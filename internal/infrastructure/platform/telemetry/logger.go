package telemetry

import (
	"detectviz-platform/pkg/platform/contracts"
	"log"
)

// OtelZapLogger 實現了 pkg/platform/contracts.Logger 介面。
// 職責: 提供基於 Zap 庫並集成 OpenTelemetry 的日誌功能。
// 測試說明: 這層的單元測試將專注於驗證其與 Zap/OTel 的集成是否正確，輸出格式是否符合預期。
type OtelZapLogger struct {
	// 實際中會包含 *zap.Logger 實例
	baseLogger *log.Logger // 簡化為標準庫 log，實際為 *zap.Logger
}

// NewOtelZapLogger 構造函數，根據配置初始化日誌級別等。
func NewOtelZapLogger(cfg map[string]interface{}) contracts.Logger {
	level := "info"
	if lvl, ok := cfg["level"].(string); ok {
		level = lvl
	}
	log.Printf("[INFRA][OtelZapLogger] 初始化完成，級別: %s。\n", level)
	return &OtelZapLogger{baseLogger: log.Default()}
}

func (l *OtelZapLogger) GetName() string { return "otelzap" }
func (l *OtelZapLogger) Debug(msg string, fields ...interface{}) {
	l.baseLogger.Printf("[DEBUG] "+msg+"\n", fields...)
}
func (l *OtelZapLogger) Info(msg string, fields ...interface{}) {
	l.baseLogger.Printf("[INFO] "+msg+"\n", fields...)
}
func (l *OtelZapLogger) Warn(msg string, fields ...interface{}) {
	l.baseLogger.Printf("[WARN] "+msg+"\n", fields...)
}
func (l *OtelZapLogger) Error(msg string, fields ...interface{}) {
	l.baseLogger.Printf("[ERROR] "+msg+"\n", fields...)
}
func (l *OtelZapLogger) Fatal(msg string, fields ...interface{}) {
	l.baseLogger.Fatalf("[FATAL] "+msg+"\n", fields...)
}
func (l *OtelZapLogger) WithFields(fields ...interface{}) contracts.Logger {
	// 實際中會返回帶有這些附加字段的新 Logger 實例 (例如 zap.With)
	return l // 簡化
}
func (l *OtelZapLogger) WithContext(ctx interface{}) contracts.Logger {
	// 實際中會從 context 中提取 trace 信息等
	return l // 簡化
}
