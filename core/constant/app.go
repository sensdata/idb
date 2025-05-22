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

/*
	{
		"Name": "iDB_compose_name",
		"Label": "Compose Project Name",
		"Key": "Compose Name",
		"Type": "string",
		"Default": "__name__",
		"Required": true,
		"Hint": "Provide the name of the iDB instance.",
		"Options": null,
		"Validation": {
			"MinLength": 3,
			"MaxLength": 30,
			"Pattern": "^[a-zA-Z0-9_]+$",
			"MinValue": 0,
			"MaxValue": 0
		}
	},

	{
		"Name": "iDB_service_name",
		"Label": "Service Name",
		"Key": "Service Name",
		"Type": "string",
		"Default": "__mysql__",
		"Required": false,
		"Hint": "Specify the name of the database service.",
		"Options": null,
		"Validation": {
			"MinLength": 3,
			"MaxLength": 30,
			"Pattern": "^[a-zA-Z0-9_]+$",
			"MinValue": 0,
			"MaxValue": 0
		}
	},

	{
		"Name": "iDB_service_container_name",
		"Label": "Container Name",
		"Key": "Container Name",
		"Type": "string",
		"Default": "__iDB_service_container_name__",
		"Required": false,
		"Hint": "Enter the name for the container instance.",
		"Options": null,
		"Validation": {
			"MinLength": 3,
			"MaxLength": 30,
			"Pattern": "^[a-zA-Z0-9_]+$",
			"MinValue": 0,
			"MaxValue": 0
		}
	},

	{
		"Name": "iDB_service_mysql_root_password",
		"Label": "MySQL Root Password",
		"Key": "Root Password",
		"Type": "password",
		"Default": "__iDB_service_mysql_root_password__",
		"Required": true,
		"Hint": "Set a secure password for the MySQL root user.",
		"Options": null,
		"Validation": {
			"MinLength": 8,
			"MaxLength": 128,
			"Pattern": "^((?=.*[A-Z])(?=.*\\d)[\\w!@#$%^&*()\\-+=\\[\\]{}|;:'\",.<>?/\\\\~`]+)$",
			"MinValue": 0,
			"MaxValue": 0
		}
	},

	{
		"Name": "iDB_service_port",
		"Label": "Service Port",
		"Key": "Port",
		"Type": "number",
		"Default": "3306",
		"Required": true,
		"Hint": "Port for service access.",
		"Options": null,
		"Validation": {
			"MinLength": 0,
			"MaxLength": 0,
			"Pattern": "",
			"MinValue": 1024,
			"MaxValue": 65535
		}
	},

	{
		"Name": "iDB_service_network_mode",
		"Label": "Network Mode",
		"Key": "Network Mode",
		"Type": "select",
		"Default": "bridge",
		"Required": true,
		"Hint": "Choose the network mode for the service container.",
		"Options": [
			"bridge",
			"host",
			"none"
		],
		"Validation": {
			"MinLength": 0,
			"MaxLength": 0,
			"Pattern": "^(bridge|host|none)$",
			"MinValue": 0,
			"MaxValue": 0
		}
	},

	{
		"Name": "iDB_service_config_path",
		"Label": "Configuration Path",
		"Key": "Conf Path",
		"Type": "string",
		"Default": "./config/",
		"Required": true,
		"Hint": "Path for configuration files.",
		"Options": null,
		"Validation": null
	},

	{
		"Name": "iDB_service_data_path",
		"Label": "Data Storage Path",
		"Key": "Data Path",
		"Type": "string",
		"Default": "./data/",
		"Required": true,
		"Hint": "Directory path for data storage.",
		"Options": null,
		"Validation": null
	},

	{
		"Name": "iDB_service_log_path",
		"Label": "Log Storage Path",
		"Key": "Log Path",
		"Type": "string",
		"Default": "./log/",
		"Required": true,
		"Hint": "Directory path for storing logs.",
		"Options": null,
		"Validation": null
	},

	{
		"Name": "iDB_service_cert_path",
		"Label": "Certificate Path",
		"Key": "Cert Path",
		"Type": "string",
		"Default": "/some/idb/certmanager/",
		"Required": false,
		"Hint": "Path for certificates used by the container.",
		"Options": null,
		"Validation": null
	}
*/
const (
	IDB_compose_name           = "Compose Name"
	IDB_service_name           = "Service Name"
	IDB_service_container_name = "Container Name"
	IDB_service_port           = "Port"
	IDB_service_network_mode   = "Network Mode"
	IDB_service_config_path    = "Conf Path"
	IDB_service_data_path      = "Data Path"
	IDB_service_log_path       = "Log Path"
	IDB_service_cert_path      = "Cert Path"
)
