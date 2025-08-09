# profiles/grafana-cloud

目標：任意環境透過 Alloy 以 OTLP 直送 Grafana Cloud。

環境變數：
- GRAFANA_CLOUD_OTLP_ENDPOINT：如 `https://otlp-gateway-prod-us-central-0.grafana.net:443`
- GRAFANA_CLOUD_OTLP_AUTH：HTTP Header Authorization 值；格式 `Basic <base64(instanceId:apiKey)>`

執行：
```
docker run -p 4317:4317 -p 4318:4318   -e GRAFANA_CLOUD_OTLP_ENDPOINT=...   -e GRAFANA_CLOUD_OTLP_AUTH="Basic xxx"   -v $(pwd)/alloy/alloy.river:/etc/alloy/config.alloy grafana/alloy:latest run /etc/alloy/config.alloy
```
