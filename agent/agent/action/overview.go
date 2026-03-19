package action

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

func GetStatus() (*model.HostStatus, error) {
	var status model.HostStatus

	// cpu
	percents, err := cpu.Percent(0, false)
	if err != nil {
		global.LOG.Error("failed to get cpu: %v", err)
	}
	if len(percents) > 0 {
		status.Cpu = math.Round(percents[0]*100) / 100
	}

	// mem
	v, err := mem.VirtualMemory()
	if err != nil {
		global.LOG.Error("failed to get mem: %v", err)
	}
	if v != nil {
		status.Memory = math.Round(v.UsedPercent*100) / 100
		status.MemTotal = utils.FormatMemorySize(v.Available + v.Used)
		status.MemUsed = utils.FormatMemorySize(v.Used)
	}

	// disk
	diskUsage, err := disk.Usage("/")
	if err != nil {
		global.LOG.Error("failed to get disk: %v", err)
	}
	if diskUsage != nil {
		status.Disk = math.Round(diskUsage.UsedPercent*100) / 100
	}

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	status.HeapAlloc = memStats.HeapAlloc
	status.HeapSys = memStats.HeapSys
	status.StackInuse = memStats.StackInuse
	status.Goroutines = runtime.NumGoroutine()
	status.OpenFDs = countOpenFDs()
	status.ProcessRSS = readProcessRSSBytes()

	if proc, err := process.NewProcess(int32(os.Getpid())); err == nil {
		if memInfo, memErr := proc.MemoryInfo(); memErr == nil && memInfo != nil && memInfo.RSS > 0 {
			status.ProcessRSS = memInfo.RSS
		}
	}

	return &status, nil
}

func readProcessRSSBytes() uint64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		kb, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr != nil {
			return 0
		}
		return kb * 1024
	}

	return 0
}

func countOpenFDs() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0
	}
	return len(entries)
}

