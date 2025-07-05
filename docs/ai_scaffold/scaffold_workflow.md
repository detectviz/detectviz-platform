# **Detectviz AI 腳手架工作流程 (AI Scaffolding Workflow)**

本文件詳細定義了 Detectviz 平台中 AI 輔助腳手架的標準工作流程、核心原則與強制性規範，旨在確保 AI 在生成、驗證和組裝程式碼時，能夠與平台架構和設計意圖高度一致。

## **1. 核心設計理念：AI 驅動的契約優先 (AI-Driven, Contract-First)**

Detectviz 平台的核心設計哲學是「一切皆插件 (Everything is a Plugin)」，並深度整合 AI 進行自動化。這要求所有介面（契約）和實現都必須是 AI 友好的。

* **契約優先 (Contract-First)**：在任何程式碼實作之前，必須先在 interface_spec.md 中定義清晰的 Go 介面。這些介面是 AI 理解功能和生成程式碼的基礎。  
* **AI 友好設計 (AI-Friendly Design)**：程式碼結構、命名規範和文件註解（特別是 AI 標籤）都必須能夠被 AI 高效解析和理解。  
* **自動化與驗證 (Automation & Validation)**：AI 生成的程式碼必須能夠通過自動化測試和 JSON Schema 驗證，確保其正確性和合規性。

## **2. AI 腳手架核心工作流程**

Detectviz AI 腳手架將遵循以下步驟來輔助開發者：

1. **需求解析與介面識別**：  
   * AI 接收開發者需求（例如：「創建一個新的日誌供應商插件」）。  
   * AI 分析需求，並在 interface_spec.md 中識別或建議需要實現的核心介面（例如 contracts.Logger）。  
2. **基於 AI 標籤生成插件骨架 (強制性)**：  
   * **AI 必須嚴格遵循 interface_spec.md 中定義的 AI 標籤**來生成插件的初始骨架。這些標籤是 AI 識別插件類型、預期實現路徑和構造函數的**強制性指令**。  
   * **示例 AI 標籤及其強制性要求**：  
     // interface_spec.md 中的示例  
     // LLMProvider 定義了大型語言模型推論功能的通用介面。  
     // AI_PLUGIN_TYPE: "llm_provider" // 識別此介面對應的插件類型  
     // AI_IMPL_PACKAGE: "detectviz-platform/internal/platform/providers/llm/gemini_llm" // 強制性：指定預期的實現包路徑  
     // AI_IMPL_CONSTRUCTOR: "NewGeminiLLMProvider" // 強制性：指定預期的構造函數名稱  
     type LLMProvider interface {  
         GenerateText(ctx context.Context, prompt string, options map[string]any) (string, error)  
         GetName() string  
     }

   * **AI 生成的插件骨架將嚴格遵守 AI_IMPL_PACKAGE 指定的路徑和 AI_IMPL_CONSTRUCTOR 指定的構造函數名稱。** 這確保了平台組裝時的一致性。  
3. **自動生成插件工廠 (NewFactory()) (強制性)**：  
   * **每個插件的實現包 (例如 internal/platform/providers/llm/gemini_llm) 都必須提供一個 NewFactory() 函數。** 這是平台動態組裝插件的標準入口。  
   * **AI 在生成插件骨架時，應自動包含此 NewFactory() 函數。**  
   * NewFactory() 的職責：  
     * 接收 contracts.ConfigProvider 實例（或其他必要的基礎依賴）。  
     * 返回一個能夠創建該插件實例的函數（通常是 func(ctx context.Context) (contracts.Plugin, error) 或類似簽名）。  
     * 該函數內部應包含插件實例化的邏輯，包括讀取配置、初始化內部組件等。  
   * **示例 NewFactory() 簽名**：  
     // internal/platform/providers/llm/gemini_llm/factory.go (或 main.go)  
     package gemini_llm

     import (  
         "context"  
         "detectviz-platform/pkg/platform/contracts"  
     )

     // NewFactory 創建並返回一個用於構建 GeminiLLMProvider 實例的工廠函數。  
     // 這個工廠函數將被 PluginRegistry 用於動態組裝。  
     func NewFactory() func(ctx context.Context, configProvider contracts.ConfigProvider) (contracts.LLMProvider, error) {  
         return func(ctx context.Context, configProvider contracts.ConfigProvider) (contracts.LLMProvider, error) {  
             // 在這裡從 configProvider 讀取 Gemini LLM 的配置  
             // 例如：cfg, err := configProvider.LoadPluginConfig("gemini_llm_config")  
             // ... 實際的配置讀取和解析邏輯 ...

             // 實例化 Gemini LLM Provider  
             provider := &GeminiLLMProvider{  
                 // ... 根據配置初始化字段 ...  
             }  
             return provider, nil  
         }  
     }

