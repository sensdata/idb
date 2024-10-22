package conn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
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
	ExecuteCommand(addr string, command string) (string, error)
	TestConnection(host model.Host) error
	InstallAgent(host model.Host) error
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

func (s *SSHService) ExecuteCommand(addr string, command string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.sshClients[addr]
	if !exists {
		return "", errors.New("host disconnected")
	}

	return executeCommand(s.sshClients[addr], command)
}

func (s *SSHService) TestConnection(host model.Host) error {
	// 已存在
	_, exists := s.sshClients[host.Addr]
	if exists {
		return nil
	}

	resultCh := make(chan error, 1)
	go s.connectToHost(&host, resultCh)

	err := <-resultCh
	if err != nil {
		global.LOG.Error("Failed to connect to host %s: %v", host.Addr, err)
		return err
	}

	return nil
}

func (s *SSHService) InstallAgent(host model.Host) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	global.LOG.Info("Install agent to host %s begin", host.Addr)

	// 1. 检查 agent 安装包路径
	agentPackagePath := filepath.Join(constant.CenterAgentDir, constant.CenterAgentDir)
	if _, err := os.Stat(agentPackagePath); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("Agent package not found at %s", agentPackagePath)
		global.LOG.Error(errMsg)
		return errors.New(errMsg)
	}

	// 2. 检查并确保 SSH 连接存在
	client, exists := s.sshClients[host.Addr]
	if !exists {
		// 如果连接不存在，尝试建立连接
		resultCh := make(chan error, 1)
		go s.connectToHost(&host, resultCh)
		err := <-resultCh
		if err != nil {
			global.LOG.Error("Failed to connect to host %s: %v", host.Addr, err)
			return fmt.Errorf("failed to connect to host %s: %v", host.Addr, err)
		}
		client = s.sshClients[host.Addr]
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
        sudo ./install_agent.sh && 
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

func (s *SSHService) transferFile(client *ssh.Client, localPath, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer w.Close()
		fmt.Fprintf(w, "C0644 %d %s\n", stat.Size(), remotePath)

		buf := make([]byte, 32*1024) // 32KB buffer
		for {
			n, err := file.Read(buf)
			if err != nil {
				if err != io.EOF {
					global.LOG.Error("Error reading file: %v", err)
				}
				break
			}
			if _, err := w.Write(buf[:n]); err != nil {
				global.LOG.Error("Error writing to SSH session: %v", err)
				break
			}
		}
		fmt.Fprint(w, "\x00") // Transfer end with null byte
	}()

	if err := session.Run("/usr/bin/scp -t " + remotePath); err != nil {
		return fmt.Errorf("failed to run scp: %v", err)
	}

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
		addr := host.Addr
		// 判断sshClients中是否包含addr的数据
		_, exists := s.sshClients[addr]
		if exists {
			continue
		} else {
			resultCh := make(chan error, 1)
			go s.connectToHost(&host, resultCh)
			// handle the result if needed
			err := <-resultCh
			if err != nil {
				global.LOG.Error("Failed to connect to host %s: %v", host.Addr, err)
			} else {
				go s.InstallAgent(host)
			}
		}
	}
}

func (s *SSHService) connectToHost(host *model.Host, resultCh chan<- error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
		// Decode private key after retrieving
		decodedPrivateKey, err := base64.StdEncoding.DecodeString(host.PrivateKey)
		if err != nil {
			global.LOG.Error("Failed to config private key to host %s, %v", host.Addr, err)
			resultCh <- err
			return
		}
		passPhrase := []byte(host.PassPhrase)

		signer, err := makePrivateKeySigner(decodedPrivateKey, passPhrase)
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
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
