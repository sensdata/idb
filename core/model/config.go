package model

// 系统配置
type SystemConfig struct {
	MaxWatchedFiles int `json:"max_watched_files"` //最大监控文件数
	MaxOpenedFiles  int `json:"max_opend_files"`   //最大文件打开数
}
