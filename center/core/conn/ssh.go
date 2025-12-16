package conn

import (
	"fmt"
	"io"
	"io/fs"
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
	"github.com/sensdata/idb/core/logstream/pkg/types"
	"github.com/sensdata/idb/core/logstream/pkg/writer"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"golang.org/x/crypto/ssh"
)

type SSHService struct {
	done          chan struct{}
	mu            sync.Mutex
	sshClients    map[string]*ssh.Client
	responseChMap map[string]chan string
}

type ISSHService interface {
	Start() error
	Stop() error
	TestConnection(host model.Host) error
	InstallAgent(host model.Host, taskId string, upgrade bool) error
	UninstallAgent(host model.Host, taskId string) error
	RestartAgent(host model.Host) error
	TransferDir(host model.Host, taskId string, localDir string, remoteDir string, wp *writer.Writer) error
}

func NewSSHService() ISSHService {
	return &SSHService{
		sshClients:    make(map[string]*ssh.Client),
		responseChMap: make(map[string]chan string),
	}
}

func (s *SSHService) Start() error {

	global.LOG.Info("SSHService Starting")

	// 尝试连接所有的host
	go s.ensureConnections()

	return nil
}

func (s *SSHService) Stop() error {
	close(s.done)

	// 关闭所有的ssh连接
	s.mu.Lock()
	for _, client := range s.sshClients {
		client.Close()
	}
	s.mu.Unlock()

	return nil
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
	global.LOG.Info("authmode: %s, %s, %s", host.AuthMode, host.Password, host.PrivateKey)
	if host.AuthMode == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(host.Password)}
	} else {
		// 读取宿主机文件, 需要利用agent连接来读取文件内容
		privateKey, err := getPrivateKey(host.PrivateKey)
		if err != nil {
			global.LOG.Error("failed to read private key file: %v", err)
			return errors.New(constant.ErrFileRead)
		}
		passPhrase := []byte(host.PassPhrase)

		signer, err := makePrivateKeySigner([]byte(privateKey.Content), passPhrase)
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

func (s *SSHService) getClient(host model.Host) (*ssh.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	client, exists := s.sshClients[host.Addr]
	if !exists || client == nil || !isValidSSHClient(client) {
		global.LOG.Error("ssh connection to host %s not exists", host.Addr)
		return nil, fmt.Errorf("ssh connection to host %s not exists", host.Addr)
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

func taskLog(wp *writer.Writer, level types.LogLevel, message string) {
	if wp != nil {
		(*wp).Write(level, message, map[string]string{})
	}
}

func taskStatus(taskId string, status types.TaskStatus) {
	if taskId == "" {
		return
	}

	// 延迟1秒更新状态
	time.Sleep(time.Second)
	if err := global.LogStream.UpdateTaskStatus(taskId, status); err != nil {
		global.LOG.Error("Failed to update task status to %s : %v", status, err)
	}
}

func (s *SSHService) InstallAgent(host model.Host, taskId string, upgrade bool) error {

	taskStatus(taskId, types.TaskStatusRunning)

	var writer *writer.Writer
	if taskId != "" {
		w, err := global.LogStream.GetWriter(taskId)
		if err != nil {
			taskStatus(taskId, types.TaskStatusCanceled)
			global.LOG.Error("Failed to get log writer for task %s: %v", taskId, err)
			return fmt.Errorf("failed to get log writer for task %s: %v", taskId, err)
		}
		writer = &w
	}

	global.LOG.Info("Install agent to host %s begin", host.Addr)
	taskLog(writer, types.LogLevelInfo, fmt.Sprintf("Install agent to host %s begin", host.Addr))

	// 1. 检查 agent 安装包路径
	taskLog(writer, types.LogLevelInfo, "Checking agent package path")
	agentPackagePath := filepath.Join(constant.CenterAgentDir, constant.CenterAgentPkg)
	if _, err := os.Stat(agentPackagePath); os.IsNotExist(err) {
		global.LOG.Error("Agent package not found at %s", agentPackagePath)
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Agent package not found at %s", agentPackagePath))
		taskStatus(taskId, types.TaskStatusCanceled)
		return fmt.Errorf("agent package not found at %s", agentPackagePath)
	}

	// 2. 拿 SSH 连接
	taskLog(writer, types.LogLevelInfo, "Checking SSH connection")
	client, err := s.getClient(host)
	if err != nil {
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to connect to host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusCanceled)
		return err
	}

	// 3. 检查目标机器是否已安装 agent
	taskLog(writer, types.LogLevelInfo, "Checking agent installation status")
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
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to check agent installation status on host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusCanceled)
		return fmt.Errorf("failed to check agent installation status: %v", err)
	}

	if !upgrade && strings.TrimSpace(output) == "installed" {
		global.LOG.Info("Agent is already installed on host %s", host.Addr)
		taskLog(writer, types.LogLevelWarn, fmt.Sprintf("Agent is already installed on host %s", host.Addr))
		taskStatus(taskId, types.TaskStatusCanceled)
		return nil
	}

	// 4. 传输 agent 包文件
	taskLog(writer, types.LogLevelInfo, "Start transferring agent package")
	err = s.transferFile(client, agentPackagePath, "/tmp/idb-agent.tar.gz", writer)
	if err != nil {
		global.LOG.Error("Failed to transfer agent package to host %s: %v", host.Addr, err)
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to transfer agent package to host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusFailed)
		return fmt.Errorf("failed to transfer agent package: %v", err)
	}

	// 5. 执行解压和安装命令
	taskLog(writer, types.LogLevelInfo, "Unpacking and installing agent")
	tmpCmd := `
        sudo mkdir -p /tmp/idb-agent && 
        sudo tar -xzvf /tmp/idb-agent.tar.gz -C /tmp/idb-agent && 
        cd /tmp/idb-agent && 
        sudo sed -i "s/secret_key=.*/secret_key=${AGENT_KEY}/" idb-agent.conf &&
        sudo sh install-agent.sh && 
        sudo rm -rf /tmp/idb-agent /tmp/idb-agent.tar.gz
    `
	installCmd := fmt.Sprintf(`AGENT_KEY="%s" && %s`, host.AgentKey, tmpCmd)
	global.LOG.Info("installCmd: %s", installCmd)
	output, err = executeCommand(client, installCmd)
	if err != nil {
		global.LOG.Error("Failed to install agent to host %s: %v", host.Addr, err)
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to install agent to host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusFailed)
		return fmt.Errorf("failed to install agent: %v", err)
	}

	global.LOG.Info("Agent installation output: %s", output)
	global.LOG.Info("Install agent to host %s completed", host.Addr)
	taskLog(writer, types.LogLevelInfo, fmt.Sprintf("Install agent to host %s completed", host.Addr))
	taskStatus(taskId, types.TaskStatusSuccess)

	// 更新host和安装状态
	installed := "installed"
	hostStatus := core.NewHostStatusInfo()
	hostStatus.Installed = installed
	global.SetHostStatus(host.ID, hostStatus)
	global.SetInstalledStatus(host.ID, &installed)

	return nil
}

