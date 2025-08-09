# profiles/gcp

目標：服務部署於 GKE/VM/Cloud Run，資料上送 Google Cloud（Cloud Trace / Monitoring / Logging）。

兩種方式：
1) Alloy 只做接收，**轉送**到旁路 OTel Collector（推薦企業既有集中 Collector 的場景）。
2) Alloy 直接使用 `googlecloud` exporter（需 Alloy 版本內建該 exporter）。

範例均提供。

必要環境：
- 使用 Workload Identity 或服務帳戶具備 `Cloud Trace Agent`、`Monitoring Metric Writer`、`Logs Writer` 角色。
