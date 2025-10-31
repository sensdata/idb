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

	cpuInfos, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	physicalCPUSet := make(map[string]struct{})
	coreSet := make(map[string]struct{})

	for _, ci := range cpuInfos {
		// 统计物理CPU颗数
		physicalCPUSet[ci.PhysicalID] = struct{}{}

		// 统计核心数 (PhysicalID + CoreID 唯一标识一个核心)
		coreKey := fmt.Sprintf("%s-%s", ci.PhysicalID, ci.CoreID)
		coreSet[coreKey] = struct{}{}
	}

	hardware.CpuCount = len(physicalCPUSet)  // 物理 CPU 颗数
	hardware.CpuCores = len(coreSet)         // 总核心数
	hardware.Processor, _ = cpu.Counts(true) // 逻辑 CPU（线程数）

	// 型号
	for _, ci := range cpuInfos {
		hardware.ModuleNames = append(hardware.ModuleNames, ci.ModelName)
	}

	// 内存大小
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	hardware.Memory = formatMemorySize(vmStat.Total)

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
