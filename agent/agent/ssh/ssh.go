package ssh

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
	"github.com/sensdata/idb/core/utils/systemctl"
	"golang.org/x/crypto/ssh"
)

const sshPath = "/etc/ssh/sshd_config"

type SSHService struct{}

type ISSHService interface {
	GetConfig() (*model.SSHInfo, error)
	UpdateConfig(req model.SSHUpdate) error

	GetContent() (*model.SSHConfigContent, error)
	UpdateContent(req model.ContentUpdate) error

	OperateSSH(req model.SSHOperate) error

	CreateKey(req model.GenerateKey) error
	ListKeys(req model.ListKey) (*model.PageResult, error)
	EnableKey(req model.EnableKey) error
	RemoveKey(req model.RemoveKey) error
	SetKeyPassword(req model.SetKeyPassword) error
	ChangeKeyPassword(req model.UpdateKeyPassword) error
	ClearKeyPassword(req model.SetKeyPassword) error
	ListAuthKeys() (*model.PageResult, error)
	AddAuthKey(req model.AddAuthKey) error
	RemoveAuthKey(req model.RemoveAuthKey) error

	LoadLog(req model.SearchSSHLog) (*model.SSHLog, error)
}

func NewISSHService() ISSHService {
	return &SSHService{}
}

func (u *SSHService) GetConfig() (*model.SSHInfo, error) {
	data := model.SSHInfo{
		AutoStart:              true,
		Status:                 constant.StatusEnable,
		Message:                "",
		Port:                   "22",
		ListenAddress:          "",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}
	serviceName, err := loadServiceName()
	if err != nil {
		data.Status = constant.StatusDisable
		data.Message = err.Error()
	} else {
		active, err := systemctl.IsActive(serviceName)
		if !active {
			data.Status = constant.StatusDisable
			data.Message = err.Error()
		} else {
			data.Status = constant.StatusEnable
		}
	}

	out, err := systemctl.RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		data.AutoStart = false
	} else {
		if out == "alias\n" {
			data.AutoStart, _ = systemctl.IsEnable("ssh")
		} else {
			data.AutoStart = out == "enabled\n"
		}
	}

	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		data.Message = err.Error()
		data.Status = constant.StatusDisable
	}
	lines := strings.Split(string(sshConf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Port ") {
			data.Port = strings.ReplaceAll(line, "Port ", "")
		}
		if strings.HasPrefix(line, "ListenAddress ") {
			itemAddr := strings.ReplaceAll(line, "ListenAddress ", "")
			if len(data.ListenAddress) != 0 {
				data.ListenAddress += ("," + itemAddr)
			} else {
				data.ListenAddress = itemAddr
			}
		}
		if strings.HasPrefix(line, "PasswordAuthentication ") {
			data.PasswordAuthentication = strings.ReplaceAll(line, "PasswordAuthentication ", "")
		}
		if strings.HasPrefix(line, "PubkeyAuthentication ") {
			data.PubkeyAuthentication = strings.ReplaceAll(line, "PubkeyAuthentication ", "")
		}
		if strings.HasPrefix(line, "PermitRootLogin ") {
			data.PermitRootLogin = strings.ReplaceAll(strings.ReplaceAll(line, "PermitRootLogin ", ""), "prohibit-password", "without-password")
		}
		if strings.HasPrefix(line, "UseDNS ") {
			data.UseDNS = strings.ReplaceAll(line, "UseDNS ", "")
		}
	}
	return &data, nil
}

func (u *SSHService) UpdateConfig(req model.SSHUpdate) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}

	sshConf, err := os.ReadFile(sshPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(sshConf), "\n")
	newFiles := updateSSHConf(lines, req)
	file, err := os.OpenFile(sshPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(strings.Join(newFiles, "\n")); err != nil {
		return err
	}

	// 重启
	sudo := utils.SudoHandleCmd()
	_, _ = utils.Execf("%s systemctl restart %s", sudo, serviceName)
	return nil
}

func (u *SSHService) GetContent() (*model.SSHConfigContent, error) {
	var result model.SSHConfigContent
	if _, err := os.Stat("/etc/ssh/sshd_config"); err != nil {
		return &result, errors.New(constant.ErrFileNotFound)
	}
	content, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return &result, err
	}
	result = model.SSHConfigContent{
		Content: string(content),
	}
	return &result, nil
}

