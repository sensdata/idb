package plugin

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/hashicorp/go-hclog"
	hplugin "github.com/hashicorp/go-plugin"
	"github.com/sensdata/idb/center/core/plugin/shared"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
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

	return nil
}

func (s *PluginServer) Stop() error {
	global.LOG.Info("stopping plugin server")

	s.cancel() // 通知所有 goroutine 退出
	s.wg.Wait()

	s.stopPlugins()

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

func (s *PluginServer) Settings() (*model.SettingInfo, error) {
	SettingsRepo := repo.NewSettingsRepo()
	bindIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return nil, err
	}
	bindPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
	if err != nil {
		return nil, err
	}
	bindPortValue, err := strconv.Atoi(bindPort.Value)
	if err != nil {
		return nil, err
	}
	bindDomain, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindDomain"))
	if err != nil {
		return nil, err
	}
	https, err := SettingsRepo.Get(SettingsRepo.WithByKey("Https"))
	if err != nil {
		return nil, err
	}
	httpsCertType, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return nil, err
	}
	httpsCertPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return nil, err
	}
	httpsKeyPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return nil, err
	}

	return &model.SettingInfo{
		BindIP:        bindIP.Value,
		BindPort:      bindPortValue,
		BindDomain:    bindDomain.Value,
		Https:         https.Value,
		HttpsCertType: httpsCertType.Value,
		HttpsCertPath: httpsCertPath.Value,
		HttpsKeyPath:  httpsKeyPath.Value,
	}, nil
}

func (s *PluginServer) loadPlugins() {

	hostRepo := repo.NewHostRepo()
	defaultHost, _ := hostRepo.Get(hostRepo.WithByDefault())

	settingInfo, _ := s.Settings()
	scheme := "http"
	if settingInfo.Https == "yes" {
		scheme = "https"
	}
	host := global.Host
	if settingInfo.BindDomain != "" && settingInfo.BindDomain != host {
		host = settingInfo.BindDomain
	}
	baseUrl := fmt.Sprintf("%s://%s:%d/api/v1", scheme, host, settingInfo.BindPort)

	initConfig := shared.PluginInitConfig{
		API:      baseUrl,
		HTTPS:    scheme == "https",
		Cert:     string(global.CertPem),
		Key:      string(global.KeyPem),
		WorkDir:  constant.CenterBinDir,
		WorkHost: defaultHost.ID,
		AppDir:   constant.AgentDockerDir,
	}

	// 转成 JSON
	jsonBytes, err := json.Marshal(initConfig)
	if err != nil {
		global.LOG.Error("failed to marshal config: %v", err)
		return
	}

	// Base64 编码
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)

	// 写入临时文件
	tmpFile := "/tmp/plugin_boot_config"
	if err := os.WriteFile(tmpFile, []byte(encoded), 0644); err != nil {
		global.LOG.Error("failed to write config file: %v", err)
		return
	}

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
				global.LOG.Error("plugin %s not found at %s, skipping", e.Name, e.Path)
				return
			}

			if err := s.ensurePluginExecutable(e); err != nil {
				global.LOG.Error("ensure plugin %s executable: %v", e.Name, err)
				return
			}

			if err := s.loadPlugin(e); err != nil {
				global.LOG.Error("load plugin %s: %v", e.Name, err)
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
	execPath := filepath.Join(e.Path, e.Name)
	info, err := os.Stat(execPath)
	return err == nil && !info.IsDir()
}

func (s *PluginServer) ensurePluginExecutable(e PluginEntry) error {
	execPath := filepath.Join(e.Path, e.Name)
	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("failed to stat plugin executable %s: %w", execPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("plugin executable path is a directory: %s", execPath)
	}
	if info.Mode()&0111 != 0 {
		return nil
	}

	if err := os.Chmod(execPath, 0755); err != nil {
		return fmt.Errorf("failed to chmod plugin executable %s: %w", execPath, err)
	}
	global.LOG.Warn("plugin %s executable bit was missing, fixed permissions on %s", e.Name, execPath)
	return nil
}

func (s *PluginServer) loadPlugin(entry PluginEntry) error {
	global.LOG.Info("starting to load plugin: %s (path: %s)", entry.Name, entry.Path)

	logger := hclog.New(&hclog.LoggerOptions{
		Output: &PluginLogWriter{},
		Level:  hclog.Info,
	})

	execPath := filepath.Join(entry.Path, entry.Name)
	cmd := exec.Command(execPath, "--config", "/tmp/plugin_boot_config")

	client := hplugin.NewClient(&hplugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              cmd,
		AllowedProtocols: []hplugin.Protocol{hplugin.ProtocolGRPC},
		Logger:           logger,
	})

	rpcClient, err := client.Client()
	if err != nil {
		global.LOG.Error("failed to create rpc client for plugin %s: %v", entry.Name, err)
		client.Kill()
		return fmt.Errorf("failed to connect to plugin %s: %w", entry.Name, err)
	}
	global.LOG.Info("rpc client created for plugin: %s", entry.Name)

	// Request the plugin
	raw, err := rpcClient.Dispense(entry.Name)
	if err != nil {
		global.LOG.Error("failed to dispense grpc interface for plugin %s: %v", entry.Name, err)
		client.Kill()
		return fmt.Errorf("failed to dispense plugin %s: %w", entry.Name, err)
	}
	if raw == nil {
		global.LOG.Error("dispense returned nil for plugin: %s", entry.Name)
		client.Kill()
		return fmt.Errorf("dispense returned nil")
	}
	// 动态检查插件类型
	var ok bool
	switch entry.Name {
	case "scriptmanager":
		_, ok = raw.(shared.ScriptManager)
	case "mysqlmanager":
		_, ok = raw.(shared.MysqlManager)
	case "postgresql":
		_, ok = raw.(shared.PostgreSql)
	case "redis":
		_, ok = raw.(shared.Redis)
	case "idb-rsync":
		_, ok = raw.(shared.Rsync)
	case "pma":
		_, ok = raw.(shared.Pma)
	default:
		// 对于未来的插件，可以在这里添加新的case
		global.LOG.Error("unknown plugin type: %s", entry.Name)
		client.Kill()
		return fmt.Errorf("unknown plugin type: %s", entry.Name)
	}
	if !ok {
		global.LOG.Error("dispensed plugin %s does not implement the required interface", entry.Name)
		client.Kill()
		return fmt.Errorf("invalid plugin type for %s", entry.Name)
	}
	global.LOG.Info("dispense grpc interface successful for plugin: %s", entry.Name)

	s.mu.Lock()
	s.plugins[entry.Name] = &PluginInstance{
		Name:   entry.Name,
		Client: client,
		Stub:   raw,
	}
	s.mu.Unlock()

	global.LOG.Info("plugin %s fully registered and available", entry.Name)

	return nil
}
