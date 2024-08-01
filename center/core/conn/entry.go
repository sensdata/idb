package conn

import (
	"fmt"

	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/plugin"
)

var (
	CommonRepo = repo.NewCommonRepo()

	RoleRepo      = repo.NewRoleRepo()
	UserRepo      = repo.NewUserRepo()
	GroupRepo     = repo.NewGroupRepo()
	HostRepo      = repo.NewHostRepo()
	HostGroupRepo = repo.NewHostGroupRepo()

	CONFMAN *config.Manager
	SSH     ISSHService
	CENTER  ICenter
	PLUGINS []plugin.IdbPlugin
)

func RegisterIdbPlugin(p plugin.IdbPlugin) {
	global.LOG.Info("RegisterIdbPlugin")
	PLUGINS = append(PLUGINS, p)
}

func InitializePlugins() {
	global.LOG.Info("InitializePlugins")
	for i, p := range PLUGINS {
		global.LOG.Info(fmt.Sprintf("InitializePlugins %d", i))
		p.Initialize()
	}
}
