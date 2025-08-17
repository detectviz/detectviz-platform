# go-platform/cmd

本目錄提供單一入口 CLI：`detectviz`。

## 子命令
- `detectviz plugin serve`：啟動 ToolBridge gRPC（對齊 adk_bridge.proto）
- `detectviz plugin new <category>/<name>`：於 `go-platform/internal/pluginhost/plugins/` 產生 **Go 插件**骨架
- `detectviz plugin validate <path>`：提示使用 `contracts/tools/validate_module_card.py`
- `detectviz config validate -f <config.yaml>`：驗證平台設定

## 建置
```bash
go build -o bin/detectviz ./go-platform/cmd/detectviz
```

## 啟動（範例）
```bash
bin/detectviz plugin serve --listen :6606 --config ./configs/config.yaml
```

## 產生插件骨架（範例）
```bash
bin/detectviz plugin new gateway/http_echo
# 會建立：go-platform/internal/pluginhost/plugins/gateway/http_echo/{plugin.go,module.card.json,README.md,tests/...}
```
