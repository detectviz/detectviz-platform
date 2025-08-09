# profiles/local

目標：本地開發。應用只打 OTLP 到 Alloy（4317/4318），Alloy 轉發到本地 LGTM：Loki/Tempo/Mimir/Pyroscope。

服務埠：
- Alloy: 4317(gRPC), 4318(HTTP)
- Loki: 3100（內部）
- Tempo: 4317（內部 OTLP，或 3200 查詢）
- Mimir: 9009（Prometheus Remote Write）
- Pyroscope: 4040（UI/API）

啟動建議：使用 docker compose，先拉起 loki/tempo/mimir/pyroscope，再啟 Alloy。
