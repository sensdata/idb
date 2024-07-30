package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYaml 读取和解析 YAML 配置文件
func LoadYaml(filePath string, config interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}
