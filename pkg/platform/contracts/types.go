package contracts

// ServiceInstance 代表一個服務的實例。
// 職責: 提供服務實例的唯一標識、地址和元資訊。
type ServiceInstance struct {
	ID       string
	Address  string
	Port     int
	Metadata map[string]string
}
