### **Detectviz AI Scaffolding Workflow (檢測視覺 AI 腳手架工作流程)**

本文件描述了 Detectviz 平台 AI 腳手架的工作流程，指導 AI 如何根據輸入的配置（例如 composition.yaml）和定義的模板來生成或修改程式碼。

#### **目標**

* **自動化平台組件組裝：** 根據 composition.yaml 定義，自動生成 main.go 中所有插件和提供者的初始化、依賴注入與組裝邏輯。  
* **確保程式碼一致性：** 使用預定義的程式碼模板，保證生成程式碼的結構和風格符合平台規範。  
* **提高開發效率：** 減少手動編寫重複性組裝程式碼的工作量。

#### **輸入**

* composition.yaml: 描述了 Detectviz 平台中所有需要組裝的插件 (plugins) 和提供者 (providers)，包括它們的名稱、類型、配置參數以及依賴關係。  
  * **範例結構參考：**  
    # composition.yaml (範例片段)  
    plugins:  
      logger_otelzap:  
        type: logger  
        package_path: detectviz-platform/internal/platform/plugins/logger/otelzap_logger  
      config_viper:  
        type: config  
        package_path: detectviz-platform/internal/platform/plugins/config/viper_config  
      llm_gemini:  
        type: llm_provider  
        package_path: detectviz-platform/internal/platform/providers/llm/gemini_provider

    composition:  
      - name: logger  
        plugin_id: logger_otelzap  
        config:  
          level: info  
        dependencies: []  
      - name: config  
        plugin_id: config_viper  
        config: {}  
        dependencies: []  
      - name: llmProvider  
        plugin_id: llm_gemini  
        config:  
          api_key_env: GEMINI_API_KEY  
        dependencies:  
          - name: logger  
            as: contracts.Logger  
          - name: config  
            as: contracts.ConfigProvider

* docs/templates/ai_scaffolding/main_go_assembly.tmpl: main.go 程式碼的骨架模板，包含用於動態生成部分的佔位符。  
* **docs/architecture/interface_spec.md (參考文件):** **新增！** 這份文件包含了 Detectviz 平台所有核心 Go 介面（來自 pkg/domain/interfaces 和 pkg/platform/contracts）的規範性定義。AI 應參考此文件，以確保在組裝過程中進行正確的介面類型斷言和理解組件契約。

#### **輸出**

* main.go: 包含所有必要導入、插件工廠註冊以及插件/提供者初始化與組裝邏輯的完整 Go 檔案。

#### **工作流程步驟**

1. **讀取和解析輸入：**  
   * **讀取 composition.yaml:** 載入並解析 composition.yaml 檔案的內容，將其轉換為可供處理的數據結構（例如 JSON 或 Go struct）。  
   * **讀取 main_go_assembly.tmpl:** 載入 docs/templates/ai_scaffolding/main_go_assembly.tmpl 模板檔案。  
   * **讀取 docs/architecture/interface_spec.md:** 載入這份規範性文件，用於 AI 理解平台介面結構和其用途。  
2. **生成動態導入 (Dynamic Imports)：**  
   * **遍歷 composition.yaml 中的 plugins 區塊：** 收集所有 package_path 值。  
   * **產生導入語句：** 對於每個 package_path，生成一個 Go import 語句。  
   * **填充 {{.DynamicPluginImports}} 佔位符：** 將生成的導入語句插入到模板中 {{.DynamicPluginImports}} 的位置。  
3. **生成插件工廠註冊程式碼 (Plugin Factory Registration Code)：**  
   * **遍歷 composition.yaml 中的 plugins 區塊：** 對於每個插件，根據其 plugin_id 和 package_path 生成對應的工廠註冊程式碼。  
   * **假定工廠命名規則：** 假設每個插件包都提供一個 NewFactory() 函數。  
   * **產生註冊語句：** 產生類似 registry.RegisterPluginFactory("plugin_id", package_name.NewFactory()) 的程式碼。  
   * **填充 {{.PluginFactoryRegistrationCode}} 佔位符：** 將生成的註冊語句插入到模板中 {{.PluginFactoryRegistrationCode}} 的位置。  
4. **生成插件實例初始化與組裝程式碼 (Plugin Assembly Code)：**  
   * **分析依賴圖：** 遍歷 composition.yaml 中的 composition 區塊。對於每個組件，識別其 dependencies。構建一個有向無環圖 (DAG) 來表示組件之間的依賴關係。  
   * **拓撲排序：** 對依賴圖進行拓撲排序，確定組件的正確初始化順序。這確保在初始化一個組件之前，其所有依賴都已準備好。  
   * **生成初始化和注入邏輯：**  
     * **容器初始化：** 在組裝開始前，生成一個 components := make(map[string]any) 類似的映射，用於儲存已初始化的組件實例。  
     * **遍歷排序後的組件：**  
       * **取得工廠：** 根據 plugin_id 從 registry 中取得對應的工廠。  
       * **準備配置和依賴：**  
         * 從 composition.yaml 中提取組件自身的 config。  
         * 對於每個 dependency，從 components 映射中取出已初始化且類型斷言正確的依賴實例。**（參考 docs/architecture/interface_spec.md 來確認介面名稱）**  
         * 將這些配置和依賴組合成傳遞給 Create 方法的 map[string]any。  
       * **呼叫 Create 方法：** 生成呼叫 factory.Create(ctx, configAndDependencies) 的程式碼。  
       * **錯誤處理：** 包含基本的錯誤檢查。  
       * **儲存實例：** 將創建的實例儲存到 components 映射中，並進行正確的介面類型斷言（根據 docs/architecture/interface_spec.md）。  
   * **填充 {{.PluginAssemblyCode}} 佔位符：** 將生成的組裝邏輯插入到模板中 {{.PluginAssemblyCode}} 的位置。  
5. **填充其他通用程式碼：**  
   * **日誌系統初始化 ({{.LoggerInitializationCode}})：** 根據 composition.yaml 中定義的日誌配置，生成實際的日誌系統初始化程式碼。如果 logger 是第一個需要初始化的組件（通常是這樣），則需要特別處理其初始化，以確保後續日誌輸出可用。  
   * **配置加載程式碼 ({{.ConfigurationLoadingCode}})：** 根據 composition.yaml 中定義的配置提供者，生成平台配置加載程式碼。  
   * **平台服務啟動程式碼 ({{.PlatformServiceStartupCode}})：** 根據 composition.yaml 中定義的服務（例如 Web 服務、gRPC 服務），生成相應的服務啟動邏輯。  
   * **優雅關閉程式碼 ({{.GracefulShutdownCode}})：** 填充處理作業系統信號並執行清理工作的程式碼。  
6. **生成最終的 main.go 檔案：**  
   * 將所有佔位符替換為生成的程式碼片段。  
   * 將完成的內容作為 main.go 檔案輸出。

**注意事項：**

* **錯誤處理：** 生成的程式碼應包含基本的錯誤處理邏輯，例如檢查 NewFactory() 和 Create() 呼叫的錯誤。  
* **類型斷言：** 在組裝過程中，確保對插件實例進行正確的介面類型斷言，這需要 AI **明確參考 docs/architecture/interface_spec.md** 中定義的 Go 介面名稱。  
* **依賴順序：** 拓撲排序是組裝正確性的核心，必須正確實現。  
* **組件配置：** Create 方法需要接收 map[string]any 形式的配置，AI 應將 composition.yaml 中對應組件的 config 轉換為此格式。

