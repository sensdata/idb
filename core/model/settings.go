package model

type Profile struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type BindIp struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

type AvailableIps struct {
	IPs []BindIp `json:"ips"`
}

type About struct {
	Version    string `json:"version"`
	NewVersion string `json:"new_version"`
}

type SettingInfo struct {
	BindIP        string `json:"bind_ip"`
	BindPort      int    `json:"bind_port"`
	BindDomain    string `json:"bind_domain"`
	Https         string `json:"https"`
	HttpsCertType string `json:"https_cert_type"`
	HttpsCertPath string `json:"https_cert_path"`
	HttpsCertData string `json:"https_cert_data"`
	HttpsKeyPath  string `json:"https_key_path"`
	HttpsKeyData  string `json:"https_key_data"`
}

type UpdateSettingRequest struct {
	BindIP        string `json:"bind_ip" validate:"required"`
	BindPort      int    `json:"bind_port" validate:"required"`
	BindDomain    string `json:"bind_domain,omitempty"`
	Https         string `json:"https" validate:"required,oneof=no yes"`
	HttpsCertType string `json:"https_cert_type,omitempty"`
	HttpsCertPath string `json:"https_cert_path,omitempty"`
	HttpsKeyPath  string `json:"https_key_path,omitempty"`
}

type UpdateSettingResponse struct {
	RedirectUrl string `json:"redirect_url"`
}

type SettingsStatus struct {
	CollectedAt int64                `json:"collected_at"`
	Center      CenterRuntimeStatus  `json:"center"`
	Agent       *AgentRuntimeStatus  `json:"agent,omitempty"`
	Checks      SettingsStatusChecks `json:"checks"`
}

type CenterRuntimeStatus struct {
	PID          int     `json:"pid"`
	Version      string  `json:"version"`
	BindIP       string  `json:"bind_ip"`
	BindPort     int     `json:"bind_port"`
	BindDomain   string  `json:"bind_domain"`
	HttpsEnabled bool    `json:"https_enabled"`
	AccessURL    string  `json:"access_url"`
	StartedAt    int64   `json:"started_at"`
	Uptime       int64   `json:"uptime"`
	CPUPercent   float64 `json:"cpu_percent"`
	CPUSeconds   float64 `json:"cpu_seconds"`
	MemoryRSS    uint64  `json:"memory_rss"`
	HeapAlloc    uint64  `json:"heap_alloc"`
	HeapSys      uint64  `json:"heap_sys"`
	StackInuse   uint64  `json:"stack_inuse"`
	Goroutines   int     `json:"goroutines"`
	OpenFDs      int     `json:"open_fds"`
}

type AgentRuntimeStatus struct {
	HostID             uint    `json:"host_id"`
	HostName           string  `json:"host_name"`
	HostAddr           string  `json:"host_addr"`
	AgentAddr          string  `json:"agent_addr"`
	AgentPort          int     `json:"agent_port"`
	AgentVersion       string  `json:"agent_version"`
	Installed          string  `json:"installed"`
	Connected          string  `json:"connected"`
	LastHeartbeat      int64   `json:"last_heartbeat"`
	CPU                float64 `json:"cpu"`
	Memory             float64 `json:"memory"`
	MemTotal           string  `json:"mem_total"`
	MemUsed            string  `json:"mem_used"`
	Disk               float64 `json:"disk"`
	BootTime           string  `json:"boot_time"`
	RunTime            int64   `json:"run_time"`
	ProcessRSS         uint64  `json:"process_rss"`
	HeapAlloc          uint64  `json:"heap_alloc"`
	HeapSys            uint64  `json:"heap_sys"`
	StackInuse         uint64  `json:"stack_inuse"`
	Goroutines         int     `json:"goroutines"`
	OpenFDs            int     `json:"open_fds"`
	ActiveSessions     int     `json:"active_sessions"`
	ActiveLogFollowers int     `json:"active_log_followers"`
}

type SettingsStatusChecks struct {
	DatabaseConnected  bool `json:"database_connected"`
	DefaultHostFound   bool `json:"default_host_found"`
	DefaultAgentOnline bool `json:"default_agent_online"`
	DefaultAgentFresh  bool `json:"default_agent_fresh"`
}
