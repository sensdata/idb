package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func EnsurePaths(paths []string) error {
	for _, path := range paths {
		if err := checkDir(path); err != nil {
			return fmt.Errorf("failed to ensure path %s: %v", path, err)
		}
	}
	return nil
}

func checkDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}
	return nil
}

func EnsureFile(filePath string) error {
	if err := checkFile(filePath); err != nil {
		return err
	}
	return nil
}

func checkFile(filePath string) error {
	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// 如果文件不存在，创建文件
		// 确保文件的目录存在
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		// 创建文件
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		return nil // 成功创建文件后返回 nil
	}
	if err != nil {
		// 返回任何其他错误
		return err
	}
	return nil // 文件已存在，返回 nil
}