4. **JSON Schema 的自動推斷與驗證 (強制性)**：  
   * **AI 應具備從 Go 介面和其 Config 結構體定義中，推斷並生成對應 JSON Schema 的能力。** 這將確保配置的結構與程式碼的期望保持同步。  
   * **AI 在生成新插件時，應同時生成其配置的 JSON Schema 骨架**，並提示開發者根據詳細業務邏輯填充更精確的驗證規則（例如 pattern, minimum, maximum 等）。  
   * **所有由 AI 生成或輔助生成的配置，都必須通過其對應的 JSON Schema 驗證。**  
   * **驗證時機**：  
     * **開發時**：AI 在生成配置時進行即時驗證。  
     * **程式碼提交前**：通過預提交鉤子 (pre-commit hook) 或 CI/CD 流程中的靜態分析工具自動執行。  
     * **應用啟動時**：ConfigProvider 在載入應用程式配置和插件配置時，必須使用 JSON Schema 進行嚴格驗證，不符合 Schema 的配置應導致啟動失敗並提供清晰的錯誤訊息。  
5. **組裝邏輯的自動生成與更新 (AI 協同點)**：  
   * 基於 composition.yaml（定義了哪些插件被啟用及其配置）和所有插件的 NewFactory() 函數，AI 將能夠**自動生成或更新** cmd/api/main.go 或 internal/bootstrap/platform_initializer.go 中的組裝邏輯。  
   * 這包括處理插件之間的依賴關係（例如，DetectorService 可能依賴 LLMProvider），AI 應能進行基本的**拓撲排序**來確保依賴被正確注入。

## **3. 開發者與 AI 的協同準則**

* **開發者職責**：  
  * 定義清晰、符合規範的 Go 介面 (interface_spec.md)。  
  * 為 Config 結構體編寫詳細且 AI 友好的 GoDoc 註解。  
  * 完善 AI 生成的 JSON Schema 骨架，添加精確的驗證規則。  
  * 編寫插件的業務邏輯實現。  
* **AI 職責**：  
  * 依據介面定義和 AI 標籤，生成符合規範的插件骨架。  
  * 自動生成 NewFactory() 函數。  
  * 從 Go 結構體推斷並生成 JSON Schema 骨架。  
  * 根據 composition.yaml 和插件工廠，自動生成或更新平台組裝邏輯。  
  * 執行配置的 JSON Schema 驗證。

## **4. 持續集成與部署 (CI/CD) 中的腳手架驗證**

Detectviz 平台的 CI/CD 流程將包含以下與 AI 腳手架相關的強制性驗證步驟：

* **介面與實現一致性檢查**：自動檢查插件的包路徑和構造函數名稱是否與 interface_spec.md 中的 AI 標籤保持一致。  
* **JSON Schema 驗證**：所有配置檔案在提交前和應用啟動時，必須通過其對應 JSON Schema 的驗證。  
* **工廠函數存在性檢查**：自動檢查所有插件實現包是否都提供了符合規範的 NewFactory() 函數。

透過這些規範的加強，Detectviz 平台將能更有效地利用 AI 進行自動化，顯著提升開發效率和系統的一致性。