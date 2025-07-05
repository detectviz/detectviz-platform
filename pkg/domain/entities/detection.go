package entities

import "time"

// Detection 是表示一個特定偵測事件的領域實體。
// 職責: 封裝偵測事件的上下文，例如觸發時間、來源數據等。
type Detection struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DetectorID  string                 `json:"detector_id"`
	Description string                 `json:"description"`
}
