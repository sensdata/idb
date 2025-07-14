package plugin

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/callback"
	cbpb "github.com/sensdata/idb/center/core/plugin/callback/pb"
	factory "github.com/sensdata/idb/center/core/plugin/factory"
	"github.com/sensdata/idb/center/global"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type IPluginService interface {
	Start() error
	Stop() error
	GetPlugin(name string) (*PluginInstance, error)
}

type PluginServer struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	registry *Registry
	plugins  map[string]*PluginInstance
	mu       sync.RWMutex

	callback   *callback.CallbackServer
	broker     *hplugin.GRPCBroker
	brokerID   uint32
	grpcServer *grpc.Server
}

func NewPluginService() IPluginService {
	ctx, cancel := context.WithCancel(context.Background())
	return &PluginServer{
		ctx:     ctx,
		cancel:  cancel,
		plugins: make(map[string]*PluginInstance),
	}
}

var PLUGINSERVER IPluginService = NewPluginService()

//go:embed registry.yaml
var registryData []byte

func (s *PluginServer) Start() error {
	global.LOG.Info("plugin server start")

	r, err := loadRegistry(registryData)
	if err != nil {
		global.LOG.Error("failed to load registry: %v", err)
		return err
	}
	s.registry = r

	s.wg.Add(1)
	go s.loadPlugins()

	s.wg.Add(1)
	go s.initCallback()

	return nil
}

func (s *PluginServer) Stop() error {
	global.LOG.Info("stopping plugin server")

	s.cancel() // 通知所有 goroutine 退出
	s.wg.Wait()

	s.stopPlugins()

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	global.LOG.Info("plugin server stopped")
	return nil
}

func (s *PluginServer) GetPlugin(name string) (*PluginInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.plugins[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("plugin %s not loaded", name)
}

func loadRegistry(data []byte) (*Registry, error) {
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry yaml: %w", err)
	}

	return &reg, nil
}

func (s *PluginServer) loadPlugins() {
	defer s.wg.Done()
	global.LOG.Info("loading plugins")
	for _, entry := range s.registry.Plugins {
		select {
		case <-s.ctx.Done():
			global.LOG.Info("loadPlugins cancelled")
			return
		default:
		}

		if !entry.Enabled {
			global.LOG.Info("plugin %s is disabled", entry.Name)
			return
		}

		s.wg.Add(1)
		go func(e PluginEntry) {
			defer s.wg.Done()

			if !s.isPluginInstalled(e) {
				global.LOG.Info("plugin %s not installed, installing...", e.Name)
				if err := s.installPlugin(e); err != nil {
					global.LOG.Error("failed to install plugin %s: %v", e.Name, err)
					return
				}
			}

			if err := s.loadPlugin(e); err != nil {
				global.LOG.Error("failed to load plugin %s: %v", e.Name, err)
				return
			}

			global.LOG.Info("plugin %s loaded successfully", e.Name)
		}(entry)
	}
}

func (s *PluginServer) stopPlugins() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, p := range s.plugins {
		global.LOG.Info("shutting down plugin: %s", name)
		p.Client.Kill()
	}
	s.plugins = map[string]*PluginInstance{}
}

func (s *PluginServer) isPluginInstalled(e PluginEntry) bool {
	info, err := os.Stat(e.Path)
	return err == nil && !info.IsDir()
}

func (s *PluginServer) installPlugin(e PluginEntry) error {
	return fmt.Errorf("auto install not implemented for plugin %s", e.Name)
}

func (s *PluginServer) loadPlugin(entry PluginEntry) error {
	factory, ok := factory.PluginFactories[entry.Name]
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

	s.mu.Lock()
	s.plugins[entry.Name] = &PluginInstance{
		Name:   entry.Name,
		Client: client,
		Stub:   disp,
	}
	s.mu.Unlock()

	return nil
}

func (s *PluginServer) initCallback() {
	defer s.wg.Done()
	global.LOG.Info("initializing callback server")

	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("panic in callback server: %v, stack: %s", r, debug.Stack())
		}
	}()

	s.callback = &callback.CallbackServer{}
	s.broker = &hplugin.GRPCBroker{}
	s.brokerID = s.broker.NextId()

	s.grpcServer = grpc.NewServer()
	cbpb.RegisterCenterCallbackServer(s.grpcServer, s.callback)

	// 用一个 goroutine 监听 context cancel 来优雅关闭 grpc server
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		<-s.ctx.Done()
		if s.grpcServer != nil {
			s.grpcServer.GracefulStop()
		}
	}()

	s.broker.AcceptAndServe(s.brokerID, func(_ []grpc.ServerOption) *grpc.Server {
		return s.grpcServer
	})
}
