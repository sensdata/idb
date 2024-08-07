package action

import (
	"encoding/json"
	"fmt"

	"github.com/sensdata/idb/core/model"
	"github.com/shirou/gopsutil/v4/host"
)

func GetSystemInfo() (string, error) {
	var systemInfo model.SystemInfo

	info, _ := host.Info()
	systemInfo.HostName = info.Hostname
	systemInfo.Version = fmt.Sprintf("%s-%s", info.Platform, info.PlatformVersion)
	systemInfo.Kernel = info.KernelVersion
	systemInfo.Platform = info.KernelArch
	systemInfo.Vertual = info.VirtualizationSystem

	jsonString, err := json.Marshal(systemInfo)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(jsonString), nil
}
