package action

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
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
			utils.ExecCmd("sudo systemctl stop systemd-timesyncd")
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
		if err := utils.ExecCmd(fmt.Sprintf("sudo resolvectl dns eth0 %s", dnsServers)); err != nil {
			return fmt.Errorf("update DNS settings failed: %v", err)
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
	// 修改最大监控文件个数
	if req.MaxWatchFiles > 0 {
		// 立即生效
		if err := utils.ExecCmd(fmt.Sprintf("sudo sysctl -w fs.inotify.max_user_watches=%d", req.MaxWatchFiles)); err != nil {
			return fmt.Errorf("set max watch files failed: %v", err)
		}

		// 持久化到 sysctl.d
		confFile := "/etc/sysctl.d/90-max-watches.conf"
		content := fmt.Sprintf("fs.inotify.max_user_watches = %d\n", req.MaxWatchFiles)
		if err := os.WriteFile(confFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("persist max watch files setting failed: %v", err)
		}
	}

	// 修改最大文件打开数量
	if req.MaxOpenFiles > 0 {
		confFile := "/etc/systemd/system.conf"
		cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, confFile)
		if err != nil {
			// 文件不存在 → 创建
			cfg = ini.Empty()
		}
		sec := cfg.Section("Manager")
		if k := sec.Key("DefaultLimitNOFILE"); k != nil && k.Value() != "" {
			k.SetValue(strconv.Itoa(req.MaxOpenFiles))
		} else {
			if _, err := sec.NewKey("DefaultLimitNOFILE", strconv.Itoa(req.MaxOpenFiles)); err != nil {
				return fmt.Errorf("failed to create key DefaultLimitNOFILE: %w", err)
			}
		}
		if err := cfg.SaveTo(confFile); err != nil {
			return fmt.Errorf("update system.conf failed: %v", err)
		}
		// 让 systemd 立即重新加载配置
		if err := utils.ExecCmd("sudo systemctl daemon-reexec"); err != nil {
			return fmt.Errorf("reload systemd daemon failed: %v", err)
		}
	}

	// 重新加载 sysctl 配置
	if err := utils.ExecCmd("sudo sysctl --system"); err != nil {
		return fmt.Errorf("reload sysctl settings failed: %v", err)
	}

	return nil
}

func GetSystemSettings() (*model.SystemSettings, error) {
	var maxWatchFilesInt, maxOpenFilesInt int

	// 获取最大监控文件个数
	data, err := os.ReadFile("/proc/sys/fs/inotify/max_user_watches")
	if err != nil {
		return nil, fmt.Errorf("get max watch files failed: %v", err)
	}
	if maxWatchFilesInt, err = strconv.Atoi(strings.TrimSpace(string(data))); err != nil {
		return nil, fmt.Errorf("parse max watch files failed: %v", err)
	}

	// 获取最大文件打开数量
	// 读取 systemd 配置
	confFile := "/etc/systemd/system.conf"
	if cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, confFile); err == nil {
		if noFile := cfg.Section("Manager").Key("DefaultLimitNOFILE").String(); noFile != "" {
			if maxOpenFilesInt, err = strconv.Atoi(noFile); err != nil {
				maxOpenFilesInt = 0
			}
		}
	}
	// fallback：读取当前 shell 的ulimit值或系统最大值
	if maxOpenFilesInt == 0 {
		if out, err := utils.Exec("ulimit -n"); err == nil {
			if v, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
				maxOpenFilesInt = v
			}
		}
	}
	if maxOpenFilesInt == 0 {
		if data, err := os.ReadFile("/proc/sys/fs/file-max"); err == nil {
			if v, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
				maxOpenFilesInt = v
			}
		}
	}
	if maxOpenFilesInt == 0 {
		maxOpenFilesInt = 65535
	}
	return &model.SystemSettings{
		MaxWatchFiles: maxWatchFilesInt,
		MaxOpenFiles:  maxOpenFilesInt,
	}, nil
}
