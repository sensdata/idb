package conn

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	core "github.com/sensdata/idb/core/model"
	"golang.org/x/crypto/ssh"
)

type SSHService struct {
	mu            sync.Mutex
	sshClients    map[string]*ssh.Client
	responseChMap map[string]chan string
}

type ISSHService interface {
	Start() error
	Stop()
	TestConnection(host model.Host) error
	InstallAgent(host model.Host) error
	AgentStatus(host model.Host) (*core.AgentStatus, error)
	RestartAgent(host model.Host) error
}

func NewSSHService() ISSHService {
	return &SSHService{
		sshClients:    make(map[string]*ssh.Client),
		responseChMap: make(map[string]chan string),
	}
}

func (s *SSHService) Start() error {

	global.LOG.Info("SSHService started")

	// 尝试连接所有的host
	go s.ensureConnections()

	return nil
}

func (s *SSHService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.sshClients {
		client.Close()
	}
}

func (s *SSHService) TestConnection(host model.Host) error {
	proto := "tcp"
	addr := host.Addr
	if strings.Contains(host.Addr, ":") {
		addr = fmt.Sprintf("[%s]", host.Addr)
		proto = "tcp6"
	}
	dialAddr := fmt.Sprintf("%s:%d", addr, host.Port)

	global.LOG.Info("try connect to host ssh: %s", dialAddr)

	//connection config
	config := &ssh.ClientConfig{}
	config.SetDefaults()
	config.User = host.User
	global.LOG.Info("authmode: %s, %s", host.AuthMode, host.Password)
	if host.AuthMode == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(host.Password)}
	} else {
		// 读取文件
		privateKey, err := os.ReadFile(host.PrivateKey)
		if err != nil {
			global.LOG.Error("failed to read private key file: %v", err)
			return errors.New(constant.ErrFileRead)
		}
		passPhrase := []byte(host.PassPhrase)

		signer, err := makePrivateKeySigner(privateKey, passPhrase)
		if err != nil {
			global.LOG.Error("Failed to config private key to host %s, %v", host.Addr, err)
			return fmt.Errorf("failed to config private key to host %s, %v", host.Addr, err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	config.Timeout = 5 * time.Second
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial(proto, dialAddr, config)
	if err != nil {
		global.LOG.Error("Failed to create ssh connection to host %s, %v", host.Addr, err)
		return fmt.Errorf("failed to create ssh connection to host %s, %v", host.Addr, err)
	}
	client.Close()

	global.LOG.Info("SSH test to %s passed", host.Addr)

	return nil
}

func (s *SSHService) checkClient(host model.Host) (*ssh.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.sshClients[host.Addr]
	if !exists || client == nil || !isValidSSHClient(client) {
		delete(s.sshClients, host.Addr)

		// 如果连接不存在，或者连接无效，尝试建立连接
		resultCh := make(chan error, 1)
		go s.connectToHost(&host, resultCh)
		err := <-resultCh
		if err != nil {
			global.LOG.Error("Failed to connect to host %s: %v", host.Addr, err)
			return nil, fmt.Errorf("failed to connect to host %s: %v", host.Addr, err)
		}
		client = s.sshClients[host.Addr]
	}

	return client, nil
}

func isValidSSHClient(client *ssh.Client) bool {
	// 尝试发送一个简单的命令来验证连接
	session, err := client.NewSession()
	if err != nil {
		return false
	}
	defer session.Close()

	// 使用轻量级的命令，如 "echo"
	err = session.Run("echo") // 只执行一个无害的命令
	return err == nil
}

func (s *SSHService) InstallAgent(host model.Host) error {
	global.LOG.Info("Install agent to host %s begin", host.Addr)

	// 1. 检查 agent 安装包路径
	agentPackagePath := filepath.Join(constant.CenterAgentDir, fmt.Sprintf(constant.CenterAgentPkg, global.Version))
	if _, err := os.Stat(agentPackagePath); os.IsNotExist(err) {
		global.LOG.Error("Agent package not found at %s", agentPackagePath)
		return fmt.Errorf("agent package not found at %s", agentPackagePath)
	}

	// 2. 检查并确保 SSH 连接存在
	client, err := s.checkClient(host)
	if err != nil {
		return err
	}

	// 3. 检查目标机器是否已安装 agent
	checkCmd := `
        if systemctl is-active --quiet idb-agent.service && [ -f /var/lib/idb-agent/idb-agent ]; then
            echo "installed"
        else
            echo "not installed"
        fi
    `
	output, err := executeCommand(client, checkCmd)
	if err != nil {
		global.LOG.Error("Failed to check agent installation status on host %s: %v", host.Addr, err)
		return fmt.Errorf("failed to check agent installation status: %v", err)
	}

	if strings.TrimSpace(output) == "installed" {
		global.LOG.Info("Agent is already installed on host %s", host.Addr)
		return nil
	}

	// 4. 传输 agent 包文件
	err = s.transferFile(client, agentPackagePath, "/tmp/idb-agent.tar.gz")
	if err != nil {
		global.LOG.Error("Failed to transfer agent package to host %s: %v", host.Addr, err)
		return fmt.Errorf("failed to transfer agent package: %v", err)
	}

	// 5. 执行解压和安装命令
	installCmd := `
        mkdir -p /tmp/idb-agent && 
        tar -xzvf /tmp/idb-agent.tar.gz -C /tmp/idb-agent && 
        cd /tmp/idb-agent && 
        sh install-agent.sh && 
        rm -rf /tmp/idb-agent /tmp/idb-agent.tar.gz
    `
	output, err = executeCommand(client, installCmd)
	if err != nil {
		global.LOG.Error("Failed to install agent to host %s: %v", host.Addr, err)
		return fmt.Errorf("failed to install agent: %v", err)
	}

	global.LOG.Info("Agent installation output: %s", output)
	global.LOG.Info("Install agent to host %s completed", host.Addr)

	return nil
}

func (s *SSHService) AgentStatus(host model.Host) (*core.AgentStatus, error) {
	// 1. 检查并确保 SSH 连接存在
	client, err := s.checkClient(host)
	if err != nil {
		return nil, err
	}

	// 2. 检查目标机器是否已安装 agent
	checkCmd := `
        if systemctl is-active --quiet idb-agent.service && [ -f /var/lib/idb-agent/idb-agent ]; then
            echo "installed"
        else
            echo "not installed"
        fi
    `
	output, err := executeCommand(client, checkCmd)
	if err != nil {
		global.LOG.Error("Failed to check agent installation status on host %s: %v", host.Addr, err)
		return nil, fmt.Errorf("failed to check agent installation status: %v", err)
	}

	status := strings.TrimSpace(output)
	global.LOG.Info("Agent %s in host %s", status, host.Addr)

	return &core.AgentStatus{Status: status}, nil
}

func (s *SSHService) RestartAgent(host model.Host) error {
	// 检查并确保 SSH 连接存在
	client, err := s.checkClient(host)
	if err != nil {
		return err
	}

	restartCmd := `systemctl restart idb-agent.service`
	output, err := executeCommand(client, restartCmd)
	if err != nil {
		global.LOG.Error("Failed to restart agent in host %s: %v", host.Addr, err)
		return fmt.Errorf("failed to restart agent in host %s: %v", host.Addr, err)
	}

	global.LOG.Info("Agent restart output: %s", output)

	return nil
}

func (s *SSHService) transferFile(client *ssh.Client, localPath, remotePath string) error {
	// 打开 SFTP 会话
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		global.LOG.Error("failed to start SFTP session: %v", err)
		return err
	}
	defer sftpClient.Close()

	// 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		global.LOG.Error("failed to open local file: %v", err)
		return err
	}
	defer localFile.Close()

	// 创建远程文件
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		global.LOG.Error("failed to create remote file: %v", err)
		return err
	}
	defer remoteFile.Close()

	// 复制文件内容
	_, err = localFile.WriteTo(remoteFile)
	if err != nil {
		global.LOG.Error("failed to write file to remote server: %v", err)
		return err
	}

	global.LOG.Info("File transferred successfully to %s", remotePath)
	return nil
}

