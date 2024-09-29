package model

type ServiceBootReq struct {
	HostID      uint   `json:"host_id"`
	Command     string `json:"command"`
	ServiceName string `json:"service_name"`
}
