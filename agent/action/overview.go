package action

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

func GetOverview() (string, error) {
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
	overview.RunTime = utils.FormatDuration(int64(info.Uptime))

	times, _ := cpu.Times(false)
	if len(times) > 0 {
		t := times[0]
		idleSeconds := int64(t.Idle)
		overview.IdleTime = utils.FormatDuration(idleSeconds)
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

	jsonString, err := json.Marshal(overview)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(jsonString), nil
}
