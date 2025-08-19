#!/bin/bash

# Detectviz Platform - Development Environment Setup Script
set -e

cd "$(dirname "$0")/.."

echo "🚀 Detectviz Platform - 環境設置開始"
echo "=================================="

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 檢查 Docker 和 Docker Compose
check_docker() {
    echo -e "${YELLOW}檢查 Docker 環境...${NC}"
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker 未安裝${NC}"
        echo "請先安裝 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose 未安裝${NC}"
        echo "請先安裝 Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker 環境檢查通過${NC}"
}

# 創建必要的目錄結構
create_directories() {
    echo -e "${YELLOW}創建目錄結構...${NC}"
    
    # 配置目錄
    mkdir -p configs/prometheus
    mkdir -p configs/grafana/{provisioning/{dashboards,datasources},dashboards}
    mkdir -p configs/loki
    mkdir -p configs/tempo
    mkdir -p configs/grafana-alloy
    mkdir -p configs/postgres
    
    # 數據目錄（可選，Docker 會自動創建 volumes）
    mkdir -p data/{prometheus,grafana,postgres,redis,loki,tempo,pyroscope,alloy}
    
    echo -e "${GREEN}✅ 目錄結構創建完成${NC}"
}

# 創建 Grafana 數據源配置
create_grafana_datasources() {
    echo -e "${YELLOW}配置 Grafana 數據源...${NC}"
    
    cat > configs/grafana/provisioning/datasources/datasources.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: true

  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    editable: true

  - name: Pyroscope
    type: pyroscope
    access: proxy
    url: http://pyroscope:4040
    editable: true

  - name: PostgreSQL
    type: postgres
    access: proxy
    url: postgres:5432
    database: detectviz
    user: detectviz
    secureJsonData:
      password: detectviz123
    jsonData:
      sslmode: disable
    editable: true

  - name: Redis
    type: redis-datasource
    access: proxy
    url: redis://redis:6379
    editable: true
EOF
    
    echo -e "${GREEN}✅ Grafana 數據源配置完成${NC}"
}

# 創建 Loki 配置
create_loki_config() {
    echo -e "${YELLOW}配置 Loki...${NC}"
    
    cat > configs/loki/loki-config.yaml << 'EOF'
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9095

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://localhost:9093

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 10
  ingestion_burst_size_mb: 20
EOF
    
    echo -e "${GREEN}✅ Loki 配置完成${NC}"
}

# 創建 Tempo 配置
create_tempo_config() {
    echo -e "${YELLOW}配置 Tempo...${NC}"
    
    cat > configs/tempo/tempo-config.yaml << 'EOF'
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:9095
        http:
          endpoint: 0.0.0.0:4318
    jaeger:
      protocols:
        thrift_http:
          endpoint: 0.0.0.0:14268
    zipkin:
      endpoint: 0.0.0.0:9411

ingester:
  max_block_duration: 5m

compactor:
  compaction:
    block_retention: 48h

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/blocks
    wal:
      path: /tmp/tempo/wal

metrics_generator:
  registry:
    external_labels:
      source: tempo
      cluster: docker-compose
  storage:
    path: /tmp/tempo/generator/wal
    remote_write:
      - url: http://prometheus:9090/api/v1/write
        send_exemplars: true
EOF
    
    echo -e "${GREEN}✅ Tempo 配置完成${NC}"
}

