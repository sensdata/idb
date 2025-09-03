package plugin

import (
	"archive/tar"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	execPath := filepath.Join(e.Path, e.Name)
	info, err := os.Stat(execPath)
	return err == nil && !info.IsDir()
}

func (s *PluginServer) installPlugin(e PluginEntry) error {
	global.LOG.Info("fetching latest version info for plugin %s from %s", e.Name, e.Url)

	// 先获取 latest 文件内容（版本号）
	resp, err := http.Get(e.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch latest version info for plugin %s: %w", e.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d when fetching latest version info for plugin %s", resp.StatusCode, e.Name)
	}

	latestBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read latest version info for plugin %s: %w", e.Name, err)
	}
	latest := strings.TrimSpace(string(latestBytes))
	if latest == "" {
		return fmt.Errorf("latest version info for plugin %s is empty", e.Name)
	}

	// 拼接真实的 tar.gz 下载地址
	tarballURL := fmt.Sprintf("https://static.sensdata.com/idb/plugins/%s/%s/%s.tar.gz",
		e.Name, latest, e.Name)

	global.LOG.Info("downloading plugin %s version %s from %s", e.Name, latest, tarballURL)

	// 下载 tar.gz 包
	resp, err = http.Get(tarballURL)
	if err != nil {
		return fmt.Errorf("failed to download plugin %s: %w", e.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		global.LOG.Error("unexpected HTTP status %d for plugin %s", resp.StatusCode, e.Name)
		return fmt.Errorf("unexpected HTTP status %d for plugin %s", resp.StatusCode, e.Name)
	}

	// 创建插件目录
	if err := os.MkdirAll(e.Path, 0755); err != nil {
		global.LOG.Error("failed to create plugin path %s: %v", e.Path, err)
		return fmt.Errorf("failed to create plugin path %s: %w", e.Path, err)
	}

	// 解压 tar.gz 到 e.Path
	if err := extractTarGz(resp.Body, e.Path); err != nil {
		global.LOG.Error("failed to extract plugin tar.gz: %v", err)
		return fmt.Errorf("failed to extract plugin tar.gz: %w", err)
	}

	// 设置主执行文件为可执行
	execPath := filepath.Join(e.Path, e.Name)
	if err := os.Chmod(execPath, 0755); err != nil {
		global.LOG.Error("failed to chmod plugin exec file: %v", err)
		return fmt.Errorf("failed to chmod plugin exec file: %w", err)
	}

	global.LOG.Info("plugin %s installed at %s", e.Name, e.Path)
	return nil
}

// 解压 tar.gz 到指定目录
func extractTarGz(gzipStream io.Reader, dest string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("gzip reader error: %w", err)
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 解压完毕
		}
		if err != nil {
			return fmt.Errorf("tar read error: %w", err)
		}

		name := header.Name
		relPath := filepath.Clean(name)
		targetPath := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("mkdir error: %w", err)
			}
		case tar.TypeReg:
			// 保证父目录存在
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("mkdir for file error: %w", err)
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("create file error: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("copy file error: %w", err)
			}
			outFile.Close()
		default:
			global.LOG.Warn("unsupported tar entry: %s", name)
		}
	}

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
	case "auth":
		_, ok = raw.(shared.Auth)
	case "scriptmanager":
		_, ok = raw.(shared.ScriptManager)
	case "mysqlmanager":
		_, ok = raw.(shared.MysqlManager)
	case "postgresql":
		_, ok = raw.(shared.PostgreSql)
	case "redis":
		_, ok = raw.(shared.Redis)
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
