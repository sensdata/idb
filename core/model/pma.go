package model

type GetServersRequest struct {
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GetServersResponse struct {
	Total   int              `json:"total"`
	Servers []*PmaServerInfo `json:"servers"`
}

type PmaServerInfo struct {
	Verbose string `json:"verbose"`
	Host    string `json:"host"`
	Port    string `json:"port"`
}

type AddServerRequest struct {
	Name    string `json:"name"`
	Verbose string `json:"verbose"`
	Host    string `json:"host"`
	Port    string `json:"port"`
}

type RemoveServerRequest struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port string `json:"port"`
}
