package plugin

import hplugin "github.com/hashicorp/go-plugin"

type PluginInstance struct {
	Name   string
	Client *hplugin.Client
	Stub   interface{} // 通用 gRPC 客户端接口（经过 wrapper）
}

type Registry struct {
	Plugins []PluginEntry `yaml:"plugins"`
}

type PluginEntry struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Url     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}
