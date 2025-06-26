# **Detectviz 平台部署指南**

本文件提供 Detectviz 平台部署到生產環境的詳細指南，涵蓋了基礎設施規劃、配置、部署步驟以及關鍵的運維考慮，特別強調負載均衡、可觀測性、健康檢查與服務發現。

## **1. 部署概覽**

Detectviz 平台設計為雲原生應用，推薦部署在 **Kubernetes (K8s)** 環境中，以充分利用其容器編排、服務發現、自動擴展和自我修復能力。

**核心部署流程**：

1. **基礎設施準備**：設定 Kubernetes 集群、VPC 網路、資料庫、消息佇列等。  
2. **配置管理**：準備 app_config.yaml 和 composition.yaml，並安全地管理敏感配置。  
3. **容器化**：將 Detectviz 應用打包為 Docker 映像檔。  
4. **Kubernetes 部署**：使用 Helm 或 K8s Manifests 部署應用程式、外部服務連接器和可觀測性組件。  
5. **流量管理**：配置 Ingress Controller 和 Service Load Balancer。  
6. **可觀測性配置**：設定日誌、指標和追蹤的收集與可視化。

## **2. 基礎設施準備**

### **2.1 Kubernetes 集群**

* **推薦**：使用託管的 Kubernetes 服務（如 GKE, EKS, AKS）以降低運維負擔。  
* **版本**：確保 Kubernetes 版本符合應用程式和相關工具（Helm, Ingress Controller）的要求。  
* **節點類型**：根據應用程式的資源需求選擇合適的節點類型和數量，考慮 CPU、記憶體、網路帶寬。

### **2.2 資料庫**

* **MySQL**：推薦使用雲服務商提供的託管 MySQL 服務（如 Cloud SQL for MySQL, AWS RDS for MySQL），以獲得高可用性、備份和擴展性。  
* **配置**：準備好資料庫連接 DSN (Data Source Name)，並通過 Secrets Provider 安全地注入到應用程式配置中。

### **2.3 消息佇列 (NATS)**

* **NATS JetStream**：部署 NATS 集群，推薦在 K8s 上使用官方 Helm Chart 或 Operator 進行部署，以獲得持久化和高可用性。  
* **配置**：確保 Detectviz 應用程式能夠通過服務發現連接到 NATS 集群。

### **2.4 緩存 (Redis)**

* **Redis**：推薦使用託管的 Redis 服務（如 Memorystore for Redis, AWS ElastiCache for Redis）或在 K8s 上部署 Redis Cluster。  
* **配置**：確保 Detectviz 應用程式能夠連接到 Redis 實例。

## **3. 配置與秘密管理**

### **3.1 應用程式配置 (app_config.yaml, composition.yaml)**

* **版本控制**：所有配置檔案必須納入 Git 版本控制，並遵循 GitOps 原則進行管理。  
* **環境特定配置**：為不同環境（開發、測試、生產）維護獨立的配置版本，或使用 Helm Values 檔案來管理環境差異。  
* **JSON Schema 驗證**：在 CI/CD 流程中，自動驗證配置檔案是否符合其 JSON Schema，防止部署無效配置。

### **3.2 秘密管理 (Secrets Management)**

* **嚴禁硬編碼**：資料庫密碼、API 金鑰等敏感資訊嚴禁硬編碼在任何配置檔案或程式碼中。  
* **推薦工具**：  
  * **Kubernetes Secrets**：用於存儲非機密但需要加密的基本敏感數據。  
  * **外部秘密管理系統**：對於高度機密的數據或跨雲環境，推薦使用：  
    * **Vault (HashiCorp)**：強大的秘密管理工具，提供動態秘密、租賃和審計功能。  
    * **雲服務商 Secrets Manager**：例如 AWS Secrets Manager, Google Cloud Secret Manager。  
  * **Secret Injection**：通過 CSI (Container Storage Interface) 驅動或 Sidecar 模式將秘密安全地注入到 Pod 中。  
* **使用規範**：應用程式應通過 contracts.SecretsProvider 介面來讀取秘密，不直接訪問底層存儲。

## **4. 負載均衡與流量管理**

### **4.1 Kubernetes Service**

* **ClusterIP**：用於集群內部服務間的通信，不對外暴露。  
* **NodePort**：允許通過 K8s 節點的 IP 和指定端口訪問服務（主要用於開發/測試）。  
* **LoadBalancer**：在雲環境中，自動創建一個雲供應商的負載均衡器，將外部流量分發到後端 Pod。推薦用於對外暴露的服務。

### **4.2 Ingress Controller**