func (s *SSHService) UninstallAgent(host model.Host, taskId string) error {

	taskStatus(taskId, types.TaskStatusRunning)

	var writer *writer.Writer
	if taskId != "" {
		w, err := global.LogStream.GetWriter(taskId)
		if err != nil {
			global.LOG.Error("Failed to get log writer for task %s: %v", taskId, err)
			taskStatus(taskId, types.TaskStatusCanceled)
			return fmt.Errorf("failed to get log writer for task %s: %v", taskId, err)
		}
		writer = &w
	}

	global.LOG.Info("Uninstall agent in host %s begin", host.Addr)
	taskLog(writer, types.LogLevelInfo, fmt.Sprintf("Uninstall agent in host %s begin", host.Addr))

	// 1. 检查并确保 SSH 连接存在
	taskLog(writer, types.LogLevelInfo, "Checking SSH connection")
	client, err := s.getClient(host)
	if err != nil {
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to connect to host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusCanceled)
		return err
	}

	// 2. 检查 agent 是否已安装
	taskLog(writer, types.LogLevelInfo, "Checking agent installation status")
	checkCmd := `
		if systemctl is-active --quiet idb-agent.service || [ -f /var/lib/idb-agent/idb-agent ]; then
			echo "installed"
		else
			echo "not installed"
		fi
	`
	output, err := executeCommand(client, checkCmd)
	if err != nil {
		global.LOG.Error("Failed to check agent status on host %s: %v", host.Addr, err)
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to check agent status on host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusCanceled)
		return fmt.Errorf("failed to check agent status: %v", err)
	}

	if strings.TrimSpace(output) != "installed" {
		global.LOG.Info("Agent is not installed on host %s", host.Addr)
		taskLog(writer, types.LogLevelWarn, fmt.Sprintf("Agent is not installed on host %s", host.Addr))
		taskStatus(taskId, types.TaskStatusCanceled)
		return nil
	}

	// 3. 执行卸载的一系列动作命令
	taskLog(writer, types.LogLevelInfo, "Uninstalling agent service")
	uninstallCmd := `
		if ! sudo -n true 2>/dev/null; then
			echo "no_sudo_access"
			exit 1
		fi &&
		(sudo systemctl stop idb-agent.service || true) &&
		(sudo systemctl disable idb-agent.service || true) &&
		sudo rm -f /etc/systemd/system/idb-agent.service &&
		sudo rm -rf /var/lib/idb-agent &&
		sudo rm -rf /etc/idb-agent &&
		sudo rm -rf /var/log/idb-agent &&
		sudo rm -rf /run/idb-agent &&
		sudo systemctl daemon-reload
	`
	output, err = executeCommand(client, uninstallCmd)
	if err != nil {
		if strings.Contains(output, "no_sudo_access") {
			global.LOG.Error("No sudo access on host %s", host.Addr)
			taskLog(writer, types.LogLevelError, fmt.Sprintf("No sudo access on host %s", host.Addr))
			taskStatus(taskId, types.TaskStatusFailed)
			return fmt.Errorf("no sudo access on host %s", host.Addr)
		}
		global.LOG.Error("Failed to uninstall agent on host %s: %v", host.Addr, err)
		taskLog(writer, types.LogLevelError, fmt.Sprintf("Failed to uninstall agent on host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusFailed)
		return fmt.Errorf("failed to uninstall agent: %v", err)
	}

	global.LOG.Info("Agent uninstall output: %s", output)
	global.LOG.Info("Uninstall agent in host %s completed", host.Addr)
	taskLog(writer, types.LogLevelInfo, fmt.Sprintf("Uninstall agent in host %s completed", host.Addr))
	taskStatus(taskId, types.TaskStatusSuccess)

	// 更新host和安装状态
	notInstalled := "not installed"
	hostStatus := core.NewHostStatusInfo()
	hostStatus.Installed = notInstalled
	global.SetHostStatus(host.ID, hostStatus)
	global.SetInstalledStatus(host.ID, &notInstalled)

	return nil
}

func (s *SSHService) agentInstalled(client *ssh.Client, host model.Host) (string, error) {
	installed := "unknown"

	// 检查目标机器是否已安装 agent
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
		return installed, fmt.Errorf("failed to check agent installation status: %v", err)
	}

	installed = strings.TrimSpace(output)
	global.LOG.Info("Agent %s in host %s", installed, host.Addr)

	return installed, nil
}

