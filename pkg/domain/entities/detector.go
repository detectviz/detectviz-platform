package entities

import "time"

// Detector 是 Detectviz 平台的核心偵測器實體。
// 職責: 封裝偵測器的配置、狀態及與偵測器相關的業務行為。
type Detector struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
