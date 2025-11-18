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
