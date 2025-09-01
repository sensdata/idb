package model

type GetComposesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type GetComposesResponse struct {
	Total    int             `json:"total"`
	Composes []*ComposeBrief `json:"composes"`
}

type ComposeBrief struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Port    string `json:"port"`
	Status  string `json:"status"`
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

type GetRemoteAccessRequest struct {
	Name string `json:"name"`
}

type GetRemoteAccessResponse struct {
	RemoteAccess bool `json:"remote_access"`
}

type SetRemoteAccessRequest struct {
	Name         string `json:"name"`
	RemoteAccess bool   `json:"remote_access"`
}

type GetRootPasswordRequest struct {
	Name string `json:"name"`
}

type GetRootPasswordResponse struct {
	Password string `json:"password"`
}

type SetRootPasswordRequest struct {
	Name    string `json:"name"`
	NewPass string `json:"new_pass"`
}