func (s *SSHService) RestartAgent(host model.Host) error {
	// 检查并确保 SSH 连接存在
	client, err := s.getClient(host)
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

func (s *SSHService) transferFile(client *ssh.Client, localPath, remotePath string, wp *writer.Writer) error {
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

	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	totalSize := fileInfo.Size()

	// 创建远程文件
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		global.LOG.Error("failed to create remote file: %v", err)
		return err
	}
	defer remoteFile.Close()

	// 复制文件内容
	buf := make([]byte, 32*1024) // 32KB 缓冲区
	var copied int64
	lastProgress := 0

	for {
		n, err := localFile.Read(buf)
		if n > 0 {
			_, writeErr := remoteFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write to remote file: %v", writeErr)
			}
			copied += int64(n)

			// 计算并报告进度
			progress := int(float64(copied) / float64(totalSize) * 100)
			if progress > lastProgress {
				taskLog(wp, types.LogLevelInfo, fmt.Sprintf("Transferring agent package: %d%%", progress))
				lastProgress = progress
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read local file: %v", err)
		}
	}

	taskLog(wp, types.LogLevelInfo, "Agent package transfer completed")

	global.LOG.Info("File transferred successfully to %s", remotePath)
	return nil
}

func (s *SSHService) TransferDir(host model.Host, taskId string, localDir string, remoteDir string, wp *writer.Writer) error {
	// 检查并确保 SSH 连接存在
	taskLog(wp, types.LogLevelInfo, "Checking SSH connection")
	client, err := s.getClient(host)
	if err != nil {
		taskLog(wp, types.LogLevelError, fmt.Sprintf("Failed to connect to host %s: %v", host.Addr, err))
		taskStatus(taskId, types.TaskStatusCanceled)
		return err
	}
	return s.transferDir(client, localDir, remoteDir, wp)
}

func (s *SSHService) transferDir(client *ssh.Client, localDir string, remoteDir string, wp *writer.Writer) error {
	// 打开 SFTP 会话
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		global.LOG.Error("failed to start SFTP session: %v", err)
		return fmt.Errorf("failed to start SFTP session: %w", err)
	}
	defer sftpClient.Close()

	// 确保远端目标目录存在
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		global.LOG.Error("failed to create remote dir %s: %v", remoteDir, err)
		return fmt.Errorf("failed to create remote dir %s: %w", remoteDir, err)
	}

	taskLog(wp, types.LogLevelInfo,
		fmt.Sprintf("Start transferring directory: %s -> %s", localDir, remoteDir))

	// 遍历本地目录
	err = filepath.WalkDir(localDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		remotePath := filepath.ToSlash(filepath.Join(remoteDir, relPath))

		if d.IsDir() {
			// 创建远端目录
			if err := sftpClient.MkdirAll(remotePath); err != nil {
				return fmt.Errorf("failed to create remote directory %s: %w", remotePath, err)
			}
			return nil
		}

		// 处理普通文件
		taskLog(wp, types.LogLevelInfo,
			fmt.Sprintf("Transferring file: %s", relPath))

		return transferSingleFile(sftpClient, path, remotePath)
	})

	if err != nil {
		taskLog(wp, types.LogLevelError,
			fmt.Sprintf("Directory transfer failed: %v", err))
		return err
	}

	taskLog(wp, types.LogLevelInfo, "Directory transfer completed")
	return nil
}

