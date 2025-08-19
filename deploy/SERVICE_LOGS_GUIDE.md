# 服務日誌收集指南

## 當前狀態
✅ Alloy 檔案日誌收集已正常工作
✅ OTLP 日誌 API 正常工作  
✅ Profile 收集正常工作
✅ 測試日誌正常被收集並發送到 Grafana Cloud

## 服務日誌收集解決方案

### 方案 1: 應用程式日誌檔案輸出 (推薦)

#### 1.1 修改服務容器配置
在需要日誌收集的服務容器中添加 volume 映射：

```yaml
# 在 docker-compose.yml 中
your-service:
  image: your-app:latest
  volumes:
    - /tmp/service-logs:/app/logs  # 映射日誌目錄
  environment:
    - LOG_FILE_PATH=/app/logs/your-service.log
```

#### 1.2 應用程式日誌配置
```go
// Go 應用範例
logFile, err := os.OpenFile("/app/logs/detectviz-app.log", 
    os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatalln("Failed to open log file:", err)
}
log.SetOutput(logFile)

// 或使用 structured logging
logger := logrus.New()
logger.SetOutput(logFile)
logger.SetFormatter(&logrus.JSONFormatter{})
```

```python
# Python 應用範例
import logging

logging.basicConfig(
    filename='/app/logs/detectviz-app.log',
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(message)s'
)
```

### 方案 2: OTLP 日誌直接發送 (推薦用於新服務)

#### 2.1 Go 應用 OTLP 整合
```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
    "go.opentelemetry.io/otel/log/global"
    sdklog "go.opentelemetry.io/otel/sdk/log"
    "go.opentelemetry.io/otel/sdk/resource"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initOTLPLogging() {
    // Create OTLP log exporter
    exporter, err := otlploggrpc.New(context.Background(),
        otlploggrpc.WithEndpoint("localhost:4317"),
        otlploggrpc.WithInsecure(),
    )
    if err != nil {
        panic(err)
    }

    // Create resource
    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String("detectviz-service"),
        semconv.ServiceVersionKey.String("1.0.0"),
    )

    // Create log processor and provider
    processor := sdklog.NewBatchProcessor(exporter)
    provider := sdklog.NewLoggerProvider(
        sdklog.WithResource(res),
        sdklog.WithProcessor(processor),
    )
    global.SetLoggerProvider(provider)

    // Get logger and log
    logger := global.GetLoggerProvider().Logger("detectviz")
    logger.Emit(context.Background(), sdklog.Record{
        Timestamp: time.Now(),
        Body: log.StringValue("Service started successfully"),
        Severity: log.SeverityInfo,
    })
}
```

#### 2.2 Python 應用 OTLP 整合
```python
from opentelemetry import logs
from opentelemetry.exporter.otlp.proto.grpc.logs_exporter import OTLPLogsExporter
from opentelemetry.sdk.logs import LoggingHandler, LoggerProvider
from opentelemetry.sdk.logs.export import BatchLogRecordProcessor
from opentelemetry.sdk.resources import Resource
import logging

def init_otlp_logging():
    # Create resource
    resource = Resource.create({
        "service.name": "detectviz-python-service",
        "service.version": "1.0.0"
    })
    
    # Create OTLP exporter
    exporter = OTLPLogsExporter(
        endpoint="http://localhost:4317",
        insecure=True
    )
    
    # Create logger provider
    logger_provider = LoggerProvider(resource=resource)
    logger_provider.add_log_record_processor(
        BatchLogRecordProcessor(exporter)
    )
    logs.set_logger_provider(logger_provider)
    
    # Setup Python logging to use OTLP
    handler = LoggingHandler(
        level=logging.INFO,
        logger_provider=logger_provider
    )
    
    # Configure root logger
    logging.getLogger().addHandler(handler)
    logging.getLogger().setLevel(logging.INFO)
    
    # Test logging
    logging.info("Python service started with OTLP logging")

# 在應用啟動時調用
init_otlp_logging()

# 正常使用 logging
logging.info("This log will be sent via OTLP")
logging.error("Error messages are also collected")
```

### 方案 3: Docker 日誌驅動程式 (系統級)

#### 3.1 配置 Docker 日誌驅動
```yaml
# docker-compose.yml
your-service:
  image: your-app:latest
  logging:
    driver: "local"
    options:
      max-size: "10m"
      max-file: "3"
  # 或直接輸出到檔案
  command: >
    sh -c "your-app 2>&1 | tee /tmp/service-logs/your-service.log"
```

## 測試和驗證

### 測試檔案日誌收集
```bash
# 添加測試日誌
echo "$(date): [INFO] Application started" >> /tmp/service-logs/your-service.log

# 檢查 Alloy 是否收集
docker logs detectviz-alloy | grep "tail routine"
```

### 測試 OTLP 日誌發送
```bash
# 直接發送測試日誌到 OTLP
curl -X POST http://localhost:4318/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "resourceLogs": [{
      "resource": {
        "attributes": [{
          "key": "service.name",
          "value": {"stringValue": "test-service"}
        }]
      },
      "scopeLogs": [{
        "logRecords": [{
          "timeUnixNano": "'$(date +%s)000000000'",
          "severityText": "INFO",
          "body": {"stringValue": "Direct OTLP test log"}
        }]
      }]
    }]
  }'
```

## 建議的實施順序

1. **立即解決**: 使用檔案日誌方案，修改現有服務輸出日誌到 `/tmp/service-logs/`
2. **中期規劃**: 新服務直接使用 OTLP 日誌整合
3. **長期目標**: 全面遷移到 OTLP 原生日誌收集

## 檢查清單

- [ ] 服務容器添加 volume 映射到 `/tmp/service-logs/`
- [ ] 應用程式配置日誌檔案輸出
- [ ] 測試日誌檔案是否正確寫入
- [ ] 確認 Alloy 能讀取服務日誌檔案
- [ ] 在 Grafana Cloud 中驗證日誌接收