* **推薦**：部署一個 Ingress Controller (例如 Nginx Ingress, Traefik, Istio Ingress Gateway)，作為所有 HTTP/HTTPS 流量的統一入口。  
* **職責**：  
  * **流量路由**：根據域名、路徑將請求路由到不同的後端 Service。  
  * **SSL/TLS 終止**：處理 HTTPS 加密，簡化後端服務配置。  
  * **負載均衡**：提供基本的 L7 負載均衡。  
  * **URL 重寫、流量分割**：實現 A/B 測試、金絲雀發布等高級流量管理。  
  * **認證集成**：與外部認證服務集成，統一身份驗證。  
  * **速率限制**：在 Ingress 層實施請求限流。  
* **配置範例 (Ingress YAML)**：  
  apiVersion: networking.k8s.io/v1  
  kind: Ingress  
  metadata:  
    name: detectviz-ingress  
    annotations:  
      nginx.ingress.kubernetes.io/rewrite-target: /$1  
      # 其他速率限制、CORS 等 annotation  
  spec:  
    rules:  
    - host: api.detectviz.com # 您的 API 域名  
      http:  
        paths:  
        - path: /api/(.*)  
          pathType: Prefix  
          backend:  
            service:  
              name: detectviz-api-service  
              port:  
                number: 8080  
    - host: ui.detectviz.com # 您的 UI 域名  
      http:  
        paths:  
        - path: /(.*)  
          pathType: Prefix  
          backend:  
            service:  
              name: detectviz-ui-service  
              port:  
                number: 80

## **5. 可觀測性 (Observability)**

Detectviz 平台將集成全面的可觀測性工具棧，以實現對系統的實時監控、故障排查和性能優化。

### **5.1 日誌收集與分析 (Logs)**

* **日誌庫**：應用程式使用 **Zap** 庫輸出結構化 JSON 日誌。  
* **收集器**：在每個 K8s 節點上部署 **Promtail** (或 Fluentd/Fluent Bit) 作為日誌收集代理。  
* **集中式日誌管理**：日誌由 Promtail 收集後，統一發送到 **Grafana Loki**。  
* **可視化與查詢**：通過 **Grafana** (連接 Loki 數據源) 進行日誌的聚合、查詢和可視化。  
* **使用規範**：  
  * 所有日誌輸出必須為結構化 JSON 格式。  
  * 日誌應包含標準字段 (timestamp, level, caller, message, trace_id, span_id) 和業務相關的自定義字段。  
  * 應用程式的日誌級別應可通過配置動態調整（例如：生產環境 INFO，調試環境 DEBUG）。

### **5.2 指標可視化與警報 (Metrics)**

* **指標庫**：應用程式通過 **OpenTelemetry Metrics** 導出指標。  
* **收集器**：在 K8s 集群中部署 **Prometheus**，配置其從應用程式的 /metrics 端點抓取 (scrape) 指標。  
* **可視化**：在 **Grafana** (連接 Prometheus 數據源) 中創建儀表板，可視化關鍵性能指標 (KPIs) 和系統健康狀態。  
* **警報配置**：在 **Prometheus Alertmanager** 中定義警報規則，當指標達到預設閾值時，通過郵件、Slack、PagerDuty 等通道發送警報。  
* **使用規範**：  
  * 定義關鍵業務指標（例如：請求總量、錯誤率、平均響應時間、登入成功率）。  
  * 定義基礎設施指標（例如：CPU/記憶體使用率、網路流量、磁碟 I/O）。  
  * 所有指標應具備統一的命名規範和標籤 (labels)，便於查詢和聚合。

### **5.3 追蹤可視化 (Tracing)**

* **追蹤庫**：應用程式通過 **OpenTelemetry Tracing** 導出分散式追蹤數據。  
* **收集器**：部署 **OpenTelemetry Collector**，用於接收、處理和導出追蹤數據。  
* **可視化**：OpenTelemetry Collector 將追蹤數據發送到 **Jaeger**。在 Jaeger UI 中可視化服務間的請求流、延遲瓶頸和錯誤。  
* **使用規範**：  
  * 確保所有服務間調用、數據庫操作、消息發布/消費和外部 API 調用都能生成 Span。  
  * 關鍵業務邏輯應添加自定義 Span，包含相關屬性，以便深入分析。  
  * Trace ID 和 Span ID 應被注入到日誌中，以便在日誌和追蹤之間進行關聯分析。

## **6. 健康檢查與服務發現**

### **6.1 健康檢查端點**

