# **CONTRIBUTING.md - 貢獻指南**

我們熱烈歡迎所有對 Detectviz 平台感興趣的貢獻者。您的貢獻對於平台的成長和成功至關重要！本指南將幫助您了解如何為 Detectviz 平台貢獻程式碼、報告問題或提出功能建議。

## **1. 行為準則 (Code of Conduct)**

我們致力於維護一個開放、包容和友好的社區環境。所有貢獻者都應遵守我們的行為準則。請尊重他人，避免騷擾行為。

## **2. 如何貢獻？**

### **2.1 報告 Bug (Bug Reports)**

如果您發現了任何 Bug，請透過 GitHub Issue 追蹤器提交。在提交 Bug 報告時，請盡可能提供詳細的資訊，包括：

* **問題描述**：清晰簡潔地描述您遇到的問題。  
* **重現步驟**：詳細說明如何重現這個 Bug。  
* **預期行為**：描述您認為程式應該如何運作。  
* **實際行為**：描述程式實際的運作方式。  
* **環境資訊**：例如 Go 版本、作業系統、相關配置等。  
* **日誌輸出**：如果可以，請附上相關的日誌片段。

### **2.2 提出功能請求 (Feature Requests)**

如果您有新的功能想法或改進建議，也請透過 GitHub Issue 追蹤器提交。請詳細說明您的想法，包括：

* **功能描述**：清晰地描述您希望新增的功能或改進。  
* **問題背景**：說明這個功能解決了什麼問題或帶來了什麼價值。  
* **使用場景**：提供具體的使用場景範例。  
* **潛在影響**：考慮對現有系統可能造成的影響。

### **2.3 提交程式碼 (Pull Requests)**

我們鼓勵您透過 Pull Request (PR) 的方式貢獻程式碼。在提交 PR 之前，請確保您已閱讀並理解以下指南。

## **3. 開發環境設置 (Development Setup)**

在開始編寫程式碼之前，請確保您的開發環境已正確設置。詳細的設置步驟請參考 [DEVELOPMENT.md](http://docs.google.com/docs/DEVELOPMENT.md) 文件。

## **4. 程式碼風格與規範 (Code Style & Guidelines)**

為了保持程式碼庫的一致性和可維護性，所有貢獻都必須遵循 Detectviz 平台的程式碼風格和工程規範。請務必查閱 [ENGINEERING_SPEC.md](http://docs.google.com/docs/ENGINEERING_SPEC.md) 文件，其中包含了：

* Go 程式碼風格指南  
* 專案結構與檔案命名慣例  
* 錯誤處理原則  
* 日誌規範  
* 依賴注入與插件開發規範  
* 文檔標準與註解 (GoDoc)

**AI 協同提示**：AI 在生成程式碼時會嚴格遵循 ENGINEERING_SPEC.md 中的規範。在您審閱 AI 生成的程式碼時，請也以此文件為基準。

## **5. 測試 (Testing)**

所有程式碼貢獻都應包含相應的測試，以確保功能的正確性和穩定性。Detectviz 平台採用以下測試策略（詳細請參閱 [ENGINEERING_SPEC.md](http://docs.google.com/docs/ENGINEERING_SPEC.md) 的測試規範）：

* **單元測試 (Unit Tests)**：針對獨立的函數和方法。  
* **集成測試 (Integration Tests)**：測試多個模組或服務之間的協同工作。  
* **端到端測試 (End-to-End Tests)**：模擬真實用戶場景，驗證整個系統。

在提交 PR 之前，請確保所有測試都已通過。

**AI 協同提示**：AI 可以根據程式碼和 GoDoc 註解自動生成單元測試的骨架和初步測試用例。請利用此能力來加速測試的編寫。

## **6. 提交訊息規範 (Commit Message Guidelines)**

清晰的提交訊息對於程式碼審查和專案歷史追溯至關重要。我們建議使用 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 規範，例如：

* feat: add new user authentication module  
* fix: resolve database connection issue  
* docs: update contributing guide  
* refactor: improve error handling in config provider

## **7. Pull Request 流程**

1. **Fork 專案**：將 detectviz-platform 儲存庫 Fork 到您的 GitHub 帳戶。  
2. **克隆 Fork 的儲存庫**：  
   git clone https://github.com/您的用戶名/detectviz-platform.git  
   cd detectviz-platform

3. **創建新分支**：為您的貢獻創建一個新的功能分支。  
   git checkout -b feature/your-feature-name

4. **編寫程式碼**：實現您的功能或修復 Bug，並編寫相應的測試。  
5. **運行測試**：在提交之前，請確保所有測試都已通過。  
   go test ./...

6. **提交更改**：使用清晰的提交訊息。  
   git add .  
   git commit -m "feat: your descriptive commit message"

7. **推送到您的 Fork**：  
   git push origin feature/your-feature-name

8. **創建 Pull Request**：  
   * 前往您的 Fork 儲存庫頁面，點擊 "New pull request" 按鈕。  
   * 確保目標分支是 main 或 develop (根據專案實際主分支而定)。  
   * 提供清晰的 PR 描述，解釋您的更改內容、解決的問題以及如何測試。  
   * 如果您的 PR 關聯了任何 Issue，請在描述中引用它們 (例如：Closes #123)。

## **8. 程式碼審查 (Code Review)**

提交 PR 後，專案維護者將會對您的程式碼進行審查。請耐心等待並積極回應審查意見。我們可能會要求您進行修改，以確保程式碼質量和符合專案規範。

感謝您的貢獻！

