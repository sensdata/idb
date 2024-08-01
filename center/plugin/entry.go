package plugin

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/sysinfo"
)

func RegisterPlugins() {
	// 注册sysinfo
	conn.RegisterIdbPlugin(&sysinfo.SysInfo{})

	// 执行所有模块的初始化
	conn.InitializePlugins()
}
