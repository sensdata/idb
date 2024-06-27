package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
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
	ConnectToHost(addr string) error
}

func NewISSHService() ISSHService {
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

func (s *SSHService) ensureConnections() {
	//获取所有的host
	hosts, err := HostRepo.GetList()
	if err != nil {
		global.LOG.Error("Failed to get host list: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 挨个确认是否已经建立连接
	for _, host := range hosts {
		addr := host.Addr
		// 判断sshClients中是否包含addr的数据
		_, exists := s.sshClients[addr]
		if exists {
			continue
		} else {
			go s.connectToHost(&host)
		}
	}
}

func (s *SSHService) connectToHost(host *model.Host) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//connection config
	config := &ssh.ClientConfig{}

	proto := "tcp"
	addr := host.Addr
	if strings.Contains(host.Addr, ":") {
		addr = fmt.Sprintf("[%s]", host.Addr)
		proto = "tcp6"
	}
	config.SetDefaults()
	addr = fmt.Sprintf("%s:%d", addr, host.Port)
	config.User = host.User

	if host.AuthMode == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(host.Password)}
	} else {
		privateKey := []byte(host.PrivateKey)
		passPhrase := []byte(host.PrivateKey)

		signer, err := makePrivateKeySigner(privateKey, passPhrase)
		if err != nil {
			return
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	config.Timeout = 5 * time.Second

	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	// get client
	client, err := ssh.Dial(
		proto,
		fmt.Sprintf("%s:%d", host.Addr, host.Port),
		config,
	)
	if err != nil {
		global.LOG.Error("Failed to create ssh connection to host %s, %v", host.Addr, err)
		return
	}
	s.sshClients[host.Addr] = client

	global.LOG.Info("SSH connection to %s created", host.Addr)
}

func makePrivateKeySigner(privateKey []byte, passPhrase []byte) (ssh.Signer, error) {
	if len(passPhrase) != 0 {
		return ssh.ParsePrivateKeyWithPassphrase(privateKey, passPhrase)
	}
	return ssh.ParsePrivateKey(privateKey)
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

func (s *SSHService) ConnectToHost(addr string) error {
	//获取host
	host, err := HostRepo.Get(HostRepo.WithByAddr(addr))
	if err != nil {
		global.LOG.Error("Failed to get host: %v", err)
		return errors.New("host not exist")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 已存在
	_, exists := s.sshClients[host.Addr]
	if exists {
		return nil
	}

	go s.connectToHost(&host)

	return nil
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