func (u *SSHService) UpdateContent(req model.ContentUpdate) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}

	// 检查文件是否存在
	if _, err := os.Stat("/etc/ssh/sshd_config"); err != nil {
		return errors.New(constant.ErrFileNotFound)
	}

	// 创建临时文件以保证写入的原子性
	tempFile, err := os.CreateTemp("/etc/ssh", "sshd_config_temp")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name()) // 确保在出错时删除临时文件
	defer tempFile.Close()

	// 将内容写入临时文件
	if _, err := tempFile.WriteString(req.Content); err != nil {
		return err
	}

	// 确保写入的内容被完全写入磁盘
	if err := tempFile.Sync(); err != nil {
		return err
	}

	// 获取原文件的权限模式
	fileInfo, err := os.Stat("/etc/ssh/sshd_config")
	if err != nil {
		return err
	}

	// 将临时文件的权限修改为与原文件一致
	if err := os.Chmod(tempFile.Name(), fileInfo.Mode()); err != nil {
		return err
	}

	// 替换原始文件
	if err := os.Rename(tempFile.Name(), "/etc/ssh/sshd_config"); err != nil {
		return err
	}

	// 重启
	sudo := utils.SudoHandleCmd()
	_, _ = utils.Execf("%s systemctl restart %s", sudo, serviceName)
	return nil
}

func (u *SSHService) OperateSSH(req model.SSHOperate) error {
	serviceName, err := loadServiceName()
	if err != nil {
		return err
	}
	sudo := utils.SudoHandleCmd()

	switch req.Operation {
	case "enable", "disable":
		serviceName += ".service"
	case "stop", "reload", "restart":
		// do nothing
	default:
		return fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	stdout, err := utils.Execf("%s systemctl %s %s", sudo, req.Operation, serviceName)
	if err != nil {
		if strings.Contains(stdout, "alias name or linked unit file") {
			stdout, err := utils.Execf("%s systemctl %s ssh", sudo, req.Operation)
			if err != nil {
				return fmt.Errorf("%s ssh(alias name or linked unit file) failed, stdout: %s, err: %v", req.Operation, stdout, err)
			}
		}
		return fmt.Errorf("%s %s failed, stdout: %s, err: %v", req.Operation, serviceName, stdout, err)
	}
	return nil
}

func (u *SSHService) CreateKey(req model.GenerateKey) error {
	// 检查输入是否合法
	if utils.CheckIllegal(req.KeyName, req.EncryptionMode, req.Password) {
		return errors.New(constant.ErrCmdIllegal)
	}

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}

	// 检查目标文件是否已存在
	targetPrivateKey := fmt.Sprintf("%s/.ssh/%s", currentUser.HomeDir, req.KeyName)
	targetPublicKey := fmt.Sprintf("%s/.ssh/%s.pub", currentUser.HomeDir, req.KeyName)

	if _, err := os.Stat(targetPrivateKey); err == nil {
		return fmt.Errorf("private key file already exists: %s", req.KeyName)
	}
	if _, err := os.Stat(targetPublicKey); err == nil {
		return fmt.Errorf("public key file already exists: %s.pub", req.KeyName)
	}

	// 定义文件路径
	timestamp := time.Now().Format("20060102150405")
	secretFile := fmt.Sprintf("%s/.ssh/tmp_%s_%s", currentUser.HomeDir, req.EncryptionMode, timestamp)
	secretPubFile := fmt.Sprintf("%s/.ssh/tmp_%s_%s.pub", currentUser.HomeDir, req.EncryptionMode, timestamp)
	authFile := currentUser.HomeDir + "/.ssh/authorized_keys"

	// 构造注释信息（用户@主机名）
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	comment := fmt.Sprintf("%s@%s", currentUser.Username, hostname)

	// 构造 ssh-keygen 命令
	var command []string
	if len(req.Password) != 0 {
		command = []string{"ssh-keygen", "-t", req.EncryptionMode, "-b", fmt.Sprintf("%d", req.KeyBits), "-N", req.Password, "-C", comment, "-f", secretFile}
	} else {
		command = []string{"ssh-keygen", "-t", req.EncryptionMode, "-b", fmt.Sprintf("%d", req.KeyBits), "-C", comment, "-f", secretFile}
	}

	// 使用 exec.Command 执行 ssh-keygen 命令
	cmd := exec.Command(command[0], command[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("generate failed, err: %v, message: %s", err, stdoutStderr)
	}

	// 确保清理临时文件
	defer func() {
		_ = os.Remove(secretFile)
		_ = os.Remove(secretPubFile)
	}()

	// 如果 authorized_keys 文件不存在，则创建它
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		file, err := os.Create(authFile)
		if err != nil {
			return fmt.Errorf("create authorized_keys failed, err: %v", err)
		}
		defer file.Close()
	}

	// 如果启用密钥，将公钥追加到 authorized_keys 文件
	if req.Enabled {
		pubKey, err := os.ReadFile(secretPubFile)
		if err != nil {
			return fmt.Errorf("read public key file failed, err: %v", err)
		}
		// 使用os.OpenFile打开文件以支持追加模式
		authFileHandle, err := os.OpenFile(authFile, os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("open authorized_keys failed, err: %v", err)
		}
		defer authFileHandle.Close()
		if _, err := authFileHandle.Write(pubKey); err != nil {
			return fmt.Errorf("append public key to authorized_keys failed, err: %v", err)
		}
	}

	// 重命名文件到用户指定的路径（无论是否启用密钥）
	fileOp := files.NewFileOp()
	if err := fileOp.Rename(secretFile, targetPrivateKey); err != nil {
		return fmt.Errorf("failed to rename private key file: %v", err)
	}
	if err := fileOp.Rename(secretPubFile, targetPublicKey); err != nil {
		// 如果公钥重命名失败，尝试回滚私钥的重命名
		_ = fileOp.Rename(targetPrivateKey, secretFile)
		return fmt.Errorf("failed to rename public key file: %v", err)
	}

	return nil
}

