# Detectviz API 參考

本文件檔詳細介紹了 Detectviz 平台提供的所有 API 端點。隨著新插件的添加，此 API 列表將會不斷擴充。

## 基礎 API

### 健康檢查

*   **端點**: `GET /health`
*   **描述**: 檢查平台的運行狀況。
*   **請求**: 無
*   **響應**: `200 OK`

    ```json
    {
        "status": "healthy",
        "service": "detectviz-platform",
        "timestamp": "2023-10-27T10:00:00Z"
    }
    ```

### 平台資訊

*   **端點**: `GET /api/v1/info`
*   **描述**: 獲取有關平台的版本和狀態的基本資訊。
*   **請求**: 無
*   **響應**: `200 OK`

    ```json
    {
        "name": "Detectviz Platform",
        "version": "0.1.0",
        "status": "running"
    }
    ```

## 未來的 API

隨著新插件的開發，將會添加更多的 API 端點。例如，數據導入插件可能會添加用於上傳數據的端點，而分析插件可能會添加用於觸發分析和檢索結果的端點。

有關如何為插件創建新的 API 端點的資訊，請參閱 `PLUGIN_GUIDE.md`。
