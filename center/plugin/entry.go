package plugin

import (
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/center/plugin/certificate"
	"github.com/sensdata/idb/center/plugin/crontab/backend/crontab"
	"github.com/sensdata/idb/center/plugin/docker/backend/docker"
	"github.com/sensdata/idb/center/plugin/fileman/backend/fileman"
	"github.com/sensdata/idb/center/plugin/logrotate/backend/logrotate"
	"github.com/sensdata/idb/center/plugin/manager"
	"github.com/sensdata/idb/center/plugin/nftable/backend/nftable"
	"github.com/sensdata/idb/center/plugin/script/backend/scriptman"
	"github.com/sensdata/idb/center/plugin/service/backend/serviceman"
	"github.com/sensdata/idb/center/plugin/ssh/backend/sshman"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/sysinfo"
	"github.com/sensdata/idb/center/plugin/systemctl"
	"github.com/sensdata/idb/core/constant"
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
	// 注册scripts
	conn.RegisterIdbPlugin(&scriptman.ScriptMan{})
	// 注册service
	conn.RegisterIdbPlugin(&serviceman.ServiceMan{})
	// 注册logrotate
	conn.RegisterIdbPlugin(&logrotate.LogRotate{})
	// 注册crontab
	conn.RegisterIdbPlugin(&crontab.CronTab{})
	// 注册docker, TODO: AppDir从安装传入
	conn.RegisterIdbPlugin(&docker.DockerMan{AppDir: constant.AgentDockerDir})
	// 注册nftable
	conn.RegisterIdbPlugin(&nftable.NFTable{})
	// 注册certificates
	conn.RegisterIdbPlugin(&certificate.CertificateMan{})
	// 执行所有模块的初始化
	conn.InitializePlugins()

	// 初始化插件管理器
	manager.PluginMan = manager.NewPluginManager()
	if err := manager.PluginMan.Initialize(api.API.Router); err != nil {
		global.LOG.Error("Failed to initialize plugin manager: %v", err)
	}
}

func StartPlugins() {
	conn.StartPlugins()
}
