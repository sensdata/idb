package constant

const (
	IDBType          = "net.idb.type"
	IDBName          = "net.idb.name"
	IDBVersion       = "net.idb.version"
	IDBUpdateVersion = "net.idb.update_version"
	IDBPanel         = "net.idb.panel"
)

const (
	TYPE_APP   = "app"
	TYPE_Panel = "panel"
)

const (
	/*
		com.docker.compose.* 系列标签：
		com.docker.compose.project
		com.docker.compose.project.config_files
		com.docker.compose.project.working_dir
		com.docker.compose.service
		com.docker.compose.version
		com.docker.compose.container-number
	*/
	ComposeProjectLabel   = "com.docker.compose.project"
	ComposeWorkDirLabel   = "com.docker.compose.project.working_dir"
	ComposeConfFilesLabel = "com.docker.compose.project.config_files"

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
