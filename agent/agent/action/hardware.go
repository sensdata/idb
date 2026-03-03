package action

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// 获取硬件信息
func GetHardware() (*model.HardwareInfo, error) {
	var hardware model.HardwareInfo

	cpuInfos, err := cpu.Info()
	if err != nil {
		global.LOG.Error("failed to get cpu info: %v", err)
	}

	physicalCPUSet := make(map[string]struct{})
	coreSet := make(map[string]struct{})
	modelCounter := make(map[string]int)
	modelDisplayName := make(map[string]string)

	physicalCores, err := cpu.Counts(false)
	if err != nil {
		global.LOG.Error("failed to get physical cpu cores: %v", err)
	}

	logicalProcessors, err := cpu.Counts(true)
	if err != nil {
		global.LOG.Error("failed to get logical cpu count: %v", err)
	}

	for _, ci := range cpuInfos {
		physicalID := strings.TrimSpace(ci.PhysicalID)
		coreID := strings.TrimSpace(ci.CoreID)
		modelName := normalizeCPUModelName(ci.ModelName)

		// PhysicalID 在容器/云主机中可能缺失，缺失时不参与 cpu_count 统计
		if physicalID != "" {
			physicalCPUSet[physicalID] = struct{}{}
		}

		// 统计核心数 (PhysicalID + CoreID 唯一标识一个核心)，PhysicalID 缺失时退化为 CoreID
		switch {
		case physicalID != "" && coreID != "":
			coreSet[physicalID+"-"+coreID] = struct{}{}
		case coreID != "":
			coreSet[coreID] = struct{}{}
		}

		if modelName == "" {
			continue
		}
		modelCounter[modelName]++
		if modelDisplayName[modelName] == "" {
			modelDisplayName[modelName] = modelName
		}
	}

	hardware.CpuCount = len(physicalCPUSet) // 物理 CPU 颗数（socket）
	hardware.CpuCores = physicalCores       // 总物理核心数
	hardware.Processor = logicalProcessors  // 逻辑 CPU（线程数）

	// 兜底：在部分环境下 cpu.Counts(false) 返回 0
	if hardware.CpuCores <= 0 && len(coreSet) > 0 {
		hardware.CpuCores = len(coreSet)
	}
	if hardware.CpuCount <= 0 && hardware.CpuCores > 0 {
		// 缺失 PhysicalID 时无法精确得出 socket 数，退化为 1
		hardware.CpuCount = 1
	}
	if hardware.Processor <= 0 && hardware.CpuCores > 0 {
		hardware.Processor = hardware.CpuCores
	}

	hardware.CpuModels = buildCpuModels(modelCounter, modelDisplayName)
	for _, item := range hardware.CpuModels {
		if item.Count > 1 {
			hardware.ModuleNames = append(
				hardware.ModuleNames,
				item.Model+" x"+strconv.Itoa(item.Count),
			)
		} else {
			hardware.ModuleNames = append(hardware.ModuleNames, item.Model)
		}
	}

	// 内存大小
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		global.LOG.Error("failed to get memory info: %v", err)
	} else if vmStat != nil {
		hardware.Memory = utils.FormatMemorySize(vmStat.Total)
	}

	if hardware.Memory == "" {
		hardware.Memory = "-"
	}

	hardware.MemoryMods = collectMemoryModules()
	hardware.MemorySlots = len(hardware.MemoryMods)
	hardware.Disks = collectDiskInfo()
	hardware.DiskCount = len(hardware.Disks)
	updatedAt := latestModTime(
		"/proc/cpuinfo",
		"/proc/meminfo",
		"/proc/partitions",
		"/sys/devices/virtual/dmi/id/product_name",
		"/sys/devices/virtual/dmi/id/product_uuid",
	)
	if !updatedAt.IsZero() {
		hardware.UpdatedAt = updatedAt
	}

	return &hardware, nil
}

func latestModTime(paths ...string) time.Time {
	latest := time.Time{}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if mod := info.ModTime(); mod.After(latest) {
			latest = mod
		}
	}
	return latest
}

