package model

type GetServersRequest struct {
	Name     string `form:"name" json:"name"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
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
	Name    string `form:"name" json:"name"`
	Verbose string `form:"verbose" json:"verbose"`
	Host    string `form:"host" json:"host"`
	Port    string `form:"port" json:"port"`
}

type RemoveServerRequest struct {
	Name string `form:"name" json:"name"`
	Host string `form:"host" json:"host"`
	Port string `form:"port" json:"port"`
}
