package action

import (
	"fmt"

	"github.com/sensdata/idb/core/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// 获取硬件信息
func GetHardware() (*model.HardwareInfo, error) {
	var hardware model.HardwareInfo

	// 获取CPU的物理核心数量
	cpuCount, err := cpu.Counts(true) // true 表示获取逻辑CPU核心数
	if err != nil {
		return nil, err
	}
	hardware.CpuCount = cpuCount

	// 获取每个CPU的核心数
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	if len(cpuInfo) > 0 {
		hardware.CpuCores = int(cpuInfo[0].Cores)
	}

	// 获取线程数（每个CPU的线程数）
	processorCount, err := cpu.Counts(false) // false 表示获取物理CPU核心数
	if err != nil {
		return nil, err
	}
	hardware.Processor = processorCount

	// 获取每个CPU的型号信息
	var moduleNames []string
	for _, info := range cpuInfo {
		moduleNames = append(moduleNames, info.ModelName)
	}
	hardware.ModuleNames = moduleNames

	// 获取内存大小
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	hardware.Memory = formatMemorySize(vmStat.Total) // 格式化内存大小为可读格式

	return &hardware, nil
}

// 格式化内存大小为人类可读的格式（例如：1 GB）
func formatMemorySize(size uint64) string {
	var units = []string{"B", "KB", "MB", "GB", "TB"}
	var unitIndex = 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}
	// 去掉小数点，只显示整数
	return fmt.Sprintf("%d %s", size, units[unitIndex])
}
