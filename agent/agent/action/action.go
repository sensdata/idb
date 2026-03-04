package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sensdata/idb/agent/db"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"gopkg.in/ini.v1"
)

func GetFingerprint() (*model.Fingerprint, error) {
	fingerprint, err := db.FingerprintRepo.GetFirst()
	if err != nil {
		global.LOG.Error("Failed to get fingerprint: %v", err)
		return nil, fmt.Errorf("failed to get fingerprint: %w", err)
	}
	return fingerprint, nil
}

func InitLicense() error {
	// 查找fp
	fp, err := db.FingerprintRepo.GetFirst()
	if err != nil {
		return fmt.Errorf("failed to get fingerprint: %w", err)
	}
	// 拿到licencePayload
	var licensePayload model.LicensePayload
	licenseBytes, err := base64.StdEncoding.DecodeString(fp.License)
	if err != nil {
		return fmt.Errorf("failed to decode license: %w", err)
	}
	if err := json.Unmarshal(licenseBytes, &licensePayload); err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}
	// 更新缓存
	global.SetLicense(&licensePayload)
	return nil
}

func SaveLicense(fingerprint *model.Fingerprint, publicKeyB64 []byte) error {
	defer func() {
		if err := recover(); err != nil {
			global.LOG.Error("SaveLicense failed: %v", err)
		}
	}()

	// 先验证签名
	if err := utils.VerifyLicenseSignature(fingerprint.License, fingerprint.Signature, publicKeyB64); err != nil {
		return fmt.Errorf("failed to verify license signature: %w", err)
	}

	// 根据指纹值查找
	fp, err := db.FingerprintRepo.GetFirst(
		db.FingerprintRepo.WithByFingerprint(fingerprint.Fingerprint))
	if err != nil {
		return fmt.Errorf("failed to get fingerprint: %w", err)
	}
	// 更新验证结果
	if err := db.FingerprintRepo.Update(
		fp.ID,
		map[string]interface{}{
			"license":        fingerprint.License,
			"signature":      fingerprint.Signature,
			"last_verify_at": time.Now(), //首次获取，设置验证时间为本地时间
		}); err != nil {
		return fmt.Errorf("failed to save license: %w", err)
	}

	// Base64 decode license
	licenseBytes, err := base64.StdEncoding.DecodeString(fingerprint.License)
	if err != nil {
		return fmt.Errorf("failed to decode license: %w", err)
	}

	// Parse JSON
	var licensePayload model.LicensePayload
	if err := json.Unmarshal(licenseBytes, &licensePayload); err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}
	global.SetLicense(&licensePayload)

	return nil
}

func UpdateLicense(verifyResp *model.VerifyLicenseResponse) error {
	if !verifyResp.Valid {
		// TODO: 验证失败时，进行一些管控
		return fmt.Errorf("license is invalid")
	}

	// 解析过期时间
	expireAt, err := time.Parse(time.RFC3339, verifyResp.ExpireAt)
	if err != nil {
		return fmt.Errorf("failed to parse expire time: %w", err)
	}

	// 查找fp
	fp, err := db.FingerprintRepo.GetFirst()
	if err != nil {
		return fmt.Errorf("failed to get fingerprint: %w", err)
	}

	// 更新 license 内容
	var licensePayload model.LicensePayload
	licenseBytes, err := base64.StdEncoding.DecodeString(fp.License)
	if err != nil {
		return fmt.Errorf("failed to decode license: %w", err)
	}
	if err := json.Unmarshal(licenseBytes, &licensePayload); err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}
	licensePayload.LicenseType = verifyResp.LicenseType
	licensePayload.ExpireAt = expireAt

	licenseJson, err := utils.ToJSONString(licensePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal license: %w", err)
	}
	license := base64.StdEncoding.EncodeToString([]byte(licenseJson))

	if err := db.FingerprintRepo.Update(
		fp.ID,
		map[string]interface{}{
			"last_verify_at": time.Now(),
			"license":        license,
		}); err != nil {
		return fmt.Errorf("failed to update fingerprint: %w", err)
	}

	// 更新缓存
	global.SetLicense(&licensePayload)
	return nil
}

