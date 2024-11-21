package constant

const (
	IDBType          = "net.idb.type"
	IDBName          = "net.idb.name"
	IDBVersion       = "net.idb.version"
	IDBUpdateVersion = "net.idb.update_version"
	IDBPanel         = "net.idb.panel"
)

const (
	ContainerOpStart   = "start"
	ContainerOpStop    = "stop"
	ContainerOpRestart = "restart"
	ContainerOpKill    = "kill"
	ContainerOpPause   = "pause"
	ContainerOpUnpause = "unpause"
	ContainerOpRename  = "rename"
	ContainerOpRemove  = "remove"

	ComposeOpStop    = "stop"
	ComposeOpRestart = "restart"
	ComposeOpRemove  = "remove"

	DaemonJsonPath = "/etc/docker/daemon.json"
)
