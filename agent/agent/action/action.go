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

func SyncTime() error {
	return utils.ExecCmd("sudo timedatectl set-ntp true")
}