func SetTime(req model.SetTimeReq) error {
	// 将时间戳转换为时间对象
	t := time.Unix(req.Timestamp, 0)
	timeStr := t.Format("2006-01-02 15:04:05")

	// 检查系统类型
	switch runtime.GOOS {
	case "linux":
		// 首先尝试使用 timedatectl（现代 Linux 系统）
		if err := utils.ExecCmd(fmt.Sprintf("sudo timedatectl set-time '%s'", timeStr)); err != nil {
			// 如果失败，尝试使用传统的 date 命令
			if out, err := utils.Execf("sudo date -s '%s'", timeStr); err != nil {
				return fmt.Errorf("set time failed: %s", out)
			}
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil
}

func SetTimezone(req model.SetTimezoneReq) error {
	// 检查时区文件是否存在
	if _, err := os.Stat(fmt.Sprintf("/usr/share/zoneinfo/%s", req.Timezone)); err != nil {
		return fmt.Errorf("invalid timezone: %s", req.Timezone)
	}

	// 首先尝试使用 timedatectl
	if err := utils.ExecCmd(fmt.Sprintf("sudo timedatectl set-timezone %s", req.Timezone)); err != nil {
		// 如果失败，尝试直接设置时区链接
		if err := utils.ExecCmd(fmt.Sprintf("sudo ln -sf /usr/share/zoneinfo/%s /etc/localtime", req.Timezone)); err != nil {
			return fmt.Errorf("set timezone failed: %v", err)
		}
	}
	return nil
}

func SyncTime() error {
	go func() {
		// 首先检查是否支持 systemd
		global.LOG.Info("try sync with systemd")
		if _, err := utils.Exec("command -v timedatectl"); err == nil {
			global.LOG.Info("disable ntp")

			// 先停止 NTP 服务
			global.LOG.Info("stop ntp")
			if err := utils.ExecCmd("sudo timedatectl set-ntp false"); err == nil {
				time.Sleep(1 * time.Second)
			}

			// 重新启用 NTP
			global.LOG.Info("enable ntp")
			if err := utils.ExecCmd("sudo timedatectl set-ntp true"); err == nil {
				// 等待更长时间让时间同步
				for i := 0; i < 5; i++ {
					time.Sleep(3 * time.Second)
					// 验证同步状态
					global.LOG.Info("sync checking")
					output, err := utils.Exec("timedatectl show --property=NTPSynchronized --value")
					if err == nil && strings.TrimSpace(output) == "yes" {
						global.LOG.Info("sync ok")
						// 同步成功后，同步硬件时钟
						if err := utils.ExecCmd("sudo hwclock --systohc"); err == nil {
							global.LOG.Info("sync hwclock ok")
						}
						return
					}
				}
				// 如果15秒后仍未同步成功，继续尝试其他方法
			}
		}

		// 尝试使用 systemd-timesyncd
		global.LOG.Info("try sync with systemd-timesyncd")
		if _, err := utils.Exec("systemctl list-unit-files systemd-timesyncd.service"); err == nil {
			global.LOG.Info("stop ")

			// 先停止服务
			global.LOG.Info("stop systemd-timesyncd")
			if err := utils.ExecCmd("sudo systemctl stop systemd-timesyncd"); err == nil {
				time.Sleep(1 * time.Second)
			}

			global.LOG.Info("restart systemd-timesyncd")
			if err := utils.ExecCmd("sudo systemctl restart systemd-timesyncd"); err == nil {
				// 等待更长时间让时间同步
				for i := 0; i < 5; i++ {
					time.Sleep(3 * time.Second)
					global.LOG.Info("sync checking")
					if out, err := utils.Exec("timedatectl status"); err == nil &&
						strings.Contains(out, "System clock synchronized: yes") {
						global.LOG.Info("sync ok")
						// 同步成功后，同步硬件时钟
						if err := utils.ExecCmd("sudo hwclock --systohc"); err != nil {
							global.LOG.Error("sync hwclock failed: %v", err)
						}
						return
					}
				}
			}
		}

		// 最后尝试使用 ntpdate
		global.LOG.Info("try sync with ntpdate")
		if _, err := utils.Exec("command -v ntpdate"); err == nil {
			global.LOG.Info("stop")
			// 先停止可能运行的 NTP 服务
			if err := utils.ExecCmd("sudo systemctl stop systemd-timesyncd"); err != nil {
				global.LOG.Error("stop systemd-timesyncd failed: %v", err)
			}
			time.Sleep(1 * time.Second)

			global.LOG.Info("start")
			if err := utils.ExecCmd("sudo ntpdate pool.ntp.org"); err == nil {
				global.LOG.Info("sync ok")
				// 同步成功后，同步硬件时钟
				if err := utils.ExecCmd("sudo hwclock --systohc"); err != nil {
					global.LOG.Error("sync hwclock failed: %v", err)
				}
				return
			}
		}

		global.LOG.Error("sync time failed")
	}()

	return nil
}

func ClearMemCache() error {
	// 先执行 sync 确保数据写入磁盘
	if err := utils.ExecCmd("sync"); err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	// 检查 drop_caches 文件是否存在
	if _, err := os.Stat("/proc/sys/vm/drop_caches"); err != nil {
		return fmt.Errorf("system does not support memory cache clearing")
	}

	// 尝试直接写入
	if err := utils.ExecCmd("sudo sh -c 'echo 3 > /proc/sys/vm/drop_caches'"); err != nil {
		// 如果失败，尝试使用 sysctl 命令
		if err := utils.ExecCmd("sudo sysctl -w vm.drop_caches=3"); err != nil {
			return fmt.Errorf("clear cache failed: %v", err)
		}
	}

	return nil
}

func SetAutoClearInterval(req model.AutoClearMemCacheReq) error {
	// 检查系统是否支持 crontab
	if _, err := utils.Exec("command -v crontab"); err != nil {
		return fmt.Errorf("crontab is not available on this system")
	}

	// 检查是否支持清理缓存
	if _, err := os.Stat("/proc/sys/vm/drop_caches"); err != nil {
		return fmt.Errorf("system does not support memory cache clearing")
	}

	// 移除现有的自动清理任务
	if err := utils.ExecCmd("crontab -l | grep -v 'drop_caches' | crontab -"); err != nil {
		// 如果失败，可能是因为还没有 crontab，尝试创建空的
		if err := utils.ExecCmd("echo '' | crontab -"); err != nil {
			return fmt.Errorf("initialize crontab failed: %v", err)
		}
	}

	if req.Interval <= 0 {
		return nil // 如果间隔小于等于0，表示取消自动清理
	}

	// 创建新的定时任务，使用 sysctl 命令作为备选
	cronCmd := fmt.Sprintf("echo '0 */%d * * * sync && (echo 3 | sudo tee /proc/sys/vm/drop_caches || sudo sysctl -w vm.drop_caches=3) > /dev/null 2>&1' | crontab -", req.Interval)
	if err := utils.ExecCmd(cronCmd); err != nil {
		return fmt.Errorf("set auto clear interval failed: %v", err)
	}

	return nil
}

func GetAutoClearInterval() (int, error) {
	// 检查系统是否支持 crontab
	if _, err := utils.Exec("command -v crontab"); err != nil {
		return 0, fmt.Errorf("crontab is not available on this system")
	}

	// 获取当前用户的 crontab 内容
	output, err := utils.Exec("crontab -l")
	if err != nil {
		return 0, nil // 如果没有 crontab 或获取失败，返回 0 表示未设置自动清理
	}

	// 查找包含 drop_caches 的行
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "drop_caches") {
			// 解析 cron 表达式中的小时间隔
			// 格式类似：0 */6 * * * ...
			fields := strings.Fields(line)
			if len(fields) >= 2 && strings.HasPrefix(fields[1], "*/") {
				intervalStr := strings.TrimPrefix(fields[1], "*/")
				interval, err := strconv.Atoi(intervalStr)
				if err != nil {
					return 0, fmt.Errorf("parse interval failed: %v", err)
				}
				return interval, nil
			}
		}
	}

	return 0, nil // 如果没有找到相关配置，返回 0 表示未设置自动清理
}

