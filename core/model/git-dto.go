package model

type QueryGitFile struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GetGitFileDetail struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
}

type CreateGitCategory struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
}

type UpdateGitCategory struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	NewName  string `json:"new_name"`
}

type DeleteGitCategory struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
}

type CreateGitFile struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content"`
}

type UpdateGitFile struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	NewName  string `json:"new_name"`
	Content  string `json:"content" validate:"required"`
}

type DeleteGitFile struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
}

type RestoreGitFile struct {
	Type       string `json:"type" validate:"required,oneof=global local"`
	Category   string `json:"category"`
	Name       string `json:"name" validate:"required"`
	CommitHash string `json:"commit_hash" validate:"required"`
}

type GitFileLog struct {
	Type     string `json:"type" validate:"required,oneof=global local"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GitFileDiff struct {
	Type       string `json:"type" validate:"required,oneof=global local"`
	Category   string `json:"category"`
	Name       string `json:"name" validate:"required"`
	CommitHash string `json:"commit_hash" validate:"required"`
}