func (u *SSHService) ListKeys(req model.ListKey) (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}
	var keys []model.KeyInfo

	// 获取当前用户的目录
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("load current user failed, err: %v", err)
	}
	keyDir := filepath.Join(currentUser.HomeDir, ".ssh")
	authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")

	// 遍历密钥文件目录
	if err := filepath.Walk(keyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// 只处理 .pub 结尾的公钥文件
		if !strings.HasSuffix(info.Name(), ".pub") {
			return nil
		}

		// 查看是否包含keyword
		if req.Keyword != "" && !strings.Contains(info.Name(), req.Keyword) {
			return nil
		}

		// 读取公钥文件内容
		pubKeyData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read public key file failed, err: %v", err)
		}

		// 计算公钥的 SHA256 指纹
		hash := sha256.Sum256(pubKeyData)
		fingerprint := fmt.Sprintf("%x", hash[:])

		// 检查是否存在对应的私钥文件
		privateKeyPath := strings.TrimSuffix(path, ".pub")
		hasPrivateKey := false
		if _, err := os.Stat(privateKeyPath); err == nil {
			hasPrivateKey = true
		}

		// 获取密钥位数
		keyBits := 0
		// 从公钥文件获取位数
		pubKeyParts := strings.Fields(string(pubKeyData))
		if len(pubKeyParts) >= 2 {
			switch pubKeyParts[0] {
			case "ssh-rsa":
				// 解析 base64 编码的公钥数据
				decoded, err := base64.StdEncoding.DecodeString(pubKeyParts[1])
				if err == nil {
					// RSA 公钥的位数可以从解码后的数据中获取
					keyBits = len(decoded) * 8
				}
			case "ssh-ed25519":
				keyBits = 256 // ED25519 总是 256 位
			case "ecdsa-sha2-nistp256":
				keyBits = 256
			case "ecdsa-sha2-nistp384":
				keyBits = 384
			case "ecdsa-sha2-nistp521":
				keyBits = 521
			}
		}

		// 如果从公钥无法获取位数，且存在私钥，则尝试从私钥获取
		if keyBits == 0 && hasPrivateKey {
			privateKeyData, err := os.ReadFile(privateKeyPath)
			if err == nil {
				keyBits, _ = getKeyBits(privateKeyData)
			}
		}

		// 获取用户信息
		user := ""
		// 1. 尝试从公钥注释中获取用户信息
		if len(pubKeyParts) > 2 {
			comment := strings.Join(pubKeyParts[2:], " ")
			if strings.Contains(comment, "@") {
				user = strings.Split(comment, "@")[0]
			}
		}

		// 2. 如果公钥中没有用户信息，且存在私钥，尝试从私钥文件中获取
		if user == "" && hasPrivateKey {
			privateKeyData, err := os.ReadFile(privateKeyPath)
			if err == nil {
				// 解析私钥文件
				block, _ := pem.Decode(privateKeyData)
				if block != nil && strings.Contains(block.Headers["Comment"], "@") {
					user = strings.Split(block.Headers["Comment"], "@")[0]
				}
			}
		}

		// 3. 如果都没有获取到，使用当前系统用户
		if user == "" {
			user = currentUser.Username
		}

		// 获取状态
		status, err := getKeyStatus(authFile, path)
		if err != nil {
			return fmt.Errorf("get key status failed, err: %v", err)
		}

		// 构造密钥信息
		keys = append(keys, model.KeyInfo{
			KeyName:        strings.TrimSuffix(info.Name(), ".pub"),
			Fingerprint:    fingerprint,
			User:           user,
			Status:         status,
			KeyBits:        keyBits,
			HasPrivateKey:  hasPrivateKey,
			PrivateKeyPath: privateKeyPath,
		})

		return nil
	}); err != nil {
		return nil, err
	}

	pageResult.Total = int64(len(keys))
	pageResult.Items = keys

	return &pageResult, nil
}