# 創建 PostgreSQL 初始化腳本
create_postgres_init() {
    echo -e "${YELLOW}配置 PostgreSQL...${NC}"
    
    cat > configs/postgres/init.sql << 'EOF'
-- Detectviz Platform Database Initialization

-- Create schema for knowledge base
CREATE SCHEMA IF NOT EXISTS knowledge_base;

-- Create incidents table
CREATE TABLE IF NOT EXISTS knowledge_base.incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id VARCHAR(255) UNIQUE NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    severity VARCHAR(50),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_minutes INTEGER,
    root_cause TEXT,
    impact TEXT,
    resolution TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create lessons learned table
CREATE TABLE IF NOT EXISTS knowledge_base.lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID REFERENCES knowledge_base.incidents(id),
    category VARCHAR(100),
    lesson TEXT NOT NULL,
    action_items JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create metrics snapshot table
CREATE TABLE IF NOT EXISTS knowledge_base.metrics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID REFERENCES knowledge_base.incidents(id),
    metric_name VARCHAR(255),
    metric_value JSONB,
    timestamp TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_incidents_service ON knowledge_base.incidents(service_name);
CREATE INDEX idx_incidents_time ON knowledge_base.incidents(start_time, end_time);
CREATE INDEX idx_lessons_incident ON knowledge_base.lessons(incident_id);
CREATE INDEX idx_metrics_incident ON knowledge_base.metrics_snapshots(incident_id);

-- Create similarity search function (basic implementation)
CREATE OR REPLACE FUNCTION knowledge_base.find_similar_incidents(
    p_service_name VARCHAR,
    p_root_cause TEXT,
    p_limit INTEGER DEFAULT 5
)
RETURNS TABLE (
    incident_id VARCHAR,
    service_name VARCHAR,
    root_cause TEXT,
    similarity_score FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        i.incident_id,
        i.service_name,
        i.root_cause,
        similarity(i.root_cause, p_root_cause) as similarity_score
    FROM knowledge_base.incidents i
    WHERE i.service_name = p_service_name
        AND i.root_cause IS NOT NULL
    ORDER BY similarity_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA knowledge_base TO detectviz;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA knowledge_base TO detectviz;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA knowledge_base TO detectviz;
EOF
    
    echo -e "${GREEN}✅ PostgreSQL 初始化腳本創建完成${NC}"
}

# 創建 .env 文件
create_env_file() {
    echo -e "${YELLOW}創建環境變數文件...${NC}"
    
<<<<<<< HEAD
    if [ -f .env ]; then
        echo -e "${YELLOW}⚠️  .env 文件已存在，跳過創建${NC}"
    else
        cat > .env << 'EOF'
=======
    if [ -f ../.env ]; then
        echo -e "${YELLOW}⚠️  .env 文件已存在，跳過創建${NC}"
    else
        cat > ../.env << 'EOF'
>>>>>>> 08fa581 (update)
# Detectviz Platform Environment Variables

# Core Configuration
DETECTVIZ_ENV=development
HOSTNAME=detectviz-dev

# Metrics Provider
METRICS_PROVIDER_TYPE=prometheus
PROMETHEUS_URL=http://localhost:9090

# Grafana
GRAFANA_URL=http://localhost:3000
GRAFANA_API_KEY=generate_your_api_key_here

# PostgreSQL (Knowledge Base)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=detectviz
POSTGRES_USER=detectviz
POSTGRES_PASSWORD=detectviz123

# Redis (State Management)
REDIS_ADDRESS=localhost:6379
REDIS_DB=0

# Observability
DETECTVIZ__OBSERVABILITY__MODE=lgtm_local
DETECTVIZ__OBSERVABILITY__OTLP__ENDPOINT=localhost:4317

# Network
DETECTVIZ__GRPC__LISTEN=:6606
DETECTVIZ__HTTP__HEALTH__LISTEN=:8081

# Model Provider (需要您自己的 API Key)
ADK_MODEL_PROVIDER=gemini
ADK_MODEL_API_KEY=your-gemini-api-key-here
EOF
        echo -e "${GREEN}✅ .env 文件創建完成${NC}"
        echo -e "${YELLOW}⚠️  請編輯 .env 文件，填入您的 API Keys${NC}"
    fi
}

# 啟動服務
start_services() {
    echo -e "${YELLOW}啟動 Docker 服務...${NC}"
    
    # 基礎服務優先啟動
    echo "啟動基礎服務..."
    docker-compose up -d postgres redis
    sleep 5
    
    # 監控服務
    echo "啟動監控服務..."
    docker-compose up -d prometheus grafana
    sleep 5
    
    # 可觀測性服務
    echo "啟動可觀測性服務..."
    docker-compose up -d loki tempo pyroscope alloy
    sleep 3
    
    # Exporters
    echo "啟動 Exporters..."
    docker-compose up -d node-exporter postgres-exporter redis-exporter
    
    echo -e "${GREEN}✅ 所有服務啟動完成${NC}"
}

# 檢查服務狀態
check_services() {
    echo -e "${YELLOW}檢查服務狀態...${NC}"
    
    docker-compose ps
    
    echo ""
    echo -e "${GREEN}服務訪問地址：${NC}"
    echo "=================================="
    echo "📊 Grafana:      http://localhost:3000 (admin/admin123)"
    echo "📈 Prometheus:   http://localhost:9090"
    echo "🔍 Alloy UI:     http://localhost:12345"
    echo "📝 Loki:         http://localhost:3100"
    echo "🔬 Tempo:        http://localhost:3200"
    echo "🔥 Pyroscope:    http://localhost:4040"
    echo "🗄️  PostgreSQL:   localhost:5432 (detectviz/detectviz123)"
    echo "💾 Redis:        localhost:6379"
    echo "=================================="
}

# 主函數
main() {
    echo "選擇操作："
    echo "1) 完整設置（第一次運行）"
    echo "2) 僅啟動服務"
    echo "3) 停止所有服務"
    echo "4) 清理所有數據（危險！）"
    echo "5) 檢查服務狀態"
    read -p "請選擇 (1-5): " choice
    
    case $choice in
        1)
            check_docker
            create_directories
            create_grafana_datasources
            create_loki_config
            create_tempo_config
            create_postgres_init
            create_env_file
            
            # 複製 prometheus.yml
            if [ -f configs/prometheus.yml ]; then
                cp configs/prometheus.yml configs/prometheus/
            fi
            
            start_services
            check_services
            echo -e "${GREEN}🎉 環境設置完成！${NC}"
            ;;
        2)
            docker-compose up -d
            check_services
            ;;
        3)
            docker-compose down
            echo -e "${GREEN}✅ 服務已停止${NC}"
            ;;
        4)
            read -p "⚠️  確定要清理所有數據嗎？(yes/no): " confirm
            if [ "$confirm" = "yes" ]; then
                docker-compose down -v
                rm -rf data/*
                echo -e "${GREEN}✅ 數據已清理${NC}"
            fi
            ;;
        5)
            check_services
            ;;
        *)
            echo -e "${RED}無效的選擇${NC}"
            exit 1
            ;;
    esac
}

# 執行主函數
main
