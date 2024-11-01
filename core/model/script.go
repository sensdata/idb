package model

import "time"

type Script struct {
	Script ScriptConfig `toml:"script"`
}

type ScriptConfig struct {
	DataPath string `toml:"data_path"`
	LogPath  string `toml:"log_path"`
}

type ScriptInfo struct {
	Source     string    `json:"source"`
	Type       string    `json:"type"`
	Category   string    `json:"category"`
	Name       string    `json:"name"`
	Extension  string    `json:"extension"`
	Content    string    `json:"content"`
	Size       int64     `json:"size"`
	UpdateTime time.Time `json:"update_time"`
	ModTime    time.Time `json:"mod_time"`
}

type QueryScript struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type GetScript struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
}

type CreateScript struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content"`
}

type UpdateScript struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Content  string `json:"content" validate:"required"`
}

type DeleteScript struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
}

type RestoreScript struct {
	HostID     uint   `json:"host_id" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Category   string `json:"category"`
	Name       string `json:"name" validate:"required"`
	CommitHash string `json:"commit_hash" validate:"required"`
}

type ScriptLog struct {
	HostID   uint   `json:"host_id" validate:"required"`
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ScriptDiff struct {
	HostID     uint   `json:"host_id" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Category   string `json:"category"`
	Name       string `json:"name" validate:"required"`
	CommitHash string `json:"commit_hash" validate:"required"`
}

type ExecuteScript struct {
	HostID     uint   `json:"host_id" validate:"required"`
	ScriptPath string `json:"script_path"`
}

type ScriptExec struct {
	ScriptPath string `json:"script_path"`
	LogPath    string `json:"log_path"`
}

type ScriptResult struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Out   string    `json:"out"`
	Err   string    `json:"err"`
}
