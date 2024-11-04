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
}

// MenuItem 定义菜单项
type MenuItem struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Plugin 定义插件结构
type Plugin struct {
	Info PluginInfo `yaml:"plugin"`
	Menu []MenuItem `yaml:"menu"`
}

// PluginConf
type PluginConf struct {
	WorkDir string `yaml:"work_dir"`
	LogDir  string `yaml:"log_dir"`
}