func CreateSwap(req model.CreateSwapReq) error {
	var size string
	if req.Size >= 1024 {
		size = fmt.Sprintf("%dG", req.Size/1024)
	} else {
		size = fmt.Sprintf("%dM", req.Size)
	}

	// 首先尝试使用 fallocate（更快）
	err := utils.ExecCmd(fmt.Sprintf("sudo fallocate -l %s /swapfile", size))
	if err != nil {
		// 如果 fallocate 失败，使用 dd 命令（更通用）
		blocks := req.Size * 1024 // 转换为 KB
		if err := utils.ExecCmd(fmt.Sprintf("sudo dd if=/dev/zero of=/swapfile bs=1024 count=%d", blocks)); err != nil {
			return fmt.Errorf("create swap file failed: %v", err)
		}
	}

	// 设置权限
	if err := utils.ExecCmd("sudo chmod 600 /swapfile"); err != nil {
		return fmt.Errorf("set swap file permission failed: %v", err)
	}

	// 格式化为 swap
	if err := utils.ExecCmd("sudo mkswap /swapfile"); err != nil {
		return fmt.Errorf("format swap file failed: %v", err)
	}

	// 启用 swap
	if err := utils.ExecCmd("sudo swapon /swapfile"); err != nil {
		return fmt.Errorf("enable swap failed: %v", err)
	}

	// 可选：添加到 fstab 使其开机自动挂载
	fstabEntry := "/swapfile none swap sw 0 0"
	if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee -a /etc/fstab", fstabEntry)); err != nil {
		return fmt.Errorf("add swap to fstab failed: %v", err)
	}

	return nil
}

