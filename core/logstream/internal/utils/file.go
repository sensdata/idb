package utils

import (
	"os"
	"path/filepath"
	"time"
)

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CleanOldFiles 清理指定目录下的旧文件
func CleanOldFiles(dir string, maxAge int64) error {
	now := time.Now()
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && now.Sub(info.ModTime()).Seconds() > float64(maxAge) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	})
}
