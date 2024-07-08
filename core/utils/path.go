package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func EnsurePaths(basePath string) error {
	if err := checkDir(basePath); err != nil {
		return err
	}
	return nil
}

func checkDir(dirPath string) error {
	dir := path.Dir(dirPath)
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", dir, err)
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
		file.Close()
	}
	return err
}