// EnableKey 操作SSH密钥的启用/禁用状态
func (u *SSHService) EnableKey(req model.EnableKey) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}

	// 构造文件路径
	authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")
	pubKeyFile := filepath.Join(currentUser.HomeDir, ".ssh", req.KeyName+".pub")

	// 读取公钥文件
	pubKeyData, err := os.ReadFile(pubKeyFile)
	if err != nil {
		return fmt.Errorf("read public key file failed, err: %v", err)
	}

	// 读取 authorized_keys 文件
	authData, err := os.ReadFile(authFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read authorized_keys failed, err: %v", err)
	}

	pubKeyStr := string(pubKeyData)
	authStr := string(authData)
	keyExists := strings.Contains(authStr, pubKeyStr)

	if req.Enable && !keyExists {
		// 启用密钥：将公钥追加到 authorized_keys
		authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")

		// 如果文件不存在，创建它
		if _, err := os.Stat(authFile); os.IsNotExist(err) {
			if err := os.WriteFile(authFile, []byte(""), 0600); err != nil {
				return fmt.Errorf("create authorized_keys failed, err: %v", err)
			}
		}

		// 追加公钥
		f, err := os.OpenFile(authFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("open authorized_keys failed, err: %v", err)
		}
		defer f.Close()

		if _, err := f.Write(pubKeyData); err != nil {
			return fmt.Errorf("append public key to authorized_keys failed, err: %v", err)
		}
	} else if !req.Enable && keyExists {
		// 禁用密钥：从 authorized_keys 中移除公钥
		lines := strings.Split(authStr, "\n")
		var newLines []string

		for _, line := range lines {
			if !strings.Contains(line, pubKeyStr) {
				newLines = append(newLines, line)
			}
		}

		// 写回文件
		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(authFile, []byte(newContent), 0600); err != nil {
			return fmt.Errorf("write authorized_keys failed, err: %v", err)
		}
	}

	return nil
}

// DeleteKey 删除SSH密钥
func (u *SSHService) RemoveKey(req model.RemoveKey) error {
	// 确定私钥文件路径
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}

	// 构造文件路径
	privateKeyFile := filepath.Join(currentUser.HomeDir, ".ssh", req.KeyName)
	publicKeyFile := privateKeyFile + ".pub"
	authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")

	// 1. 如果不是只删除私钥,则需要清理authorized_keys中的公钥
	if !req.OnlyPrivateKey {
		// 首先读取公钥文件
		pubKeyData, err := os.ReadFile(publicKeyFile)
		if err == nil { // 只有在能读取公钥文件的情况下才进行清理
			// 读取 authorized_keys 文件
			authData, err := os.ReadFile(authFile)
			if err == nil {
				// 从 authorized_keys 中移除对应的公钥
				lines := strings.Split(string(authData), "\n")
				var newLines []string
				pubKeyStr := string(pubKeyData)

				for _, line := range lines {
					if !strings.Contains(line, pubKeyStr) {
						newLines = append(newLines, line)
					}
				}

				// 写回文件
				newContent := strings.Join(newLines, "\n")
				if err := os.WriteFile(authFile, []byte(newContent), 0600); err != nil {
					return fmt.Errorf("write authorized_keys failed, err: %v", err)
				}
			}
		}
	}

	// 2. 删除私钥文件
	if err := os.Remove(privateKeyFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove private key file failed, err: %v", err)
	}

	// 3. 如果不是只删除私钥,则删除公钥文件
	if !req.OnlyPrivateKey {
		if err := os.Remove(publicKeyFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove public key file failed, err: %v", err)
		}
	}

	return nil
}

