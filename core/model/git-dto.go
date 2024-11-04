package model

type QueryGitFile struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GetGitFileDetail struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
}

type CreateGitFile struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content"`
}

type UpdateGitFile struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content" validate:"required"`
}

type DeleteGitFile struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
}

type RestoreGitFile struct {
	HostID     uint   `json:"host_id" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Category   string `json:"category"`
	Name       string `json:"name" validate:"required"`
	CommitHash string `json:"commit_hash" validate:"required"`
}

type GitFileLog struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GitFileDiff struct {
	HostID     uint   `json:"host_id" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Category   string `json:"category"`
	Name       string `json:"name" validate:"required"`
	CommitHash string `json:"commit_hash" validate:"required"`
}
