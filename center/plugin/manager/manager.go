package manager

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"

	"github.com/gin-gonic/gin"
	hplugin "github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/center/plugin/manager/callback"
	cbpb "github.com/sensdata/idb/center/plugin/manager/callback/pb"
	"google.golang.org/grpc"
)

type PluginInstance struct {
	Name   string
	Client *hplugin.Client
	Stub   interface{} // 通用 gRPC 客户端接口（经过 wrapper）
}

type PluginManager interface {
	Initialize(router *gin.Engine) error
	GetPlugin(name string) (*PluginInstance, error)
	IsAuthorized(name string) bool
	ShutdownAll()
}

type DefaultManager struct {
	router   *gin.Engine
	registry *Registry
	plugins  map[string]*PluginInstance
	mu       sync.RWMutex
	callback *callback.CallbackServer // 统一的回调实现
	broker   *hplugin.GRPCBroker      // broker 实例
	brokerID uint32                   // 回调 broker ID
}

var PluginMan PluginManager

//go:embed registry.yaml
var registryData []byte

func NewPluginManager() PluginManager {
	return &DefaultManager{
		plugins: make(map[string]*PluginInstance),
	}
}

func (m *DefaultManager) Initialize(router *gin.Engine) error {
	global.LOG.Info("initializing plugin manager")
	reg, err := LoadRegistry(registryData)
	if err != nil {
		global.LOG.Error("failed to load registry: %v", err)
		return err
	}
	m.registry = reg
	global.LOG.Info("registry loaded: %v", m.registry)
	m.InitCallback()
	go m.LoadAllAsync() // 异步加载插件
	return nil
}

// 初始化callback服务
func (m *DefaultManager) InitCallback() {
	global.LOG.Info("initializing callback server")
	m.callback = &callback.CallbackServer{}
	m.broker = &hplugin.GRPCBroker{}
	m.brokerID = m.broker.NextId()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.LOG.Error("panic in callback server: %v", r)
			}
		}()
		global.LOG.Info("starting callback server")
		m.broker.AcceptAndServe(m.brokerID, func(opts []grpc.ServerOption) *grpc.Server {
			defer func() {
				if r := recover(); r != nil {
					global.LOG.Error("panic in callback server: %v, stack: %s", r, debug.Stack())
				}
			}()
			s := grpc.NewServer(opts...)
			cbpb.RegisterCenterCallbackServer(s, m.callback)
			return s
		})
	}()
	global.LOG.Info("initializing callback end")
}

// GetPlugin 返回已加载的插件实例
func (m *DefaultManager) GetPlugin(name string) (*PluginInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.plugins[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("plugin %s not loaded", name)
}

// IsAuthorized 判断插件是否启用
func (m *DefaultManager) IsAuthorized(name string) bool {
	for _, p := range m.registry.Plugins {
		if p.Name == name {
			// TODO: 更多的鉴权逻辑 key
			return p.Enabled
		}
	}
	return false
}

// ShutdownAll 停止所有插件子进程
func (m *DefaultManager) ShutdownAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, p := range m.plugins {
		log.Printf("[plugin] shutting down plugin: %s", name)
		p.Client.Kill()
	}
	m.plugins = map[string]*PluginInstance{}
}

// LoadAllAsync 异步加载插件，避免阻塞初始化流程
func (m *DefaultManager) LoadAllAsync() {
	global.LOG.Info("starting plugin async load...")
	for _, entry := range m.registry.Plugins {
		go func(e PluginEntry) {
			if !e.Enabled {
				global.LOG.Info("plugin %s is disabled", e.Name)
				return
			}

			if !m.IsPluginInstalled(e) {
				global.LOG.Info("plugin %s not installed, installing...", e.Name)
				if err := m.InstallPlugin(e); err != nil {
					global.LOG.Error("failed to install plugin %s: %v", e.Name, err)
					return
				}
			}

			if err := m.LoadPlugin(e); err != nil {
				global.LOG.Error("failed to load plugin %s: %v", e.Name, err)
				return
			}

			global.LOG.Info("plugin %s loaded successfully", e.Name)
		}(entry)
	}
}

// IsPluginInstalled 判断插件文件是否存在
func (m *DefaultManager) IsPluginInstalled(e PluginEntry) bool {
	info, err := os.Stat(e.Path)
	return err == nil && !info.IsDir()
}

// InstallPlugin 安装插件（暂为占位，后续支持下载解压）
func (m *DefaultManager) InstallPlugin(e PluginEntry) error {
	return fmt.Errorf("auto install not implemented for plugin %s", e.Name)
}

// LoadPlugin 启动并注册插件
func (m *DefaultManager) LoadPlugin(entry PluginEntry) error {
	factory, ok := PluginFactories[entry.Name]
	if !ok {
		return fmt.Errorf("plugin %s has no factory registered", entry.Name)
	}

	cmd := exec.Command(entry.Path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	client := hplugin.NewClient(&hplugin.ClientConfig{
		HandshakeConfig: hplugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "PLUGIN_NAME",
			MagicCookieValue: entry.Name,
		},
		Cmd:              cmd,
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolGRPC},
		Plugins: map[string]hplugin.Plugin{
			"grpc": &GRPCPlugin{NewClient: factory},
		},
		Managed: true,
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to connect to plugin %s: %w", entry.Name, err)
	}

	disp, err := rpcClient.Dispense("grpc")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin %s: %w", entry.Name, err)
	}

	m.mu.Lock()
	m.plugins[entry.Name] = &PluginInstance{
		Name:   entry.Name,
		Client: client,
		Stub:   disp,
	}
	m.mu.Unlock()

	return nil
}
