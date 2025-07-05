package entities

// DetectionResult 是表示一個偵測事件處理後的最終結果的領域值物件。
// 職責: 封裝偵測事件被處理後的輸出，包括是否產生了分析結果以及處理狀態。
// 它是一個不可變的對象，代表了對一個 Detection 的最終裁定。
type DetectionResult struct {
	// DetectionID 關聯原始的偵測事件ID。
	DetectionID string
	// Status 表示處理結果，例如 "processed", "ignored", "failed"。
	Status string
	// AnalysisResultID 如果生成了分析結果，則記錄其ID。
	AnalysisResultID string
	// Message 提供了關於處理結果的額外信息，例如忽略原因或錯誤詳情。
	Message string
}