// SetKeyPassword 设置私钥密码（仅当私钥未设置密码时可用）
func (u *SSHService) SetKeyPassword(req model.SetKeyPassword) error {
	// 确定私钥文件路径
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	privateKeyFile := filepath.Join(currentUser.HomeDir, ".ssh", req.KeyName)

	// 读取私钥文件
	privateKeyData, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return fmt.Errorf("read private key file failed: %v", err)
	}

	// 尝试不带密码解析私钥
	_, err = ssh.ParsePrivateKey(privateKeyData)
	if err != nil {
		if strings.Contains(err.Error(), "private key is passphrase protected") {
			return fmt.Errorf("private key already has a password")
		}
		return fmt.Errorf("parse private key failed: %v", err)
	}

	// 使用ssh-keygen更改密码
	cmd := exec.Command("ssh-keygen", "-p", "-f", privateKeyFile, "-N", req.Password)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("set password failed: %s, %v", output, err)
	}

	return nil
}

// ChangeKeyPassword 修改私钥密码
func (u *SSHService) ChangeKeyPassword(req model.UpdateKeyPassword) error {
	// 确定私钥文件路径
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}
	privateKeyFile := filepath.Join(currentUser.HomeDir, ".ssh", req.KeyName)

	// 读取私钥文件
	privateKeyData, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return fmt.Errorf("read private key file failed: %v", err)
	}

	// 尝试使用旧密码解析私钥以验证密码
	_, err = ssh.ParsePrivateKeyWithPassphrase(privateKeyData, []byte(req.OldPassword))
	if err != nil {
		return fmt.Errorf("invalid old password: %v", err)
	}

	// 使用ssh-keygen更改密码
	cmd := exec.Command("ssh-keygen", "-p", "-f", privateKeyFile, "-P", req.OldPassword, "-N", req.NewPassword)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("change password failed: %s, %v", output, err)
	}

	return nil
}

// ClearKeyPassword 清除私钥密码
func (u *SSHService) ClearKeyPassword(req model.SetKeyPassword) error {
	// 直接调用 ChangeKeyPassword，将新密码设置为空字符串
	return u.ChangeKeyPassword(model.UpdateKeyPassword{
		KeyName:     req.KeyName,
		OldPassword: req.Password,
		NewPassword: "", // 清除密码就是将新密码设置为空
	})
}

// ListAuthKeys 列出authorized_keys中的所有公钥
func (u *SSHService) ListAuthKeys() (*model.PageResult, error) {
	var pageResult = model.PageResult{Total: 0, Items: nil}
	var keys []model.AuthKeyInfo

	// 获取当前用户的目录
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("load current user failed, err: %v", err)
	}

	// 读取authorized_keys文件
	authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")
	content, err := os.ReadFile(authFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，返回空结果
			return &pageResult, nil
		}
		return nil, fmt.Errorf("read authorized_keys failed, err: %v", err)
	}

	// 按行解析公钥
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析公钥行
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			keyInfo := model.AuthKeyInfo{
				Algorithm: parts[0], // 算法类型
				Key:       parts[1], // 密钥内容
				Comment:   "",       // 描述/注释
			}

			// 如果有注释部分
			if len(parts) > 2 {
				keyInfo.Comment = strings.Join(parts[2:], " ")
			}

			keys = append(keys, keyInfo)
		}
	}

	pageResult.Total = int64(len(keys))
	pageResult.Items = keys
	return &pageResult, nil
}

