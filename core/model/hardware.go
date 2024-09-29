package model

// 硬件信息
type HardwareInfo struct {
	CpuCount    int      `json:"cpu_count"`    //cpu个数
	CpuCores    int      `json:"cpu_cores"`    //cpu核心数
	Processor   int      `json:"processor"`    //线程数
	ModuleNames []string `json:"module_names"` //型号
	Memory      string   `json:"memory"`       //内存大小
}
