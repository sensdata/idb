package model

import "time"

type ScriptInfo struct {
	Source    string `json:"source"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
	Linked    bool   `json:"linked"`
}

type ScriptList struct {
	Total int64         `json:"total"`
	Items []*ScriptInfo `json:"items"`
}

type ExecuteScript struct {
	ScriptPath string `json:"script_path" validate:"required"`
}

type ScriptExec struct {
	ScriptPath string `json:"script_path"`
	LogPath    string `json:"log_path"`
}

type ScriptResult struct {
	LogHost uint      `json:"log_host"`
	LogPath string    `json:"log_path"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Out     string    `json:"out"`
	Err     string    `json:"err"`
}