// AddAuthKey 添加公钥到authorized_keys
func (u *SSHService) AddAuthKey(req model.AddAuthKey) error {
	// 获取当前用户的目录
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}

	authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")

	// 确保.ssh目录存在
	sshDir := filepath.Join(currentUser.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("create .ssh directory failed, err: %v", err)
	}

	// 构造公钥行
	keyLine := fmt.Sprintf("%s\n", strings.TrimSpace(req.Content))

	// 如果文件不存在，创建新文件
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		if err := os.WriteFile(authFile, []byte(keyLine), 0600); err != nil {
			return fmt.Errorf("create authorized_keys failed, err: %v", err)
		}
		return nil
	}

	// 检查是否已存在相同的密钥
	content, err := os.ReadFile(authFile)
	if err != nil {
		return fmt.Errorf("read authorized_keys failed, err: %v", err)
	}

	if strings.Contains(string(content), req.Content) {
		return fmt.Errorf("public key already exists")
	}

	// 追加新的公钥
	f, err := os.OpenFile(authFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open authorized_keys failed, err: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(keyLine); err != nil {
		return fmt.Errorf("append public key failed, err: %v", err)
	}

	return nil
}

// RemoveAuthKey 从authorized_keys中删除公钥
func (u *SSHService) RemoveAuthKey(req model.RemoveAuthKey) error {
	// 获取当前用户的目录
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("load current user failed, err: %v", err)
	}

	authFile := filepath.Join(currentUser.HomeDir, ".ssh", "authorized_keys")

	// 读取authorized_keys文件
	content, err := os.ReadFile(authFile)
	if err != nil {
		return fmt.Errorf("read authorized_keys failed, err: %v", err)
	}

	// 按行处理，排除要删除的公钥
	lines := strings.Split(string(content), "\n")
	var newLines []string
	keyFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 如果该行不包含要删除的公钥，保留该行
		if !strings.Contains(line, req.Content) {
			newLines = append(newLines, line)
		} else {
			keyFound = true
		}
	}

	if !keyFound {
		return fmt.Errorf("public key not found")
	}

	// 写回文件
	newContent := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(authFile, []byte(newContent), 0600); err != nil {
		return fmt.Errorf("write authorized_keys failed, err: %v", err)
	}

	return nil
}

// getKeyStatus 检查密钥是否存在于 authorized_keys 文件中
func getKeyStatus(authFile, keyFile string) (string, error) {
	// 读取 authorized_keys 文件内容
	authData, err := os.ReadFile(authFile)
	if err != nil {
		return "", fmt.Errorf("read authorized_keys failed, err: %v", err)
	}

	// 读取密钥文件内容
	keyData, err := os.ReadFile(keyFile + ".pub")
	if err != nil {
		return "", fmt.Errorf("read public key file failed, err: %v", err)
	}

	// 判断密钥是否存在于 authorized_keys 中
	if strings.Contains(string(authData), string(keyData)) {
		return "enabled", nil
	}
	return "disabled", nil
}

// getKeyBits 获取密钥的位数（根据密钥格式）
func getKeyBits(fileData []byte) (int, error) {
	block, _ := pem.Decode(fileData)
	if block == nil {
		return 0, fmt.Errorf("invalid PEM format")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return 0, fmt.Errorf("parse RSA private key failed: %v", err)
		}
		return key.N.BitLen(), nil

	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return 0, fmt.Errorf("parse EC private key failed: %v", err)
		}
		return key.Params().BitSize, nil

	case "OPENSSH PRIVATE KEY":
		// 尝试解析 OpenSSH 格式的私钥
		key, err := ssh.ParseRawPrivateKey(fileData)
		if err != nil {
			return 0, fmt.Errorf("parse OpenSSH private key failed: %v", err)
		}

		// 根据具体类型返回位数
		switch k := key.(type) {
		case *rsa.PrivateKey:
			return k.Size() * 8, nil
		case *ed25519.PrivateKey:
			return 256, nil // Ed25519 总是 256 位
		default:
			return 0, fmt.Errorf("unsupported OpenSSH key type")
		}

	default:
		return 0, fmt.Errorf("unsupported key type: %s", block.Type)
	}
}

type sshFileItem struct {
	Name string
	Year int
}

