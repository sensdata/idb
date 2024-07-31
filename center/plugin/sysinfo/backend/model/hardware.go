package model

// 硬件信息
type HardwareInfo struct {
	CpuCount    int      `json:"cpuCount"`    //cpu个数
	CpuCores    int      `json:"cpuCores"`    //cpu核心数
	Processor   int      `json:"processor"`   //线程数
	ModuleNames []string `json:"moduleNames"` //型号
	Memory      string   `json:"memory"`      //内存大小
}
