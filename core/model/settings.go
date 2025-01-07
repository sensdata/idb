package model

type About struct {
	Version string `json:"version"`
}

type SettingInfo struct {
	MonitorIP     string `json:"monitor_ip"`
	ServerPort    int    `json:"server_port"`
	BindDomain    string `json:"bind_domain"`
	Https         string `json:"https"`
	HttpsCertType string `json:"https_cert_type"`
	HttpsCertPath string `json:"https_cert_path"`
	HttpsKeyPath  string `json:"https_key_path"`
}

type UpdateSettingRequest struct {
	MonitorIP     string `json:"monitor_ip" validate:"required"`
	ServerPort    int    `json:"server_port" validate:"required"`
	BindDomain    string `json:"bind_domain,omitempty"`
	Https         string `json:"https" validate:"required,oneof=no yes"`
	HttpsCertType string `json:"https_cert_type,omitempty"`
	HttpsCertPath string `json:"https_cert_path,omitempty"`
	HttpsKeyPath  string `json:"https_key_path,omitempty"`
}
