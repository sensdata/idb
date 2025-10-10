package plugin

import (
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/plugin/certificate"
	"github.com/sensdata/idb/center/plugin/crontab/backend/crontab"
	"github.com/sensdata/idb/center/plugin/docker/backend/docker"
	"github.com/sensdata/idb/center/plugin/fileman/backend/fileman"
	"github.com/sensdata/idb/center/plugin/logrotate/backend/logrotate"
	"github.com/sensdata/idb/center/plugin/nftable/backend/nftable"
	"github.com/sensdata/idb/center/plugin/process"
	"github.com/sensdata/idb/center/plugin/service/backend/serviceman"
	"github.com/sensdata/idb/center/plugin/ssh/backend/sshman"
	"github.com/sensdata/idb/center/plugin/sysinfo/backend/sysinfo"
	"github.com/sensdata/idb/center/plugin/systemctl"
	"github.com/sensdata/idb/core/constant"
)

func RegisterPlugins() {
	// 注册 sysinfo
	conn.RegisterIdbPlugin(&sysinfo.SysInfo{})
	// 注册 files
	conn.RegisterIdbPlugin(&fileman.FileMan{})
	// 注册 ssh
	conn.RegisterIdbPlugin(&sshman.SSHMan{})
	// 注册 sysctl
	conn.RegisterIdbPlugin(&systemctl.SystemCtl{})
	// 注册 service
	conn.RegisterIdbPlugin(&serviceman.ServiceMan{})
	// 注册 logrotate
	conn.RegisterIdbPlugin(&logrotate.LogRotate{})
	// 注册 crontab
	conn.RegisterIdbPlugin(&crontab.CronTab{})
	// 注册 docker
	conn.RegisterIdbPlugin(&docker.DockerMan{AppDir: constant.AgentDockerDir})
	// 注册 nftable
	conn.RegisterIdbPlugin(&nftable.NFTable{})
	// 注册 certificates
	conn.RegisterIdbPlugin(&certificate.CertificateMan{})
	// 注册 process
	conn.RegisterIdbPlugin(&process.Process{})
	// 执行所有模块的初始化
	conn.InitializePlugins()
}

func StartPlugins() {
	conn.StartPlugins()
}