func buildCpuModels(counter map[string]int, displayName map[string]string) []model.CpuModelInfo {
	cpuModels := make([]model.CpuModelInfo, 0, len(counter))
	for normalized, count := range counter {
		modelName := displayName[normalized]
		if modelName == "" {
			modelName = normalized
		}
		cpuModels = append(cpuModels, model.CpuModelInfo{
			Model: modelName,
			Count: count,
		})
	}

	sort.Slice(cpuModels, func(i, j int) bool {
		if cpuModels[i].Count != cpuModels[j].Count {
			return cpuModels[i].Count > cpuModels[j].Count
		}
		return cpuModels[i].Model < cpuModels[j].Model
	})

	return cpuModels
}

var (
	cpuSpacesRE = regexp.MustCompile(`\s+`)
	cpuAtFreqRE = regexp.MustCompile(`\s*@\s*\d+(\.\d+)?\s*[GM]Hz\b`)
	digitsRE    = regexp.MustCompile(`\d+`)
)

func normalizeCPUModelName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return ""
	}

	// unify spaces/tabs
	name = cpuSpacesRE.ReplaceAllString(name, " ")
	// remove common trailing frequency marker
	name = cpuAtFreqRE.ReplaceAllString(name, "")
	// normalize redundant vendor tokens
	name = strings.ReplaceAll(name, "(R)", "")
	name = strings.ReplaceAll(name, "(TM)", "")
	name = strings.ReplaceAll(name, " CPU", "")
	name = cpuSpacesRE.ReplaceAllString(strings.TrimSpace(name), " ")

	return name
}