func (u *SSHService) LoadLog(req model.SearchSSHLog) (*model.SSHLog, error) {
	var fileList []sshFileItem
	var data model.SSHLog
	baseDir := "/var/log"
	if err := filepath.Walk(baseDir, func(pathItem string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasPrefix(info.Name(), "secure") || strings.HasPrefix(info.Name(), "auth")) {
			if !strings.HasSuffix(info.Name(), ".gz") {
				fileList = append(fileList, sshFileItem{Name: pathItem, Year: info.ModTime().Year()})
				return nil
			}
			itemFileName := strings.TrimSuffix(pathItem, ".gz")
			if _, err := os.Stat(itemFileName); err != nil && os.IsNotExist(err) {
				if err := handleGunzip(pathItem); err == nil {
					fileList = append(fileList, sshFileItem{Name: itemFileName, Year: info.ModTime().Year()})
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	fileList = sortFileList(fileList)

	command := ""
	if len(req.Info) != 0 {
		command = fmt.Sprintf(" | grep '%s'", req.Info)
	}

	showCountFrom := (req.Page - 1) * req.PageSize
	showCountTo := req.Page * req.PageSize
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	for _, file := range fileList {
		commandItem := ""
		if strings.HasPrefix(path.Base(file.Name), "secure") {
			switch req.Status {
			case constant.StatusSuccess:
				commandItem = fmt.Sprintf("cat %s | grep -a Accepted %s", file.Name, command)
			case constant.StatusFailed:
				commandItem = fmt.Sprintf("cat %s | grep -a 'Failed password for' %s", file.Name, command)
			default:
				commandItem = fmt.Sprintf("cat %s | grep -aE '(Failed password for|Accepted)' %s", file.Name, command)
			}
		}
		if strings.HasPrefix(path.Base(file.Name), "auth.log") {
			switch req.Status {
			case constant.StatusSuccess:
				commandItem = fmt.Sprintf("cat %s | grep -a Accepted %s", file.Name, command)
			case constant.StatusFailed:
				commandItem = fmt.Sprintf("cat %s | grep -aE 'Failed password for|Connection closed by authenticating user' %s", file.Name, command)
			default:
				commandItem = fmt.Sprintf("cat %s | grep -aE \"(Failed password for|Connection closed by authenticating user|Accepted)\" %s", file.Name, command)
			}
		}
		dataItem, successCount, failedCount := loadSSHData(commandItem, showCountFrom, showCountTo, file.Year, nyc)
		data.FailedCount += failedCount
		data.TotalCount += successCount + failedCount
		showCountFrom = showCountFrom - (successCount + failedCount)
		showCountTo = showCountTo - (successCount + failedCount)
		data.Logs = append(data.Logs, dataItem...)
	}

	data.SuccessfulCount = data.TotalCount - data.FailedCount
	return &data, nil
}

func sortFileList(fileNames []sshFileItem) []sshFileItem {
	if len(fileNames) < 2 {
		return fileNames
	}
	if strings.HasPrefix(path.Base(fileNames[0].Name), "secure") {
		var itemFile []sshFileItem
		sort.Slice(fileNames, func(i, j int) bool {
			return fileNames[i].Name > fileNames[j].Name
		})
		itemFile = append(itemFile, fileNames[len(fileNames)-1])
		itemFile = append(itemFile, fileNames[:len(fileNames)-1]...)
		return itemFile
	}
	sort.Slice(fileNames, func(i, j int) bool {
		return fileNames[i].Name < fileNames[j].Name
	})
	return fileNames
}

func updateSSHConf(oldFiles []string, req model.SSHUpdate) []string {
	var newFiles []string
	for _, v := range req.Values {
		param := v.Key
		value := v.NewValue

		var valueItems []string
		if param != "ListenAddress" {
			valueItems = append(valueItems, value)
		} else {
			if value != "" {
				valueItems = strings.Split(value, ",")
			}
		}

		for _, line := range oldFiles {
			lineItem := strings.TrimSpace(line)
			if (strings.HasPrefix(lineItem, param) || strings.HasPrefix(lineItem, fmt.Sprintf("#%s", param))) && len(valueItems) != 0 {
				newFiles = append(newFiles, fmt.Sprintf("%s %s", param, valueItems[0]))
				valueItems = valueItems[1:]
				continue
			}
			if strings.HasPrefix(lineItem, param) && len(valueItems) == 0 {
				newFiles = append(newFiles, fmt.Sprintf("#%s", line))
				continue
			}
			newFiles = append(newFiles, line)
		}
		if len(valueItems) != 0 {
			for _, item := range valueItems {
				newFiles = append(newFiles, fmt.Sprintf("%s %s", param, item))
			}
		}
	}

	return newFiles
}

func loadSSHData(command string, showCountFrom, showCountTo, currentYear int, nyc *time.Location) ([]model.SSHHistory, int, int) {
	var (
		datas        []model.SSHHistory
		successCount int
		failedCount  int
	)
	stdout2, err := utils.Exec(command)
	if err != nil {
		return datas, 0, 0
	}
	lines := strings.Split(string(stdout2), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		var itemData model.SSHHistory
		switch {
		case strings.Contains(lines[i], "Failed password for"):
			itemData = loadFailedSecureDatas(lines[i])
			if len(itemData.Address) != 0 {
				if successCount+failedCount >= showCountFrom && successCount+failedCount < showCountTo {
					itemData.Area = "unknown"
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				failedCount++
			}
		case strings.Contains(lines[i], "Connection closed by authenticating user"):
			itemData = loadFailedAuthDatas(lines[i])
			if len(itemData.Address) != 0 {
				if successCount+failedCount >= showCountFrom && successCount+failedCount < showCountTo {
					itemData.Area = "unknown"
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				failedCount++
			}
		case strings.Contains(lines[i], "Accepted "):
			itemData = loadSuccessDatas(lines[i])
			if len(itemData.Address) != 0 {
				if successCount+failedCount >= showCountFrom && successCount+failedCount < showCountTo {
					itemData.Area = "unknown"
					itemData.Date = loadDate(currentYear, itemData.DateStr, nyc)
					datas = append(datas, itemData)
				}
				successCount++
			}
		}
	}
	return datas, successCount, failedCount
}

func loadSuccessDatas(line string) model.SSHHistory {
	var data model.SSHHistory
	parts := strings.Fields(line)
	index, dataStr := analyzeDateStr(parts)
	if dataStr == "" {
		return data
	}
	data.DateStr = dataStr
	data.AuthMode = parts[4+index]
	data.User = parts[6+index]
	data.Address = parts[8+index]
	data.Port = parts[10+index]
	data.Status = constant.StatusSuccess
	return data
}

func loadFailedAuthDatas(line string) model.SSHHistory {
	var data model.SSHHistory
	parts := strings.Fields(line)
	index, dataStr := analyzeDateStr(parts)
	if dataStr == "" {
		return data
	}
	data.DateStr = dataStr
	if index == 2 {
		data.User = parts[10]
	} else {
		data.User = parts[7]
	}
	data.AuthMode = parts[6+index]
	data.Address = parts[9+index]
	data.Port = parts[11+index]
	data.Status = constant.StatusFailed
	if strings.Contains(line, ": ") {
		data.Message = strings.Split(line, ": ")[1]
	}
	return data
}
func loadFailedSecureDatas(line string) model.SSHHistory {
	var data model.SSHHistory
	parts := strings.Fields(line)
	index, dataStr := analyzeDateStr(parts)
	if dataStr == "" {
		return data
	}
	data.DateStr = dataStr
	if strings.Contains(line, " invalid ") {
		data.AuthMode = parts[4+index]
		index += 2
	} else {
		data.AuthMode = parts[4+index]
	}
	data.User = parts[6+index]
	data.Address = parts[8+index]
	data.Port = parts[10+index]
	data.Status = constant.StatusFailed
	if strings.Contains(line, ": ") {
		data.Message = strings.Split(line, ": ")[1]
	}
	return data
}

func handleGunzip(path string) error {
	if _, err := utils.Execf("gunzip %s", path); err != nil {
		return err
	}
	return nil
}

func loadServiceName() (string, error) {
	if exist, _ := systemctl.IsExist("sshd"); exist {
		return "sshd", nil
	} else if exist, _ := systemctl.IsExist("ssh"); exist {
		return "ssh", nil
	}
	return "", errors.New("the ssh or sshd service is unavailable")
}

func loadDate(currentYear int, DateStr string, nyc *time.Location) time.Time {
	itemDate, err := time.ParseInLocation("2006 Jan 2 15:04:05", fmt.Sprintf("%d %s", currentYear, DateStr), nyc)
	if err != nil {
		itemDate, _ = time.ParseInLocation("2006 Jan 2 15:04:05", DateStr, nyc)
	}
	return itemDate
}

func analyzeDateStr(parts []string) (int, string) {
	t, err := time.Parse("2006-01-02T15:04:05.999999-07:00", parts[0])
	if err != nil {
		if len(parts) < 14 {
			return 0, ""
		}
		return 2, fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
	}
	if len(parts) < 12 {
		return 0, ""
	}
	return 0, t.Format("2006 Jan 2 15:04:05")
}
