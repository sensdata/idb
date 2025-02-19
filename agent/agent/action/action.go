package action

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

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
	// 首先检查是否支持 systemd
	if _, err := utils.Exec("command -v timedatectl"); err == nil {
		// 尝试使用 timedatectl 启用 NTP
		if err := utils.ExecCmd("sudo timedatectl set-ntp true"); err == nil {
			return nil
		}
	}

	// 如果 timedatectl 不可用或失败，尝试使用 ntpd
	if _, err := utils.Exec("command -v ntpd"); err == nil {
		if err := utils.ExecCmd("sudo service ntpd restart"); err != nil {
			return fmt.Errorf("restart ntpd service failed: %v", err)
		}
		return nil
	}

	// 最后尝试使用 ntpdate
	if _, err := utils.Exec("command -v ntpdate"); err == nil {
		if err := utils.ExecCmd("sudo ntpdate pool.ntp.org"); err != nil {
			return fmt.Errorf("sync time with ntpdate failed: %v", err)
		}
		return nil
	}

	return fmt.Errorf("no available time sync service found")
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
		// 使用 systemd-resolved 的方式修改 DNS
		for _, server := range req.Servers {
			if err := utils.ExecCmd(fmt.Sprintf("sudo resolvectl dns eth0 %s", server)); err != nil {
				return fmt.Errorf("update DNS settings failed: %v", err)
			}
		}

		// 设置 DNS 配置参数
		if req.Timeout > 0 {
			if err := utils.ExecCmd(fmt.Sprintf("sudo resolvectl set-dns-option eth0 timeout:%d", req.Timeout)); err != nil {
				return fmt.Errorf("set DNS timeout failed: %v", err)
			}
		}

		if req.Retry > 0 {
			if err := utils.ExecCmd(fmt.Sprintf("sudo resolvectl set-dns-option eth0 attempts:%d", req.Retry)); err != nil {
				return fmt.Errorf("set DNS retry failed: %v", err)
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

func UpdateSystemSettings(req model.UpdateSystemSettingsReq) error {
	// 修改最大监控文件个数
	if req.MaxWatchFiles > 0 {
		// 立即生效
		if err := utils.ExecCmd(fmt.Sprintf("sudo sysctl -w fs.inotify.max_user_watches=%d", req.MaxWatchFiles)); err != nil {
			return fmt.Errorf("set max watch files failed: %v", err)
		}

		// 检查 sysctl.d 目录是否存在
		if _, err := os.Stat("/etc/sysctl.d"); err == nil {
			content := fmt.Sprintf("fs.inotify.max_user_watches = %d\n", req.MaxWatchFiles)
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee /etc/sysctl.d/90-max-watches.conf", content)); err != nil {
				return fmt.Errorf("persist max watch files setting failed: %v", err)
			}
		} else {
			// 如果不存在，则追加到 sysctl.conf
			content := fmt.Sprintf("\nfs.inotify.max_user_watches = %d\n", req.MaxWatchFiles)
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee -a /etc/sysctl.conf", content)); err != nil {
				return fmt.Errorf("persist max watch files setting failed: %v", err)
			}
		}
	}

	// 修改最大文件打开数量
	if req.MaxOpenFiles > 0 {
		// 修改系统级别的限制
		if err := utils.ExecCmd(fmt.Sprintf("sudo sysctl -w fs.file-max=%d", req.MaxOpenFiles)); err != nil {
			return fmt.Errorf("set system max open files failed: %v", err)
		}

		// 检查 limits.d 目录是否存在
		if _, err := os.Stat("/etc/security/limits.d"); err == nil {
			limits := fmt.Sprintf("* soft nofile %d\n* hard nofile %d\n", req.MaxOpenFiles, req.MaxOpenFiles)
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee /etc/security/limits.d/90-max-files.conf", limits)); err != nil {
				return fmt.Errorf("set user max open files failed: %v", err)
			}
		} else {
			// 如果不存在，则追加到 limits.conf
			limits := fmt.Sprintf("\n* soft nofile %d\n* hard nofile %d\n", req.MaxOpenFiles, req.MaxOpenFiles)
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee -a /etc/security/limits.conf", limits)); err != nil {
				return fmt.Errorf("set user max open files failed: %v", err)
			}
		}

		// 系统级别设置永久生效
		if _, err := os.Stat("/etc/sysctl.d"); err == nil {
			content := fmt.Sprintf("fs.file-max = %d\n", req.MaxOpenFiles)
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee /etc/sysctl.d/90-max-files.conf", content)); err != nil {
				return fmt.Errorf("persist system max open files setting failed: %v", err)
			}
		} else {
			content := fmt.Sprintf("\nfs.file-max = %d\n", req.MaxOpenFiles)
			if err := utils.ExecCmd(fmt.Sprintf("echo '%s' | sudo tee -a /etc/sysctl.conf", content)); err != nil {
				return fmt.Errorf("persist system max open files setting failed: %v", err)
			}
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
	// 首先尝试从 sysctl 获取
	maxWatchFiles, err := utils.Exec("sysctl -n fs.inotify.max_user_watches")
	if err != nil {
		// 如果失败，尝试直接读取 proc 文件系统
		maxWatchFiles, err = utils.Exec("cat /proc/sys/fs/inotify/max_user_watches")
		if err != nil {
			return nil, fmt.Errorf("get max watch files failed: %v", err)
		}
	}
	maxWatchFiles = strings.TrimSpace(maxWatchFiles)
	if maxWatchFiles != "" {
		maxWatchFilesInt, err = strconv.Atoi(maxWatchFiles)
		if err != nil {
			return nil, fmt.Errorf("parse max watch files failed: %v", err)
		}
	}

	// 获取最大文件打开数量
	// 首先尝试从 sysctl 获取
	maxOpenFiles, err := utils.Exec("sysctl -n fs.file-max")
	if err != nil {
		// 如果失败，尝试直接读取 proc 文件系统
		maxOpenFiles, err = utils.Exec("cat /proc/sys/fs/file-max")
		if err != nil {
			// 如果还是失败，尝试从 ulimit 获取
			maxOpenFiles, err = utils.Exec("ulimit -n")
			if err != nil {
				return nil, fmt.Errorf("get max open files failed: %v", err)
			}
		}
	}
	maxOpenFiles = strings.TrimSpace(maxOpenFiles)
	if maxOpenFiles != "" {
		maxOpenFilesInt, err = strconv.Atoi(maxOpenFiles)
		if err != nil {
			return nil, fmt.Errorf("parse max open files failed: %v", err)
		}
	}

	return &model.SystemSettings{
		MaxWatchFiles: maxWatchFilesInt,
		MaxOpenFiles:  maxOpenFilesInt,
	}, nil
}
