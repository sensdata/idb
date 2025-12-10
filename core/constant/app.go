package constant

const (
	Running     = "Running"
	UnHealthy   = "UnHealthy"
	Error       = "Error"
	Stopped     = "Stopped"
	Installing  = "Installing"
	DownloadErr = "DownloadErr"
	Upgrading   = "Upgrading"
	UpgradeErr  = "UpgradeErr"
	Rebuilding  = "Rebuilding"
	Syncing     = "Syncing"
	SyncSuccess = "SyncSuccess"
	Paused      = "Paused"
	SyncErr     = "SyncErr"

	ContainerPrefix = "1Panel-"

	AppNormal   = "Normal"
	AppTakeDown = "TakeDown"

	AppOpenresty  = "openresty"
	AppMysql      = "mysql"
	AppMariaDB    = "mariadb"
	AppPostgresql = "postgresql"
	AppRedis      = "redis"
	AppPostgres   = "postgres"
	AppMongodb    = "mongodb"
	AppMemcached  = "memcached"

	AppResourceLocal  = "local"
	AppResourceRemote = "remote"

	CPUS          = "CPUS"
	MemoryLimit   = "MEMORY_LIMIT"
	HostIP        = "HOST_IP"
	ContainerName = "CONTAINER_NAME"
)

type AppOperate string

var (
	Start   AppOperate = "start"
	Stop    AppOperate = "stop"
	Restart AppOperate = "restart"
	Delete  AppOperate = "delete"
	Sync    AppOperate = "sync"
	Backup  AppOperate = "backup"
	Update  AppOperate = "update"
	Rebuild AppOperate = "rebuild"
	Upgrade AppOperate = "upgrade"
	Reload  AppOperate = "reload"
)

const (
	IDB_compose_name           = "iDB_compose_name"
	IDB_service_name           = "iDB_service_name"
	IDB_service_container_name = "iDB_service_container_name"
	IDB_service_port           = "iDB_service_port"
	IDB_service_network_mode   = "iDB_service_network_mode"
	IDB_service_config_path    = "iDB_service_config_path"
	IDB_service_data_path      = "iDB_service_data_path"
	IDB_service_log_path       = "iDB_service_log_path"
	IDB_service_cert_path      = "iDB_service_cert_path"
	IDB_service_assets_path    = "iDB_service_assets_path"
)
