package manager

import (
	_ "embed"
	"fmt"
	"log"
	"os/exec"
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
	InitCallback()
	LoadAll() error
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

func Initialize(router *gin.Engine) error {
	reg, err := LoadRegistry(registryData)
	if err != nil {
		return err
	}
	pm := &DefaultManager{
		router:   router,
		registry: reg,
		plugins:  make(map[string]*PluginInstance),
	}
	pm.InitCallback()
	if err := pm.LoadAll(); err != nil {
		global.LOG.Error("Failed to load plugins: %v", err)
		return err
	}

	PluginMan = pm
	return nil
}

func (m *DefaultManager) InitCallback() {
	m.callback = &callback.CallbackServer{}
	m.broker = &hplugin.GRPCBroker{}
	m.brokerID = m.broker.NextId()
	go func() {
		m.broker.AcceptAndServe(m.brokerID, func(opts []grpc.ServerOption) *grpc.Server {
			s := grpc.NewServer(opts...)
			cbpb.RegisterCenterCallbackServer(s, m.callback)
			return s
		})
	}()
}

// LoadAll 启动所有 enabled 插件，建立 gRPC 连接
func (m *DefaultManager) LoadAll() error {
	global.LOG.Info("loading plugins")
	for _, entry := range m.registry.Plugins {
		if !entry.Enabled {
			global.LOG.Info("plugin %s is disabled", entry.Name)
			continue
		}

		factory, ok := PluginFactories[entry.Name]
		if !ok {
			global.LOG.Info("loaplugin %s has no factory registered", entry.Name)
			return fmt.Errorf("plugin %s has no factory registered", entry.Name)
		}

		// 日志提示
		global.LOG.Info("loading plugin %s (%s)", entry.Name, entry.Path)

		// 启动插件子进程
		client := hplugin.NewClient(&hplugin.ClientConfig{
			HandshakeConfig: hplugin.HandshakeConfig{
				ProtocolVersion:  1,
				MagicCookieKey:   "PLUGIN_NAME",
				MagicCookieValue: entry.Name,
			},
			Cmd:              exec.Command(entry.Path),
			AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolGRPC},
			Plugins: map[string]hplugin.Plugin{
				"grpc": &GRPCPlugin{NewClient: factory},
			},
		})

		rpcClient, err := client.Client()
		if err != nil {
			log.Printf("[plugin] failed to connect: %s: %v", entry.Name, err)
			client.Kill()
			continue
		}

		// 获取 grpc 插件接口
		disp, err := rpcClient.Dispense("grpc")
		if err != nil {
			log.Printf("[plugin] failed to dispense: %s: %v", entry.Name, err)
			client.Kill()
			continue
		}

		m.mu.Lock()
		m.plugins[entry.Name] = &PluginInstance{
			Name:   entry.Name,
			Client: client,
			Stub:   disp,
		}
		m.mu.Unlock()

		log.Printf("[plugin] loaded plugin: %s", entry.Name)
	}

	return nil
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
