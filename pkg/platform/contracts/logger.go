package contracts

// Logger 定義了日誌服務的通用介面。
// 職責: 提供統一的日誌記錄功能，便於調試、監控和問題追蹤。
// AI 擴展點: AI 可生成 `ZapLogger` 或 `LogrusLogger` 等實現。
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{}) // Fatal 會導致程式終止
	WithFields(fields ...interface{}) Logger // 返回一個帶有附加字段的新 Logger 實例。
	WithContext(ctx interface{}) Logger      // 返回一個帶有上下文的新 Logger 實例。
	GetName() string
}