func DeleteSwap() error {
	// 先关闭 swap
	if err := utils.ExecCmd("sudo swapoff /swapfile"); err != nil {
		return fmt.Errorf("disable swap failed: %v", err)
	}

	// 从 fstab 中移除
	if err := utils.ExecCmd("sudo sed -i '/\\/swapfile/d' /etc/fstab"); err != nil {
		return fmt.Errorf("remove swap from fstab failed: %v", err)
	}

	// 删除文件
	if err := utils.ExecCmd("sudo rm -f /swapfile"); err != nil {
		return fmt.Errorf("remove swap file failed: %v", err)
	}

	return nil
}

func UpdateDnsSettings(req model.UpdateDnsSettingsReq) error {
	// 检查是否使用 systemd-resolved
	if _, err := os.Stat("/run/systemd/resolve/resolv.conf"); err == nil {
		// 使用 systemd-resolved 的方式修改 DNS，一次性设置所有 DNS 服务器
		dnsServers := strings.Join(req.Servers, " ")
		targets := getActiveDNSInterfaces()
		if len(targets) == 0 {
			if out, cmdErr := utils.Exec("resolvectl dns"); cmdErr == nil {
				targets = parseResolvectlTargets(out)
			}
		}
		if len(targets) == 0 {
			return fmt.Errorf("no active network interface found for DNS update")
		}

		for _, iface := range targets {
			if err := utils.ExecCmd(fmt.Sprintf("sudo resolvectl dns %s %s", iface, dnsServers)); err != nil {
				return fmt.Errorf("update DNS settings failed on %s: %v", iface, err)
			}
		}

		// 对于 systemd-resolved，我们将超时和重试设置写入 resolved.conf
		if req.Timeout > 0 || req.Retry > 0 {
			resolvedConf := "/etc/systemd/resolved.conf"
			if err := utils.ExecCmd(fmt.Sprintf("sudo cp %s %s.backup", resolvedConf, resolvedConf)); err != nil {
				return fmt.Errorf("backup resolved.conf failed: %v", err)
			}

			var options []string
			if req.Timeout > 0 {
				options = append(options, fmt.Sprintf("ResolveTimeoutSec=%d", req.Timeout))
			}
			if req.Retry > 0 {
				options = append(options, fmt.Sprintf("DNSStubRetryCount=%d", req.Retry))
			}

			for _, option := range options {
				if err := utils.ExecCmd(fmt.Sprintf("sudo sed -i '/^%s=/d' %s", strings.Split(option, "=")[0], resolvedConf)); err != nil {
					return fmt.Errorf("update resolved.conf failed: %v", err)
				}
				if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee -a %s", option, resolvedConf)); err != nil {
					return fmt.Errorf("update resolved.conf failed: %v", err)
				}
			}

			// 重启 systemd-resolved 服务使配置生效
			if err := utils.ExecCmd("sudo systemctl restart systemd-resolved"); err != nil {
				return fmt.Errorf("restart systemd-resolved failed: %v", err)
			}
		}

		return nil
	}

	// 如果不是使用 systemd-resolved，则使用传统方式修改 /etc/resolv.conf
	var content strings.Builder

	// 保留 search 设置
	if data, err := os.ReadFile("/etc/resolv.conf"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "search") {
				content.WriteString(line + "\n")
				break
			}
		}
	}

	// 添加超时设置
	if req.Timeout > 0 {
		content.WriteString(fmt.Sprintf("options timeout:%d\n", req.Timeout))
	}

	// 添加重试次数设置
	if req.Retry > 0 {
		content.WriteString(fmt.Sprintf("options attempts:%d\n", req.Retry))
	}

	// 添加 DNS 服务器
	for _, server := range req.Servers {
		if server != "" {
			content.WriteString(fmt.Sprintf("nameserver %s\n", server))
		}
	}

	// 备份原文件
	if err := utils.ExecCmd("sudo cp /etc/resolv.conf /etc/resolv.conf.backup"); err != nil {
		return fmt.Errorf("backup resolv.conf failed: %v", err)
	}

	// 写入新配置
	tmpFile := "/tmp/resolv.conf"
	if err := os.WriteFile(tmpFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("write temporary file failed: %v", err)
	}

	if err := utils.ExecCmd(fmt.Sprintf("sudo mv %s /etc/resolv.conf", tmpFile)); err != nil {
		return fmt.Errorf("update DNS settings failed: %v", err)
	}

	return nil
}

