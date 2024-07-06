package utils

import (
	"fmt"
	"os"
	"path"
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
