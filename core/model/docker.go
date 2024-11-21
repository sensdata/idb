package model

import "time"

// docker

type LogOption struct {
	LogMaxSize string `json:"log_max_size"`
	LogMaxFile string `json:"log_max_file"`
}

type Ipv6Option struct {
	FixedCidrV6  string `json:"fixed_cidr_v6"`
	Ip6Tables    bool   `json:"ip6_tables" validate:"required"`
	Experimental bool   `json:"experimental"`
}

type DaemonJsonUpdateByFile struct {
	File string `json:"file"`
}

type DaemonJsonConf struct {
	IsSwarm      bool     `json:"is_swarm"`
	Status       string   `json:"status"`
	Version      string   `json:"version"`
	Mirrors      []string `json:"registry_mirrors"`
	Registries   []string `json:"insecure_registries"`
	LiveRestore  bool     `json:"live_restore"`
	IPTables     bool     `json:"ip_tables"`
	CgroupDriver string   `json:"cgroup_driver"`

	Ipv6         bool   `json:"ipv6"`
	FixedCidrV6  string `json:"fixed_cidr_v6"`
	Ip6Tables    bool   `json:"ip6_tables"`
	Experimental bool   `json:"experimental"`

	LogMaxSize string `json:"log_max_size"`
	LogMaxFile string `json:"log_max_file"`
}

type DockerOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start restart stop"`
}

// common
type Inspect struct {
	ID   string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required,oneof=container image volume network"`
}

type Prune struct {
	PruneType  string `json:"type" validate:"required,oneof=container image volume network buildcache"`
	WithTagAll bool   `json:"with_tag_all"`
}

type PruneResult struct {
	DeletedNumber  int `json:"deleted_number"`
	SpaceReclaimed int `json:"space_reclaimed"`
}

// container
type QueryContainer struct {
	PageInfo
	Name    string `json:"name"`
	State   string `json:"state" validate:"required,oneof=all created running paused restarting removing exited dead"`
	OrderBy string `json:"order_by"`
	Order   string `json:"order"`
}

type ContainerInfo struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	ImageId     string `json:"image_id"`
	ImageName   string `json:"image_name"`
	CreateTime  string `json:"create_time"`
	State       string `json:"state"`
	RunTime     string `json:"run_time"`

	Network []string `json:"network"`
	Ports   []string `json:"ports"`

	From string `json:"from"`

	// AppName        string   `json:"app_name"`
	// AppInstallName string   `json:"app_install_name"`
	// Websites       []string `json:"websites"`
}