func transferSingleFile(
	sftpClient *sftp.Client,
	localPath string,
	remotePath string,
) error {

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.OpenFile(
		remotePath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
	)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("failed to copy file %s: %w", localPath, err)
	}

	return nil
}

func (s *SSHService) ensureConnections() {
	global.LOG.Info("Ensure ssh connections started")

	interval := 10 * time.Second
	maxConcurrency := 5
	sem := make(chan struct{}, maxConcurrency)

	for {
		start := time.Now()

		select {
		case <-s.done:
			global.LOG.Info("Ensure ssh connections stopped")
			return
		default:
		}

		hosts, err := HostRepo.GetList()
		if err != nil {
			global.LOG.Error("Failed to get host list: %v", err)
			time.Sleep(interval)
			continue
		}

		var wg sync.WaitGroup
		for _, host := range hosts {
			wg.Add(1)
			sem <- struct{}{} // 占用一个并发槽位

			go func(h model.Host) {
				defer func() {
					<-sem
					wg.Done()
				}()

				s.handleHost(&h)
			}(host)
		}

		wg.Wait()

		elapsed := time.Since(start)
		if elapsed < interval {
			wait := interval - elapsed
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-s.done:
				global.LOG.Info("Ensure ssh connections done")
				timer.Stop()
				return
			}
			timer.Stop()
		}
	}
}

// handleHost 负责处理单个 host 的 SSH 检查与连接逻辑
func (s *SSHService) handleHost(host *model.Host) {
	global.LOG.Info("Ensure ssh connection for host %s", host.Addr)

	client, err := s.getClient(*host)
	if err != nil || client == nil {
		if err := s.connectToHost(host); err != nil {
			global.LOG.Warn("SSH reconnect failed for host %s: %v", host.Addr, err)
			return
		}
		global.LOG.Info("SSH connection re-established for host %s", host.Addr)
		return
	}

	installed, err := s.agentInstalled(client, *host)
	if err != nil {
		global.LOG.Warn("Check agent install failed for host %s: %v", host.Addr, err)
		return
	}
	global.SetInstalledStatus(host.ID, &installed)
	global.LOG.Info("Host agent status: %s", installed)
}

func getPrivateKey(path string) (*core.FileInfo, error) {
	var fileInfo core.FileInfo

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return &fileInfo, err
	}

	req := core.FileContentReq{
		Path:   path,
		Expand: true,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &fileInfo, err
	}

	actionRequest := core.HostAction{
		HostID: host.ID,
		Action: core.Action{
			Action: core.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return &fileInfo, err
	}

	if !actionResponse.Result {
		global.LOG.Error("action failed")
		if strings.Contains(actionResponse.Data, "no such file or directory") {
			return &fileInfo, constant.ErrFileNotExist
		}
		return &fileInfo, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &fileInfo, fmt.Errorf("json err: %v", err)
	}
	return &fileInfo, nil
}

func (s *SSHService) connectToHost(host *model.Host) error {
	// 先关闭已有的client
	s.mu.Lock()
	if client, exists := s.sshClients[host.Addr]; exists && client != nil {
		client.Close()
		delete(s.sshClients, host.Addr)
	}
	s.mu.Unlock()

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
		// 读取宿主机文件, 需要利用agent连接来读取文件内容
		privateKey, err := getPrivateKey(host.PrivateKey)
		if err != nil {
			global.LOG.Error("failed to read private key file: %v", err)
			return fmt.Errorf("failed to read private key file: %v", err)
		}

		passPhrase := []byte(host.PassPhrase)

		signer, err := makePrivateKeySigner([]byte(privateKey.Content), passPhrase)
		if err != nil {
			global.LOG.Error("Failed to config private key to host %s, %v", host.Addr, err)
			return fmt.Errorf("failed to config private key to host %s, %v", host.Addr, err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	config.Timeout = 3 * time.Second
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial(proto, dialAddr, config)
	if err != nil {
		global.LOG.Error("Failed to create ssh connection to host %s, %v", host.Addr, err)
		return fmt.Errorf("failed to create ssh connection to host %s, %v", host.Addr, err)
	}
	s.mu.Lock()
	s.sshClients[host.Addr] = client
	s.mu.Unlock()

	global.LOG.Info("SSH connection to %s created", host.Addr)
	return nil
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
