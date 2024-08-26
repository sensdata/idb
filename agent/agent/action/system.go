package action

import (
	"fmt"

	"github.com/sensdata/idb/core/model"
	"github.com/shirou/gopsutil/v4/host"
)

func GetSystemInfo() (*model.SystemInfo, error) {
	var systemInfo model.SystemInfo

	info, _ := host.Info()
	systemInfo.HostName = info.Hostname
	systemInfo.Version = fmt.Sprintf("%s-%s", info.Platform, info.PlatformVersion)
	systemInfo.Kernel = info.KernelVersion
	systemInfo.Platform = info.KernelArch
	systemInfo.Vertual = info.VirtualizationSystem

	return &systemInfo, nil
}
