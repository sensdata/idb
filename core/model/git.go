package model

import "time"

type GitFile struct {
	Source    string    `json:"source"`
	Name      string    `json:"name"`
	Extension string    `json:"extension"`
	Content   string    `json:"content"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
}

type GitCommit struct {
	CommitHash string    `json:"commit_hash"`
	Author     string    `json:"author"`
	Email      string    `json:"email"`
	Time       time.Time `json:"date"`
	Message    string    `json:"message"`
}

type GitInit struct {
	HostID   uint   `json:"host_id" validate:"required"`
	RepoPath string `json:"repo_path" validate:"required"`
	IsBare   bool   `json:"is_bare" validate:"required"`
}

type GitSync struct {
	HostID    uint   `json:"host_id" validate:"required"`
	RemoteUrl string `json:"remote_url" validate:"required"`
	RepoPath  string `json:"repo_path" validate:"required"`
}

type GitCheck struct {
	HostID   uint   `json:"host_id" validate:"required"`
	RepoPath string `json:"repo_path" validate:"required"`
}

type GitQuery struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	Extension    string `json:"extension"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

type GitGetFile struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
}

type GitCreate struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	Dir          bool   `json:"dir" validate:"required"`
	Content      string `json:"content" validate:"required"`
}

type GitUpdate struct {
	HostID          uint   `json:"host_id" validate:"required"`
	RepoPath        string `json:"repo_path" validate:"required"`
	RelativePath    string `json:"relative_path" validate:"required"`
	NewRelativePath string `json:"new_relative_path"`
	Dir             bool   `json:"dir" validate:"required"`
	Content         string `json:"content" validate:"required"`
}

type GitDelete struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	Dir          bool   `json:"dir" validate:"required"`
}

type GitRestore struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	CommitHash   string `json:"commit_hash" validate:"required"`
}

type GitLog struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

type GitDiff struct {
	HostID       uint   `json:"host_id" validate:"required"`
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	CommitHash   string `json:"commit_hash" validate:"required"`
}