func collectMemoryModules() []model.MemoryModule {
	output, err := shell.ExecuteCommand("dmidecode -t memory 2>/dev/null")
	if err != nil {
		return nil
	}

	lines := strings.Split(output, "\n")
	mods := make([]model.MemoryModule, 0)
	current := make(map[string]string)
	inDevice := false

	flush := func() {
		if len(current) == 0 {
			return
		}

		size := strings.TrimSpace(current["Size"])
		if size == "" || size == "No Module Installed" || size == "Unknown" {
			current = make(map[string]string)
			return
		}

		mod := model.MemoryModule{
			Locator:      strings.TrimSpace(current["Locator"]),
			Size:         size,
			Type:         strings.TrimSpace(current["Type"]),
			Speed:        strings.TrimSpace(current["Speed"]),
			Manufacturer: strings.TrimSpace(current["Manufacturer"]),
			PartNumber:   strings.TrimSpace(current["Part Number"]),
		}

		if mod.Locator == "" {
			mod.Locator = "-"
		}
		if mod.Type == "" {
			mod.Type = "-"
		}
		if mod.Speed == "" {
			mod.Speed = "-"
		}
		if mod.Manufacturer == "" {
			mod.Manufacturer = "-"
		}
		if mod.PartNumber == "" {
			mod.PartNumber = "-"
		}

		mods = append(mods, mod)
		current = make(map[string]string)
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Memory Device" {
			flush()
			inDevice = true
			continue
		}
		if !inDevice || trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		current[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	flush()

	return mods
}

func collectDiskInfo() []model.DiskInfo {
	output, err := shell.ExecuteCommand("lsblk -J -b -d -o NAME,SIZE,MODEL,TYPE,ROTA 2>/dev/null")
	if err != nil {
		return nil
	}

	type lsblkDevice struct {
		Name  string `json:"name"`
		Size  string `json:"size"`
		Model string `json:"model"`
		Type  string `json:"type"`
		Rota  bool   `json:"rota"`
	}
	var payload struct {
		Blockdevices []lsblkDevice `json:"blockdevices"`
	}

	if err = json.Unmarshal([]byte(output), &payload); err != nil {
		return nil
	}

	disks := make([]model.DiskInfo, 0, len(payload.Blockdevices))
	for _, dev := range payload.Blockdevices {
		if dev.Name == "" {
			continue
		}
		sizeUint, sizeErr := strconv.ParseUint(strings.TrimSpace(dev.Size), 10, 64)
		sizeText := "-"
		if sizeErr == nil {
			sizeText = utils.FormatMemorySize(sizeUint)
		}

		modelText := strings.TrimSpace(dev.Model)
		if modelText == "" {
			modelText = "-"
		}

		diskType := strings.TrimSpace(dev.Type)
		if strings.HasPrefix(dev.Name, "nvme") {
			diskType = "nvme"
		} else if diskType == "disk" {
			if dev.Rota {
				diskType = "hdd"
			} else {
				diskType = "ssd"
			}
		}
		if diskType == "" {
			diskType = "-"
		}

		disk := model.DiskInfo{
			Name:               "/dev/" + dev.Name,
			Model:              modelText,
			Size:               sizeText,
			Type:               diskType,
			Health:             "-",
			LifeUsed:           "-",
			PowerOnHours:       "-",
			PowerCycleCount:    "-",
			Temperature:        "-",
			AvailableSpare:     "-",
			ReallocatedSectors: "-",
			PendingSectors:     "-",
		}

		fillDiskHealth(&disk)
		disks = append(disks, disk)
	}

	return disks
}

type smartCtlPayload struct {
	SmartStatus struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	PowerOnTime struct {
		Hours float64 `json:"hours"`
	} `json:"power_on_time"`
	PowerCycleCount int `json:"power_cycle_count"`
	Temperature     struct {
		Current float64 `json:"current"`
	} `json:"temperature"`
	NvmeLog struct {
		PercentageUsed int     `json:"percentage_used"`
		AvailableSpare int     `json:"available_spare"`
		Temperature    float64 `json:"temperature"`
		PowerOnHours   int     `json:"power_on_hours"`
		PowerCycles    int     `json:"power_cycles"`
	} `json:"nvme_smart_health_information_log"`
	AtaSmartAttributes struct {
		Table []struct {
			Name string `json:"name"`
			Raw  struct {
				Value  int64  `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
}

func fillDiskHealth(disk *model.DiskInfo) {
	output, err := shell.ExecuteCommand("smartctl -j -H -A " + disk.Name + " 2>/dev/null")
	if err != nil {
		return
	}

	var payload smartCtlPayload
	if err = json.Unmarshal([]byte(output), &payload); err != nil {
		return
	}

	if strings.Contains(output, "\"smart_status\"") {
		if payload.SmartStatus.Passed {
			disk.Health = "passed"
		} else {
			disk.Health = "failed"
		}
	}

	if payload.PowerOnTime.Hours > 0 {
		disk.PowerOnHours = strconv.FormatInt(int64(payload.PowerOnTime.Hours), 10)
	} else if payload.NvmeLog.PowerOnHours > 0 {
		disk.PowerOnHours = strconv.Itoa(payload.NvmeLog.PowerOnHours)
	}

	if payload.PowerCycleCount > 0 {
		disk.PowerCycleCount = strconv.Itoa(payload.PowerCycleCount)
	} else if payload.NvmeLog.PowerCycles > 0 {
		disk.PowerCycleCount = strconv.Itoa(payload.NvmeLog.PowerCycles)
	}

	if payload.Temperature.Current > 0 {
		disk.Temperature = strconv.FormatInt(int64(payload.Temperature.Current), 10) + "C"
	} else if payload.NvmeLog.Temperature > 0 {
		disk.Temperature = strconv.FormatInt(int64(payload.NvmeLog.Temperature), 10) + "C"
	}

	if strings.Contains(output, "\"percentage_used\"") {
		disk.LifeUsed = strconv.Itoa(payload.NvmeLog.PercentageUsed) + "%"
	}

	if strings.Contains(output, "\"available_spare\"") {
		disk.AvailableSpare = strconv.Itoa(payload.NvmeLog.AvailableSpare) + "%"
	}

	reallocated := getAtaAttrValue(payload, "Reallocated_Sector_Ct")
	if reallocated != "" {
		disk.ReallocatedSectors = reallocated
	}
	pending := getAtaAttrValue(payload, "Current_Pending_Sector")
	if pending != "" {
		disk.PendingSectors = pending
	}

	if disk.LifeUsed == "-" {
		if v := getAtaAttrValue(
			payload,
			"Media_Wearout_Indicator",
			"Wear_Leveling_Count",
			"Percent_Lifetime_Remain",
		); v != "" {
			disk.LifeUsed = v
		}
	}
}

func getAtaAttrValue(payload smartCtlPayload, names ...string) string {
	for _, item := range payload.AtaSmartAttributes.Table {
		for _, name := range names {
			if item.Name != name {
				continue
			}
			if item.Raw.Value > 0 || item.Raw.Value == 0 {
				return strconv.FormatInt(item.Raw.Value, 10)
			}
			if n := firstNumber(item.Raw.String); n != "" {
				return n
			}
		}
	}
	return ""
}

func firstNumber(raw string) string {
	matches := digitsRE.FindString(raw)
	return strings.TrimSpace(matches)
}
