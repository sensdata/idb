package model

import "time"

type GitFile struct {
	Source       string    `json:"source"`
	RepoPath     string    `json:"repo_path"`
	RelativePath string    `json:"relative_path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Content      string    `json:"content"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
}
