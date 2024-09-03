package plugin

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/plugin/fileman/backend/fileman"
	"github.com/sensdata/idb/center/plugin/ssh/backend/sshman"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/sysinfo"
	"github.com/sensdata/idb/center/plugin/systemctl"
)

func RegisterPlugins() {
	// 注册sysinfo
	conn.RegisterIdbPlugin(&sysinfo.SysInfo{})
	// 注册files
	conn.RegisterIdbPlugin(&fileman.FileMan{})
	// 注册ssh
	conn.RegisterIdbPlugin(&sshman.SSHMan{})
	// 注册sysctl
	conn.RegisterIdbPlugin(&systemctl.SystemCtl{})

	// 执行所有模块的初始化
	conn.InitializePlugins()
}