func getActiveDNSInterfaces() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		if iface.Name == "" {
			continue
		}
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		names = append(names, iface.Name)
	}
	sort.Strings(names)
	return names
}

func parseResolvectlTargets(output string) []string {
	targetSet := map[string]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}

		left := strings.TrimSpace(line[:idx])
		if left == "" {
			continue
		}

		if strings.HasPrefix(left, "Link ") {
			start := strings.LastIndex(left, "(")
			end := strings.LastIndex(left, ")")
			if start != -1 && end > start+1 {
				left = strings.TrimSpace(left[start+1 : end])
			}
		}

		if left == "" || strings.EqualFold(left, "global") {
			continue
		}
		targetSet[left] = struct{}{}
	}

	targets := make([]string, 0, len(targetSet))
	for name := range targetSet {
		targets = append(targets, name)
	}
	sort.Strings(targets)
	return targets
}

func UpdateHostName(req model.UpdateHostNameReq) error {
	global.LOG.Info("start update hostname", "hostname", req.HostName)

	// 检查系统类型
	switch runtime.GOOS {
	case "linux":
		// 首先尝试使用 hostnamectl
		global.LOG.Info("try update with hostnamectl")
		if err := utils.ExecCmd(fmt.Sprintf("sudo hostnamectl set-hostname %s", req.HostName)); err != nil {
			// 如果 hostnamectl 失败，尝试使用传统方法
			global.LOG.Info("hostnamectl failed, try traditional method")

			// 1. 更新当前主机名
			global.LOG.Info("update current hostname")
			if err := utils.ExecCmd(fmt.Sprintf("sudo hostname %s", req.HostName)); err != nil {
				return fmt.Errorf("set current hostname failed: %v", err)
			}

			// 2. 更新 /etc/hostname
			global.LOG.Info("update /etc/hostname")
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee /etc/hostname", req.HostName)); err != nil {
				return fmt.Errorf("update /etc/hostname failed: %v", err)
			}

			// 3. 更新 /etc/hosts 中对应的条目
			// 备份原文件
			global.LOG.Info("backup hosts file")
			if err := utils.ExecCmd("sudo cp /etc/hosts /etc/hosts.backup"); err != nil {
				return fmt.Errorf("backup hosts file failed: %v", err)
			}

			// 更新 hosts 文件中的本地主机名条目
			global.LOG.Info("update /etc/hosts")
			sedCmd := fmt.Sprintf("sudo sed -i 's/127.0.1.1.*/127.0.1.1\\t%s/g' /etc/hosts", req.HostName)
			if err := utils.ExecCmd(sedCmd); err != nil {
				return fmt.Errorf("update /etc/hosts failed: %v", err)
			}
		}
		global.LOG.Info("update hostname success")
	default:
		global.LOG.Error("unsupported operating system", "os", runtime.GOOS)
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

func UpdateSystemSettings(req model.UpdateSystemSettingsReq) error {
	sysctlConfDir := "/etc/sysctl.d"
	sysctlConfFile := "/etc/sysctl.d/90-idb-sysinfo.conf"
	sysctlLines := []string{
		fmt.Sprintf("fs.inotify.max_user_watches = %d", req.MaxWatchFiles),
		fmt.Sprintf("fs.inotify.max_user_instances = %d", req.MaxWatchInstances),
		fmt.Sprintf("fs.inotify.max_queued_events = %d", req.MaxQueuedEvents),
		fmt.Sprintf("vm.swappiness = %d", req.Swappiness),
		fmt.Sprintf("vm.max_map_count = %d", req.MaxMapCount),
		fmt.Sprintf("net.core.somaxconn = %d", req.Somaxconn),
		fmt.Sprintf("net.ipv4.tcp_max_syn_backlog = %d", req.TcpMaxSynBacklog),
		fmt.Sprintf("fs.file-max = %d", req.FileMax),
		fmt.Sprintf("kernel.pid_max = %d", req.PidMax),
		fmt.Sprintf("vm.overcommit_memory = %d", req.OvercommitMemory),
		fmt.Sprintf("vm.overcommit_ratio = %d", req.OvercommitRatio),
	}

	if err := utils.ExecCmd("sudo mkdir -p " + sysctlConfDir); err != nil {
		return fmt.Errorf("prepare sysctl.d directory failed: %v", err)
	}

	sysctlContent := strings.Join(sysctlLines, "\n")
	if err := utils.ExecCmd(
		fmt.Sprintf("printf '%s\\n' | sudo tee %s >/dev/null", sysctlContent, sysctlConfFile),
	); err != nil {
		return fmt.Errorf("persist sysctl settings failed: %v", err)
	}

	// 定向加载配置，避免 sysctl --system 受其他无关配置影响
	if err := utils.ExecCmd("sudo sysctl -p " + sysctlConfFile); err != nil {
		return fmt.Errorf("reload sysctl settings failed: %v", err)
	}

	// 修改最大文件打开数量（使用 system.conf.d drop-in，避免直接改主配置）
	if hasSystemd() {
		systemdConfDir := "/etc/systemd/system.conf.d"
		systemdConfFile := "/etc/systemd/system.conf.d/90-idb.conf"
		systemdContent := fmt.Sprintf("[Manager]\nDefaultLimitNOFILE=%d", req.MaxOpenFiles)

		if err := utils.ExecCmd("sudo mkdir -p " + systemdConfDir); err != nil {
			return fmt.Errorf("prepare system.conf.d directory failed: %v", err)
		}

		if err := utils.ExecCmd(
			fmt.Sprintf("printf '%s\\n' | sudo tee %s >/dev/null", systemdContent, systemdConfFile),
		); err != nil {
			return fmt.Errorf("persist open files setting failed: %v", err)
		}

		if err := utils.ExecCmd("sudo systemctl daemon-reexec"); err != nil {
			global.LOG.Warn("skip systemd daemon reexec due to error", "err", err)
		}
	} else {
		global.LOG.Warn("systemd is not available, skip applying DefaultLimitNOFILE")
	}

	if err := applyAndPersistTHP(req.TransparentHugePage); err != nil {
		return err
	}

	return nil
}

func GetSystemSettings() (*model.SystemSettings, error) {
	settings := model.SystemSettings{
		MaxWatchFiles:       readProcIntWithDefault("/proc/sys/fs/inotify/max_user_watches", 8192),
		MaxWatchInstances:   readProcIntWithDefault("/proc/sys/fs/inotify/max_user_instances", 128),
		MaxQueuedEvents:     readProcIntWithDefault("/proc/sys/fs/inotify/max_queued_events", 16384),
		Swappiness:          readProcIntWithDefault("/proc/sys/vm/swappiness", 60),
		MaxMapCount:         readProcIntWithDefault("/proc/sys/vm/max_map_count", 65530),
		Somaxconn:           readProcIntWithDefault("/proc/sys/net/core/somaxconn", 4096),
		TcpMaxSynBacklog:    readProcIntWithDefault("/proc/sys/net/ipv4/tcp_max_syn_backlog", 4096),
		FileMax:             readProcIntWithDefault("/proc/sys/fs/file-max", 65535),
		PidMax:              readProcIntWithDefault("/proc/sys/kernel/pid_max", 32768),
		OvercommitMemory:    readProcIntWithDefault("/proc/sys/vm/overcommit_memory", 0),
		OvercommitRatio:     readProcIntWithDefault("/proc/sys/vm/overcommit_ratio", 50),
		TransparentHugePage: readTHPMode(),
	}

	// 获取最大文件打开数量
	// 优先读取 drop-in 配置，后者覆盖前者
	systemConf := "/etc/systemd/system.conf"
	dropInConf := "/etc/systemd/system.conf.d/90-idb.conf"
	files := []string{systemConf}
	if _, err := os.Stat(dropInConf); err == nil {
		files = append(files, dropInConf)
	}
	var cfg *ini.File
	var err error
	if len(files) > 0 {
		sources := make([]interface{}, 0, len(files)-1)
		for _, file := range files[1:] {
			sources = append(sources, file)
		}
		cfg, err = ini.Load(files[0], sources...)
	}
	if err == nil && cfg != nil {
		if noFile := cfg.Section("Manager").Key("DefaultLimitNOFILE").String(); noFile != "" {
			if maxOpenFilesInt, err := strconv.Atoi(noFile); err == nil {
				settings.MaxOpenFiles = maxOpenFilesInt
			}
		}
	}
	// fallback：读取当前 shell 的ulimit值或系统最大值
	if settings.MaxOpenFiles == 0 {
		if out, err := utils.Exec("ulimit -n"); err == nil {
			if v, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
				settings.MaxOpenFiles = v
			}
		}
	}
	if settings.MaxOpenFiles == 0 {
		if data, err := os.ReadFile("/proc/sys/fs/file-max"); err == nil {
			if v, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
				settings.MaxOpenFiles = v
			}
		}
	}
	if settings.MaxOpenFiles == 0 {
		settings.MaxOpenFiles = 65535
	}

	return &settings, nil
}

