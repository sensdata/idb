package action

import (
	"fmt"
	"math"
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

	return &status, nil
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

	info, _ := host.Info()
	overview.BootTime = utils.FormatTime(int64(info.BootTime))
	overview.RunTime = int64(info.Uptime)

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

	percents, _ := cpu.Percent(0, false)
	if len(percents) > 0 {
		overview.CpuUsage = fmt.Sprintf("%.2f%%", percents[0])
	}

	avg, _ := load.Avg()
	overview.CurrentLoad.ProcessCount1 = fmt.Sprintf("%.2f%%", avg.Load1*100)
	overview.CurrentLoad.ProcessCount5 = fmt.Sprintf("%.2f%%", avg.Load5*100)
	overview.CurrentLoad.ProcessCount15 = fmt.Sprintf("%.2f%%", avg.Load15*100)

	v, _ := mem.VirtualMemory()
	overview.MemoryUsage.Physical = utils.FormatMemorySize(v.Total)
	overview.MemoryUsage.Kernel = utils.FormatMemorySize(v.Buffers + v.Slab)
	overview.MemoryUsage.Total = utils.FormatMemorySize(v.Available + v.Used)
	overview.MemoryUsage.Used = utils.FormatMemorySize(v.Used)
	overview.MemoryUsage.UsedRate = math.Round(v.UsedPercent*100) / 100
	overview.MemoryUsage.Free = utils.FormatMemorySize(v.Available)
	overview.MemoryUsage.FreeRate = 100 - overview.MemoryUsage.UsedRate
	overview.MemoryUsage.Buffered = utils.FormatMemorySize(v.Buffers)
	overview.MemoryUsage.Cached = utils.FormatMemorySize(v.Cached)
	overview.MemoryUsage.RealUsed = utils.FormatMemorySize(v.Used)

	swap, _ := mem.SwapMemory()
	overview.SwapUsage.Total = utils.FormatMemorySize(swap.Total)
	overview.SwapUsage.Free = utils.FormatMemorySize(swap.Free)
	overview.SwapUsage.Used = utils.FormatMemorySize(swap.Used)
	overview.SwapUsage.UsedRate = math.Round(swap.UsedPercent*100) / 100
	overview.SwapUsage.FreeRate = math.Round((100-swap.UsedPercent)*100) / 100

	partitions, _ := disk.Partitions(false)
	for _, part := range partitions {
		usage, _ := disk.Usage(part.Mountpoint)
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
	return &overview, nil
}
