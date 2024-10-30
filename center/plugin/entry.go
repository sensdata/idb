package plugin

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/plugin/fileman/backend/fileman"
	"github.com/sensdata/idb/center/plugin/script/backend/scriptman"
	"github.com/sensdata/idb/center/plugin/ssh/backend/sshman"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/sysinfo"
	"github.com/sensdata/idb/center/plugin/systemctl"
)

func RegisterPlugins() {
	// 注册sysinfo
	conn.RegisterIdbPlugin(&sysinfo.Plugin)
	// 注册files
	conn.RegisterIdbPlugin(&fileman.Plugin)
	// 注册ssh
	conn.RegisterIdbPlugin(&sshman.Plugin)
	// 注册sysctl
	conn.RegisterIdbPlugin(&systemctl.Plugin)
	// 注册scripts
	conn.RegisterIdbPlugin(&scriptman.Plugin)

	// 执行所有模块的初始化
	conn.InitializePlugins()
}
