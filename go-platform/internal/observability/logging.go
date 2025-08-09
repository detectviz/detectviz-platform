package observability

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerClose func() error

// InitZapLoggerToFile 建議在 process 啟動時呼叫一次
// path 例如: ./var/log/detectviz/detectviz.log
// 說明：
// - 採用 zap ConsoleEncoder（非 JSON），輸出格式接近 Alloy：ts level caller msg key=value ...
// - 仍使用 lumberjack 進行檔案輪轉，並雙寫至 stdout，供 Alloy file tail 轉發至雲端。
// - 自動建立必要的目錄結構
func InitZapLoggerToFile(path string) (*otelzap.Logger, LoggerClose, error) {
	// 確保日誌檔案的目錄存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 確保日誌檔案存在（立即建立空檔案）
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create log file: %w", err)
		}
		file.Close()
	}

	lj := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    50, // MB
		MaxBackups: 7,
		MaxAge:     14, // days
		Compress:   true,
	}

	// 立即寫入一行測試內容確保檔案可寫
	if _, err := lj.Write([]byte("")); err != nil {
		return nil, nil, fmt.Errorf("failed to write to log file: %w", err)
	}

	// 使用 zapcore.Lock 確保同步寫入
	ws := zapcore.Lock(zapcore.AddSync(io.MultiWriter(os.Stdout, lj)))

	enc := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		MessageKey:       "msg",
		CallerKey:        "caller",
		EncodeTime:       zapcore.RFC3339NanoTimeEncoder,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "\t", // 用 tab 分隔，與 zap 預設 ConsoleEncoder 一致
	})

	core := zapcore.NewCore(enc, ws, zapcore.InfoLevel)
	zl := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 附加與 Alloy 語意相近的共用欄位
	env := os.Getenv("DEPLOYMENT_ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "dev"
	}
	zl = zl.With(
		zap.String("service", "go-platform"),
		zap.String("component", "toolbridge"),
		zap.String("env", env),
	)

	// 關鍵：WithTraceIDField 選項已被 upstream 移除，暫改為手動注入欄位（TraceFields）。
	ol := otelzap.New(
		zl,
		otelzap.WithMinLevel(zap.InfoLevel),
		otelzap.WithErrorStatusLevel(zap.ErrorLevel),
	)

	// 同時設定 otelzap 和 zap 的全域 logger
	otelzap.ReplaceGlobals(ol)
	zap.ReplaceGlobals(zl)

	// 自定義 close 函數，確保所有緩衝的日誌都被刷新
	closeFunc := func() error {
		// 先同步 zap logger
		zl.Sync()
		// 再關閉 lumberjack
		return lj.Close()
	}

	return ol, closeFunc, nil
}

// TraceFields 會從 context 擷取目前 Span 的 trace_id/span_id，並以 zap 欄位回傳。
// 用法：otelzap.Ctx(ctx).Info("handled", TraceFields(ctx)...)
func TraceFields(ctx context.Context) []zap.Field {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return nil
	}
	return []zap.Field{
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	}
}
