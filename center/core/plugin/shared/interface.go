package shared

import (
	"github.com/hashicorp/go-plugin"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN_MAGIC",
	MagicCookieValue: "idb_plugin_cookie",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"auth":          &AuthPlugin{},
	"scriptmanager": &ScriptMangerPlugin{},
	"mysqlmanager":  &MysqlManagerPlugin{},
}

type PluginInitConfig struct {
	API      string `json:"api"`
	HTTPS    bool   `json:"https"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	WorkDir  string `json:"work_dir"`
	WorkHost uint   `json:"work_host"`
}
