package model

// 系统配置
type SystemConfig struct {
	MaxWatchedFiles int `json:"maxWatchedFiles"` //最大监控文件数
	MaxOpenedFiles  int `json:"maxOpendFiles"`   //最大文件打开数
}
