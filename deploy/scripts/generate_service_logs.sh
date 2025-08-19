#!/bin/bash

# 服務日誌生成腳本
LOG_DIR="/tmp/service-logs"
mkdir -p "$LOG_DIR"

# 生成各種服務的模擬日誌
generate_app_logs() {
    while true; do
        echo "$(date): [INFO] Application processed request $(( RANDOM % 1000 ))" >> "$LOG_DIR/detectviz-app.log"
        sleep $(( RANDOM % 10 + 5 ))
    done
}

generate_db_logs() {
    while true; do
        echo "$(date): [LOG] Database query executed successfully, rows affected: $(( RANDOM % 100 ))" >> "$LOG_DIR/detectviz-db.log"
        sleep $(( RANDOM % 15 + 10 ))
    done
}

generate_api_logs() {
    while true; do
        local status_codes=("200" "201" "400" "404" "500")
        local status=${status_codes[$RANDOM % ${#status_codes[@]}]}
        echo "$(date): [HTTP] API request completed with status $status, response time: $(( RANDOM % 1000 ))ms" >> "$LOG_DIR/detectviz-api.log"
        sleep $(( RANDOM % 8 + 3 ))
    done
}

generate_error_logs() {
    while true; do
        local errors=("Connection timeout" "Memory usage high" "Disk space low" "Rate limit exceeded")
        local error=${errors[$RANDOM % ${#errors[@]}]}
        echo "$(date): [ERROR] $error - investigating issue" >> "$LOG_DIR/detectviz-errors.log"
        sleep $(( RANDOM % 30 + 60 ))
    done
}

echo "Starting service log generation..."
echo "$(date): [INFO] Service log generator started" >> "$LOG_DIR/generator.log"

# 在背景啟動所有日誌生成器
generate_app_logs &
generate_db_logs &
generate_api_logs &
generate_error_logs &

# 等待所有背景程序
wait