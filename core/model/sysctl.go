package model

type ServiceBootReq struct {
	Command     string `json:"command"`
	ServiceName string `json:"service_name"`
}