func readProcIntWithDefault(path string, fallback int) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fallback
	}
	return value
}

func readTHPMode() string {
	data, err := os.ReadFile("/sys/kernel/mm/transparent_hugepage/enabled")
	if err != nil {
		return "madvise"
	}
	parts := strings.Fields(string(data))
	for _, part := range parts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			return strings.Trim(part, "[]")
		}
	}
	return "madvise"
}

func applyAndPersistTHP(mode string) error {
	if mode != "always" && mode != "madvise" && mode != "never" {
		return fmt.Errorf("invalid transparent huge page mode: %s", mode)
	}

	thpPath := "/sys/kernel/mm/transparent_hugepage/enabled"
	if _, err := os.Stat(thpPath); err != nil {
		global.LOG.Warn("transparent hugepage is not available, skip applying", "err", err)
		return nil
	}
	if err := utils.ExecCmd(fmt.Sprintf("echo %s | sudo tee %s >/dev/null", mode, thpPath)); err != nil {
		return fmt.Errorf("apply THP mode failed: %v", err)
	}

	if !hasSystemd() {
		global.LOG.Warn("systemd is not available, skip persisting THP service")
		return nil
	}

	serviceDir := "/etc/systemd/system"
	serviceFile := "/etc/systemd/system/idb-thp.service"
	serviceContent := strings.Join([]string{
		"[Unit]",
		"Description=Apply Transparent Hugepage policy for IDB",
		"After=local-fs.target",
		"",
		"[Service]",
		"Type=oneshot",
		fmt.Sprintf("ExecStart=/bin/sh -c 'echo %s > %s'", mode, thpPath),
		"RemainAfterExit=yes",
		"",
		"[Install]",
		"WantedBy=multi-user.target",
	}, "\n")

	if err := utils.ExecCmd("sudo mkdir -p " + serviceDir); err != nil {
		return fmt.Errorf("prepare THP service directory failed: %v", err)
	}

	if err := utils.ExecCmd(
		fmt.Sprintf("cat <<'EOF' | sudo tee %s >/dev/null\n%s\nEOF", serviceFile, serviceContent),
	); err != nil {
		return fmt.Errorf("persist THP mode failed: %v", err)
	}

	if err := utils.ExecCmd("sudo systemctl daemon-reload"); err != nil {
		return fmt.Errorf("reload systemd daemon failed: %v", err)
	}
	if err := utils.ExecCmd("sudo systemctl unmask idb-thp.service"); err != nil {
		return fmt.Errorf("unmask THP service failed: %v", err)
	}
	if err := utils.ExecCmd("sudo systemctl enable --now idb-thp.service"); err != nil {
		return fmt.Errorf("enable THP service failed: %v", err)
	}

	return nil
}

func hasSystemd() bool {
	if !utils.Which("systemctl") {
		return false
	}
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return false
	}
	return true
}
