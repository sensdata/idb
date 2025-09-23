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

type ScriptHistory struct {
	CommitHash    string    `json:"commit_hash"`
	Author        string    `json:"author"`
	Email         string    `json:"email"`
	Time          time.Time `json:"date"`
	CommitMessage string    `json:"commit_message"`
}

type ScriptHistoryList struct {
	Total int64            `json:"total"`
	Items []*ScriptHistory `json:"items"`
}

type ExecuteScript struct {
	ScriptPath string `json:"script_path" validate:"required"`
}

type ScriptExec struct {
	ScriptPath string `json:"script_path"`
	LogPath    string `json:"log_path"`
	Remove     bool   `json:"remove"`
}

type ScriptResult struct {
	LogHost uint      `json:"log_host"`
	LogPath string    `json:"log_path"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Out     string    `json:"out"`
	Err     string    `json:"err"`
}

type RunLogInfo struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
	IsHidden  bool   `json:"is_hidden"`
	CreatedAt string `json:"created_at"`
}

type RunLogList struct {
	Total int64         `json:"total"`
	Items []*RunLogInfo `json:"items"`
}

type RunLogDetail struct {
	Source    string `json:"source"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Content   string `json:"content"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
}
