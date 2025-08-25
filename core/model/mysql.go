package model

type GetComposesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type GetComposesResponse struct {
	Total    int             `json:"total"`
	Composes []*ComposesInfo `json:"composes"`
}

type ComposesInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Port    string `json:"port"`
}

type OperateRequest struct {
	Name      string `json:"name"`
	Operation string `json:"operation" validate:"required,oneof=start stop restart"`
}

type SetPortRequest struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

type GetConfRequest struct {
	Name string `json:"name"`
}

type GetConfResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type SetConfRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
