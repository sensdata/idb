package action

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func SetTime(req model.SetTimeReq) error {
	// 将时间戳转换为时间对象
	t := time.Unix(req.Timestamp, 0)
	timeStr := t.Format("2006-01-02 15:04:05")
	// 根据不同操作系统执行不同的时间设置命令
	switch runtime.GOOS {
	case "linux":
		// 设置系统时间
		out, err := utils.Execf("date -s %s", timeStr)
		if err != nil {
			return fmt.Errorf("set time failed: %s", out)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil
}

func SetTimezone(req model.SetTimezoneReq) error {
	return utils.ExecCmd(fmt.Sprintf("sudo timedatectl set-timezone %s", req.Timezone))
}

func SyncTime() error {
	return utils.ExecCmd("sudo timedatectl set-ntp true")
}

func ClearMemCache() error {
	// 先执行 sync 确保数据写入磁盘
	if err := utils.ExecCmd("sync"); err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	// 清理页面缓存、目录项和 inode
	if err := utils.ExecCmd("sudo sh -c 'echo 3 > /proc/sys/vm/drop_caches'"); err != nil {
		return fmt.Errorf("clear cache failed: %v", err)
	}

	return nil
}

func SetAutoClearInterval(req model.AutoClearMemCacheReq) error {
	// 移除现有的自动清理任务
	if err := utils.ExecCmd("crontab -l | grep -v 'drop_caches' | crontab -"); err != nil {
		return fmt.Errorf("remove existing cron job failed: %v", err)
	}

	if req.Interval <= 0 {
		return nil // 如果间隔小于等于0，表示取消自动清理
	}

	// 创建新的定时任务
	cronCmd := fmt.Sprintf("echo '0 */%d * * * sync && echo 3 | sudo tee /proc/sys/vm/drop_caches > /dev/null' | crontab -", req.Interval)
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
	return utils.ExecCmd(fmt.Sprintf("sudo fallocate -l %s /swapfile && sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile", size))
}

func DeleteSwap() error {
	return utils.ExecCmd("sudo swapoff /swapfile && sudo rm /swapfile")
}
