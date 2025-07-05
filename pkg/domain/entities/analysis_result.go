package entities

import "time"

// AnalysisResult 是一個表示數據分析結果的領域實體。
// 職責: 捕獲並封裝分析過程產生的結構化結果，例如偵測到的異常、趨勢、或洞察。
// 它是一個不可變的記錄，代表某次分析的快照。
type AnalysisResult struct {
	// ID 是分析結果的唯一標識符。
	ID string
	// DetectorID 關聯觸發此次分析的偵測器ID。
	DetectorID string
	// Timestamp 記錄分析發生的時間。
	Timestamp time.Time
	// Summary 是對分析結果的簡要總結。
	Summary string
	// Data 包含了詳細的、結構化的分析數據。
	Data map[string]interface{}
	// Severity 表示分析結果的重要性或嚴重程度。
	Severity string
}
