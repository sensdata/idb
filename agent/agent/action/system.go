package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/shirou/gopsutil/v4/host"
)

func GetSystemInfo() (*model.SystemInfo, error) {
	var systemInfo model.SystemInfo

	info, err := host.Info()
	if err != nil {
		global.LOG.Error("failed to get host info: %v", err)
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	systemInfo.HostName = info.Hostname
	systemInfo.Distribution = info.Platform
	systemInfo.DistributionVersion = info.PlatformVersion
	systemInfo.Version = strings.TrimSpace(
		fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion),
	)
	systemInfo.Kernel = info.KernelVersion
	systemInfo.Platform = info.Platform // backward compatibility
	systemInfo.Arch = info.KernelArch
	systemInfo.OS = info.OS
	systemInfo.Virtual = detectVirtualization(info)
	systemInfo.Vertual = systemInfo.Virtual // backward compatibility
	systemInfo.Uptime = int64(info.Uptime)
	systemInfo.MachineID = info.HostID

	fqdnOutput, fqdnErr := shell.ExecuteCommand("hostname -f")
	if fqdnErr != nil {
		systemInfo.FQDN = systemInfo.HostName
	} else {
		fqdn := strings.TrimSpace(fqdnOutput)
		if fqdn == "" {
			fqdn = systemInfo.HostName
		}
		systemInfo.FQDN = fqdn
	}

	return &systemInfo, nil
}

func detectVirtualization(info *host.InfoStat) string {
	candidates := []string{
		detectByCgroup(),
		detectBySystemdVirt(),
		normalizeVirtualization(info.VirtualizationSystem),
	}

	virt := ""
	for _, item := range candidates {
		if item == "" || item == "none" {
			continue
		}
		virt = item
		break
	}

	if virt == "" {
		return ""
	}

	role := strings.TrimSpace(strings.ToLower(info.VirtualizationRole))
	switch role {
	case "guest", "host":
		return fmt.Sprintf("%s (%s)", virt, role)
	default:
		return virt
	}
}

func detectBySystemdVirt() string {
	output, err := shell.ExecuteCommand("systemd-detect-virt")
	if err != nil {
		return ""
	}
	return normalizeVirtualization(output)
}

func detectByCgroup() string {
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return ""
	}

	content := strings.ToLower(string(data))
	switch {
	case strings.Contains(content, "kubepods"):
		return "kubernetes"
	case strings.Contains(content, "containerd"):
		return "containerd"
	case strings.Contains(content, "docker"):
		return "docker"
	case strings.Contains(content, "podman"):
		return "podman"
	case strings.Contains(content, "lxc"):
		return "lxc"
	default:
		return ""
	}
}

func normalizeVirtualization(value string) string {
	virt := strings.TrimSpace(strings.ToLower(value))
	if virt == "" {
		return ""
	}

	switch virt {
	case "none", "physical", "baremetal", "bare-metal":
		return ""
	case "microsoft":
		return "hyper-v"
	case "wsl2":
		return "wsl"
	default:
		return virt
	}
}
