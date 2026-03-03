package model

// 硬件信息
type HardwareInfo struct {
	CpuCount    int            `json:"cpu_count"`      //cpu个数
	CpuCores    int            `json:"cpu_cores"`      //cpu核心数
	Processor   int            `json:"processor"`      //线程数
	ModuleNames []string       `json:"module_names"`   //兼容字段: CPU型号列表
	CpuModels   []CpuModelInfo `json:"cpu_models"`     //CPU型号明细
	Memory      string         `json:"memory"`         //内存大小
	MemorySlots int            `json:"memory_slots"`   //内存条数量
	MemoryMods  []MemoryModule `json:"memory_modules"` //内存条明细
	DiskCount   int            `json:"disk_count"`     //磁盘数量
	Disks       []DiskInfo     `json:"disks"`          //磁盘明细
}

type CpuModelInfo struct {
	Model string `json:"model"` //型号
	Count int    `json:"count"` //逻辑处理器数量
}

type MemoryModule struct {
	Locator      string `json:"locator"`      //插槽
	Size         string `json:"size"`         //容量
	Type         string `json:"type"`         //类型
	Speed        string `json:"speed"`        //速率
	Manufacturer string `json:"manufacturer"` //厂商
	PartNumber   string `json:"part_number"`  //型号
}

type DiskInfo struct {
	Name               string `json:"name"`                //设备名
	Model              string `json:"model"`               //型号
	Size               string `json:"size"`                //容量
	Type               string `json:"type"`                //类型
	Health             string `json:"health"`              //健康状态
	LifeUsed           string `json:"life_used"`           //寿命消耗
	PowerOnHours       string `json:"power_on_hours"`      //通电时长(小时)
	PowerCycleCount    string `json:"power_cycle_count"`   //上电次数
	Temperature        string `json:"temperature"`         //温度
	AvailableSpare     string `json:"available_spare"`     //可用备用空间
	ReallocatedSectors string `json:"reallocated_sectors"` //重映射扇区
	PendingSectors     string `json:"pending_sectors"`     //待处理扇区
}