* **Liveness Probe (存活探針)**：  
  * **目的**：判斷 Pod 是否正常運行。如果 Liveness Probe 失敗，K8s 將重啟該 Pod。  
  * **配置**：  
    * **類型**：HTTP GET  
    * **路徑**：/healthz (或 /health/liveness)  
    * **應用行為**：應用程式在此端點應只檢查自身是否還在運行（例如：HTTP 伺服器是否響應）。不應檢查外部依賴，避免外部依賴故障導致應用程式無限重啟。  
    * **範例 (Deployment YAML)**：  
      livenessProbe:  
        httpGet:  
          path: /healthz  
          port: 8080  
        initialDelaySeconds: 15 # Pod 啟動後等待 15 秒開始探測  
        periodSeconds: 10 # 每 10 秒探測一次  
        timeoutSeconds: 5 # 探測超時時間 5 秒  
        failureThreshold: 3 # 連續失敗 3 次後重啟 Pod

* **Readiness Probe (就緒探針)**：  
  * **目的**：判斷 Pod 是否準備好接收流量。如果 Readiness Probe 失敗，K8s 將從 Service 的 Endpoint 列表中移除該 Pod，停止向其發送流量。  
  * **配置**：  
    * **類型**：HTTP GET  
    * **路徑**：/readyz (或 /health/readiness)  
    * **應用行為**：應用程式在此端點應檢查所有**關鍵外部依賴**（例如：資料庫連接、NATS 連接、外部認證服務）是否可用。只有所有依賴都就緒，應用程式才被視為就緒。  
    * **範例 (Deployment YAML)**：  
      readinessProbe:  
        httpGet:  
          path: /readyz  
          port: 8080  
        initialDelaySeconds: 20 # Pod 啟動後等待 20 秒開始探測  
        periodSeconds: 10 # 每 10 秒探測一次  
        timeoutSeconds: 5 # 探測超時時間 5 秒  
        failureThreshold: 3 # 連續失敗 3 次後將 Pod 從 Service 中移除

### **6.2 服務發現 (Service Discovery)**

* **Kubernetes 內建服務發現**：  
  * **DNS**：Detectviz 平台內的微服務應直接使用 Kubernetes 的 DNS 服務發現機制。例如，如果有一個名為 detectviz-db-service 的 Service，應用程式可以直接通過 detectviz-db-service 或 detectviz-db-service.namespace.svc.cluster.local 來訪問它。  
  * **Endpoint Slice**：K8s 自動維護 Service 對應的 Pod Endpoint 列表。  
* **跨集群/外部服務發現**：  
  * 對於需要訪問 K8s 集群外部服務（例如託管數據庫、第三方 API）或跨 K8s 集群的服務，應通過 contracts.ServiceDiscoveryProvider 介面進行抽象。  
  * 具體實作可以集成 **Consul** 或 **Etcd** 等外部服務發現工具，或利用雲供應商的服務發現功能。  
* **使用規範**：  
  * 應用程式應盡量避免硬編碼服務地址。  
  * 對於 K8s 內部服務，優先使用 K8s DNS 進行發現。  
  * 對於外部或異構環境，使用抽象的 ServiceDiscoveryProvider。

## **7. 持續集成與持續部署 (CI/CD)**

* **GitOps 原則**：推薦採用 GitOps 工作流，將所有基礎設施和應用程式的配置（包括 K8s Manifests, Helm Charts, 配置檔案）都存儲在 Git 儲存庫中。  
* **CI/CD 工具**：  
  * **GitHub Actions / GitLab CI / Jenkins / Tekton**：用於自動化構建、測試、安全掃描、Docker 映像檔構建和推送到容器註冊表。  
  * **Flux / Argo CD**：作為 K8s 集群內的 GitOps Operator，持續監控 Git 儲存庫的變更，並自動將變更同步到集群中。  
* **部署階段**：  
  1. **程式碼提交**：開發者提交程式碼到 Git。  
  2. **CI 構建**：CI Pipeline 執行單元測試、集成測試、靜態分析、安全性掃描，並構建 Docker 映像檔。  
  3. **映像檔發布**：將 Docker 映像檔推送到容器註冊表（如 Docker Hub, GCR, ECR）。  
  4. **CD 觸發**：當新的映像檔發布或 K8s 配置更新時，CD Pipeline 被觸發。  
  5. **GitOps 同步**：Flux/Argo CD 檢測到 Git 儲存庫中 K8s Manifests 的變更（指向新映像檔），並將這些變更應用到集群。  
  6. **自動化遷移**：在應用程式啟動前或作為 CI/CD 的一部分，自動執行資料庫遷移（使用 Atlas）。

## **8. 結論**

Detectviz 平台的部署遵循雲原生最佳實踐，旨在構建一個自動化、可擴展且可靠的基礎設施。通過標準化的配置、健全的可觀測性以及自動化的 CI/CD 流程，我們能夠確保平台在生產環境中高效穩定運行，並快速響應業務需求。本指南將作為團隊部署和運維 Detectviz 平台的核心參考，並會隨著時間推移持續更新和完善。

