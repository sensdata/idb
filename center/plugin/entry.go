package plugin

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/plugin/fileman/backend/fileman"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/sysinfo"
)

func RegisterPlugins() {
	// 注册sysinfo
	conn.RegisterIdbPlugin(&sysinfo.SysInfo{})
	// 注册files
	conn.RegisterIdbPlugin(&fileman.FileMan{})

	// 执行所有模块的初始化
	conn.InitializePlugins()
}