func (s *SSHService) ensureConnections() {
	//获取所有的host
	hosts, err := HostRepo.GetList()
	if err != nil {
		global.LOG.Error("Failed to get host list: %v", err)
		return
	}
	global.LOG.Info("%d hosts to connect", len(hosts))

	// 挨个确认是否已经建立连接
	for _, host := range hosts {
		_, err := s.checkClient(host)
		if err != nil {
			continue
		} else {
			go s.InstallAgent(host)
		}
	}
}

func (s *SSHService) connectToHost(host *model.Host, resultCh chan<- error) {
	proto := "tcp"
	addr := host.Addr
	if strings.Contains(host.Addr, ":") {
		addr = fmt.Sprintf("[%s]", host.Addr)
		proto = "tcp6"
	}
	dialAddr := fmt.Sprintf("%s:%d", addr, host.Port)

	global.LOG.Info("try connect to host ssh: %s", dialAddr)

	//connection config
	config := &ssh.ClientConfig{}
	config.SetDefaults()
	config.User = host.User
	global.LOG.Info("authmode: %s, %s", host.AuthMode, host.Password)
	if host.AuthMode == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(host.Password)}
	} else {
		// 读取文件
		privateKey, err := os.ReadFile(host.PrivateKey)
		if err != nil {
			global.LOG.Error("failed to read private key file: %v", err)
			resultCh <- err
			return
		}

		passPhrase := []byte(host.PassPhrase)

		signer, err := makePrivateKeySigner(privateKey, passPhrase)
		if err != nil {
			global.LOG.Error("Failed to config private key to host %s, %v", host.Addr, err)
			resultCh <- err
			return
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	config.Timeout = 5 * time.Second
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial(proto, dialAddr, config)
	if err != nil {
		global.LOG.Error("Failed to create ssh connection to host %s, %v", host.Addr, err)
		resultCh <- err
		return
	}
	s.sshClients[host.Addr] = client

	global.LOG.Info("SSH connection to %s created", host.Addr)
	resultCh <- nil
}

func makePrivateKeySigner(privateKey []byte, passPhrase []byte) (ssh.Signer, error) {
	if len(passPhrase) != 0 {
		return ssh.ParsePrivateKeyWithPassphrase(privateKey, passPhrase)
	}
	return ssh.ParsePrivateKey(privateKey)
}

func executeCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		global.LOG.Error("failed to new sessioin: %v", err)
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		global.LOG.Error("failed to send command: %v", err)
		return "", err
	}
	return string(output), nil
}
