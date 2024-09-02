package model

type ServiceBootReq struct {
	HostID      uint   `json:"hostId"`
	Command     string `json:"command"`
	ServiceName string `json:"serviceName"`
}