type VolumeHelper struct {
	Type         string `json:"type"`
	SourceDir    string `json:"source_dir"`
	ContainerDir string `json:"container_dir"`
	Mode         string `json:"mode"`
}
type PortHelper struct {
	HostIP        string `json:"host_ip"`
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type ContainerOperate struct {
	ContainerID     string         `json:"container_id"`
	ForcePull       bool           `json:"force_pull"`
	Name            string         `json:"name" validate:"required"`
	Image           string         `json:"image" validate:"required"`
	Network         string         `json:"network"`
	Ipv4            string         `json:"ipv4"`
	Ipv6            string         `json:"ipv6"`
	PublishAllPorts bool           `json:"publish_all_ports"`
	ExposedPorts    []PortHelper   `json:"exposed_ports"`
	Tty             bool           `json:"tty"`
	OpenStdin       bool           `json:"open_stdin"`
	Cmd             []string       `json:"cmd"`
	Entrypoint      []string       `json:"entry_point"`
	CPUShares       int64          `json:"cpu_shares"`
	NanoCPUs        float64        `json:"nano_cpus"`
	Memory          float64        `json:"memory"`
	Privileged      bool           `json:"privileged"`
	AutoRemove      bool           `json:"auto_remove"`
	Volumes         []VolumeHelper `json:"volumes"`
	Labels          []string       `json:"labels"`
	Env             []string       `json:"env"`
	RestartPolicy   string         `json:"restart_policy"`
}

type ContainerUpgrade struct {
	Name      string `json:"name" validate:"required"`
	Image     string `json:"image" validate:"required"`
	ForcePull bool   `json:"force_pull"`
}

type ContainerOperation struct {
	Names     []string `json:"names" validate:"required"`
	Operation string   `json:"operation" validate:"required,oneof=start stop restart kill pause resume remove"`
}

type ContainerResourceUsage struct {
	ContainerID string `json:"container_id"`

	CPUTotalUsage uint64  `json:"cpu_total_usage"`
	SystemUsage   uint64  `json:"system_usage"`
	CPUPercent    float64 `json:"cpu_percent"`
	PercpuUsage   int     `json:"per_cpu_usage"`

	MemoryCache   uint64  `json:"memory_cache"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
}

type ContainerStats struct {
	CPUPercent float64 `json:"cpu_percent"`
	Memory     float64 `json:"memory"`
	Cache      float64 `json:"cache"`
	IORead     float64 `json:"io_read"`
	IOWrite    float64 `json:"io_write"`
	NetworkRX  float64 `json:"network_rx"`
	NetworkTX  float64 `json:"network_tx"`

	ShotTime time.Time `json:"shot_time"`
}

type ContainerResourceLimit struct {
	CPU    int    `json:"cpu"`
	Memory uint64 `json:"memory"`
}

// compose
type ComposeInfo struct {
	Name            string             `json:"name"`
	CreatedAt       string             `json:"created_at"`
	CreatedBy       string             `json:"created_by"`
	ContainerNumber int                `json:"container_number"`
	ConfigFile      string             `json:"config_file"`
	Workdir         string             `json:"work_dir"`
	Path            string             `json:"path"`
	Containers      []ComposeContainer `json:"containers"`
}

type ComposeContainer struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	CreateTime  string `json:"create_time"`
	State       string `json:"state"`
}

type ComposeCreate struct {
	Name     string `json:"name"`
	From     string `json:"from" validate:"required,oneof=edit path template"`
	File     string `json:"file"`
	Path     string `json:"path"`
	Template uint   `json:"template"`
}

type ComposeOperation struct {
	Name      string `json:"name" validate:"required"`
	Path      string `json:"path" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=start stop down"`
	WithFile  bool   `json:"with_file"`
}

type ComposeUpdate struct {
	Name    string `json:"name" validate:"required"`
	Path    string `json:"path" validate:"required"`
	Content string `json:"content" validate:"required"`
}

// image
type Image struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	IsUsed    bool      `json:"is_used"`
	Tags      []string  `json:"tags"`
	Size      string    `json:"size"`
}

type ImageLoad struct {
	Path string `json:"path" validate:"required"`
}

type ImageBuild struct {
	From       string   `json:"from" validate:"required"`
	Name       string   `json:"name" validate:"required"`
	Dockerfile string   `json:"docker_file" validate:"required"`
	Tags       []string `json:"tags"`
}

type ImagePull struct {
	ImageName string `json:"image_name" validate:"required"`
}

type ImageTag struct {
	SourceID   string `json:"source_id" validate:"required"`
	TargetName string `json:"target_name" validate:"required"`
}

type ImagePush struct {
	TagName string `json:"tag_name" validate:"required"`
	Name    string `json:"name" validate:"required"`
}

type ImageSave struct {
	TagName string `json:"tag_name" validate:"required"`
	Path    string `json:"path" validate:"required"`
	Name    string `json:"name" validate:"required"`
}

// volume
type Volume struct {
	Name       string    `json:"name"`
	Labels     []string  `json:"labels"`
	Driver     string    `json:"driver"`
	Mountpoint string    `json:"mount_point"`
	CreatedAt  time.Time `json:"created_at"`
}

type VolumeCreate struct {
	Name    string   `json:"name" validate:"required"`
	Driver  string   `json:"driver" validate:"required"`
	Options []string `json:"options"`
	Labels  []string `json:"labels"`
}

// network
type Network struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Labels     []string  `json:"labels"`
	Driver     string    `json:"driver"`
	IPAMDriver string    `json:"ipam_driver"`
	Subnet     string    `json:"subnet"`
	Gateway    string    `json:"gateway"`
	CreatedAt  time.Time `json:"created_at"`
	Attachable bool      `json:"attachable"`
}

type NetworkCreate struct {
	Name       string     `json:"name" validate:"required"`
	Driver     string     `json:"driver" validate:"required"`
	Options    []string   `json:"options"`
	Ipv4       bool       `json:"ipv4"`
	Subnet     string     `json:"subnet"`
	Gateway    string     `json:"gateway"`
	IPRange    string     `json:"ip_range"`
	AuxAddress []KeyValue `json:"aux_address"`

	Ipv6         bool       `json:"ipv6"`
	SubnetV6     string     `json:"subnet_v6"`
	GatewayV6    string     `json:"gateway_v6"`
	IPRangeV6    string     `json:"ip_range_v6"`
	AuxAddressV6 []KeyValue `json:"aux_address_v6"`
	Labels       []string   `json:"labels"`
}
