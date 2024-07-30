package plugin

import "github.com/gin-gonic/gin"

// Plugin的路由
type PluginRoute struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}

// PluginInfo 定义插件的元数据
type PluginInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	Entry       string `yaml:"sysinfo"`
}

// MenuItem 定义菜单项
type MenuItem struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Config 定义插件配置结构
type PluginConfig struct {
	Plugin PluginInfo `yaml:"plugin"`
	Menu   []MenuItem `yaml:"menu"`
}
