package model

// 系统信息
type SystemInfo struct {
	HostName string `json:"host_name"` //主机名称
	Version  string `json:"version"`   //发行版本
	Kernel   string `json:"kernel"`    //内核版本
	Platform string `json:"platform"`  //系统类型 x86_64, arm64, etc.
	Vertual  string `json:"vertual"`   //虚拟化平台
}