func GetOverview() (*model.Overview, error) {
	var overview model.Overview

	servertime, err := shell.ExecuteCommand("date +\"%Y-%m-%d %H:%M:%S\"")
	if err != nil {
		servertime = ""
	}
	overview.ServerTime = strings.TrimSpace(servertime)

	serverTimezone, err := shell.ExecuteCommand("timedatectl | grep \"Time zone\" | awk '{print $3}'")
	if err != nil {
		serverTimezone = ""
	}
	overview.ServerTimeZone = strings.TrimSpace(serverTimezone)

	info, err := host.Info()
	if err != nil {
		global.LOG.Error("failed to get host info: %v", err)
	} else {
		overview.BootTime = utils.FormatTime(int64(info.BootTime))
		overview.RunTime = int64(info.Uptime)
	}

	timesList, err := cpu.Times(true) // true 表示返回每个CPU核心的统计
	if err == nil && len(timesList) > 0 {
		var totalIdle, totalTotal float64
		for _, t := range timesList {
			totalIdle += t.Idle
			totalTotal += t.User + t.System + t.Idle + t.Nice +
				t.Iowait + t.Irq + t.Softirq + t.Steal
		}

		// 平均每核空闲时间
		avgIdle := totalIdle / float64(len(timesList))

		// 平均空闲率
		idleRate := (totalIdle / totalTotal) * 100

		overview.IdleTime = int64(math.Round(avgIdle))
		overview.IdleRate = math.Round(idleRate*100) / 100
	}

	percents, err := cpu.Percent(0, false)
	if err != nil {
		global.LOG.Error("failed to get cpu percent: %v", err)
	} else if len(percents) > 0 {
		overview.CpuUsage = fmt.Sprintf("%.2f%%", percents[0])
	}

	var logicalCPU = 1
	cpuCnts, err := cpu.Counts(true) // 逻辑 CPU 数（线程数）
	if err != nil {
		global.LOG.Error("failed to get cpu counts: %v", err)
	} else {
		logicalCPU = cpuCnts
	}
	avg, err := load.Avg()
	if err != nil {
		global.LOG.Error("failed to get load avg: %v", err)
		avg = &load.AvgStat{} // 全 0 避免 panic
	}
	calcPercent := func(loadVal float64) float64 {
		if logicalCPU == 0 {
			return 0
		}
		return (loadVal / float64(logicalCPU)) * 100
	}
	overview.CurrentLoad.ProcessCount1 = fmt.Sprintf("%.2f%%", calcPercent(avg.Load1))
	overview.CurrentLoad.ProcessCount5 = fmt.Sprintf("%.2f%%", calcPercent(avg.Load5))
	overview.CurrentLoad.ProcessCount15 = fmt.Sprintf("%.2f%%", calcPercent(avg.Load15))

	v, err := mem.VirtualMemory()
	if err != nil {
		global.LOG.Error("failed to get virtual memory: %v", err)
	} else if v != nil {
		kernelUsed := v.Buffers + v.Slab
		totalAvailable := uint64(0)
		if v.Total > kernelUsed {
			totalAvailable = v.Total - kernelUsed
		}

		realUsed := uint64(0)
		if v.Used > (v.Buffers + v.Cached) {
			realUsed = v.Used - v.Buffers - v.Cached
		}
		usedInTotal := uint64(0)
		if totalAvailable > v.Available {
			usedInTotal = totalAvailable - v.Available
		}

		usedRate := 0.0
		freeRate := 0.0
		if totalAvailable > 0 {
			usedRate = (float64(usedInTotal) / float64(totalAvailable)) * 100
			freeRate = (float64(v.Available) / float64(totalAvailable)) * 100
		}

		overview.MemoryUsage.Physical = utils.FormatMemorySize(v.Total)
		overview.MemoryUsage.Kernel = utils.FormatMemorySize(kernelUsed)
		overview.MemoryUsage.Total = utils.FormatMemorySize(totalAvailable)
		overview.MemoryUsage.Used = utils.FormatMemorySize(usedInTotal)
		overview.MemoryUsage.UsedRate = math.Round(usedRate*100) / 100
		overview.MemoryUsage.Free = utils.FormatMemorySize(v.Available)
		overview.MemoryUsage.FreeRate = math.Round(freeRate*100) / 100
		overview.MemoryUsage.Buffered = utils.FormatMemorySize(v.Buffers)
		overview.MemoryUsage.Cached = utils.FormatMemorySize(v.Cached)
		overview.MemoryUsage.RealUsed = utils.FormatMemorySize(realUsed)
	}

	swap, err := mem.SwapMemory()
	if err != nil {
		global.LOG.Error("failed to get swap memory: %v", err)
	} else if swap != nil {
		overview.SwapUsage.Total = utils.FormatMemorySize(swap.Total)
		overview.SwapUsage.Free = utils.FormatMemorySize(swap.Free)
		overview.SwapUsage.Used = utils.FormatMemorySize(swap.Used)
		overview.SwapUsage.UsedRate = math.Round(swap.UsedPercent*100) / 100
		overview.SwapUsage.FreeRate = math.Round((100-swap.UsedPercent)*100) / 100
	}

	partitions, err := disk.Partitions(false)
	if err != nil {
		global.LOG.Error("failed to get disk partitions: %v", err)
	} else {
		for _, part := range partitions {
			usage, usageErr := disk.Usage(part.Mountpoint)
			if usageErr != nil {
				global.LOG.Error("failed to get disk usage for %s: %v", part.Mountpoint, usageErr)
				continue
			}
			overview.Storage = append(
				overview.Storage,
				model.Partition{
					Name:     part.Mountpoint,
					Total:    utils.FormatMemorySize(usage.Total),
					Free:     utils.FormatMemorySize(usage.Free),
					Used:     utils.FormatMemorySize(usage.Used),
					UsedRate: math.Round(usage.UsedPercent*100) / 100,
				},
			)
		}
	}
	return &overview, nil
}
