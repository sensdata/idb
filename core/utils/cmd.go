package utils

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sensdata/idb/core/constant"
)

func Exec(cmdStr string) (string, error) {
	return ExecWithTimeOut(cmdStr, 20*time.Second)
}

func handleErr(stdout, stderr bytes.Buffer, err error) (string, error) {
	errMsg := ""
	if len(stderr.String()) != 0 {
		errMsg = fmt.Sprintf("stderr: %s", stderr.String())
	}
	if len(stdout.String()) != 0 {
		if len(errMsg) != 0 {
			errMsg = fmt.Sprintf("%s; stdout: %s", errMsg, stdout.String())
		} else {
			errMsg = fmt.Sprintf("stdout: %s", stdout.String())
		}
	}
	return errMsg, err
}

func ExecWithTimeOut(cmdStr string, timeout time.Duration) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return "", errors.New(constant.ErrCmdNotFound)
	case err := <-done:
		if err != nil {
			return handleErr(stdout, stderr, err)
		}
	}

	return stdout.String(), nil
}

// ExecContainerScript 安全地在容器中执行脚本
// 使用参数化方式执行，防止命令注入
func ExecContainerScript(containerName, cmdStr string, timeout time.Duration) error {
	// 验证容器名，防止命令注入
	if err := ValidateContainerName(containerName); err != nil {
		return fmt.Errorf("invalid container name: %w", err)
	}

	// 验证命令字符串，防止命令注入
	// 检查是否包含危险字符
	if CheckIllegal(cmdStr) {
		return fmt.Errorf("command string contains illegal characters")
	}

	// 使用参数化方式执行 docker exec
	// docker exec -i <container> bash -c '<command>'
	// 注意：由于需要通过 bash -c 执行，我们需要将命令作为单个参数传递
	// 但我们已经验证了 cmdStr 不包含危险字符，所以相对安全
	cmd := exec.Command("docker", "exec", "-i", containerName, "bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return fmt.Errorf("command timeout after %v", timeout)
	case err := <-done:
		if err != nil {
			errMsg := ""
			if len(stderr.String()) != 0 {
				errMsg = fmt.Sprintf("stderr: %s", stderr.String())
			}
			if len(stdout.String()) != 0 {
				if errMsg != "" {
					errMsg = fmt.Sprintf("%s; stdout: %s", errMsg, stdout.String())
				} else {
					errMsg = fmt.Sprintf("stdout: %s", stdout.String())
				}
			}
			if errMsg != "" {
				return fmt.Errorf("%s; err: %v", errMsg, err)
			}
			return err
		}
	}

	return nil
}

// ExecCronjobWithTimeOut 安全地执行定时任务命令
// 使用参数化方式执行，防止命令注入
func ExecCronjobWithTimeOut(cmdStr, workdir, outPath string, timeout time.Duration) error {
	// 验证命令字符串，防止命令注入
	if CheckIllegal(cmdStr) {
		return fmt.Errorf("command string contains illegal characters")
	}

	// 验证工作目录路径，防止路径遍历攻击
	if workdir != "" {
		// 检查是否包含路径遍历字符
		if strings.Contains(workdir, "..") {
			return fmt.Errorf("workdir cannot contain '..'")
		}
		// 验证目录是否存在
		if _, err := os.Stat(workdir); err != nil {
			return fmt.Errorf("workdir does not exist: %v", err)
		}
	}

	// 验证输出文件路径，防止路径遍历攻击
	if outPath != "" {
		// 检查是否包含路径遍历字符
		if strings.Contains(outPath, "..") {
			return fmt.Errorf("outPath cannot contain '..'")
		}
		// 确保输出目录存在
		outDir := ""
		if lastSlash := strings.LastIndex(outPath, "/"); lastSlash != -1 {
			outDir = outPath[:lastSlash]
		}
		if outDir != "" {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %v", err)
			}
		}
	}

	// 打开输出文件
	file, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to open output file: %v", err)
	}
	defer file.Close()

	// 使用参数化方式执行命令
	// 注意：由于需要通过 bash -c 执行复杂命令，我们仍然使用 bash -c
	// 但已经通过 CheckIllegal 验证了 cmdStr 不包含危险字符
	cmd := exec.Command("bash", "-c", cmdStr)
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Stdout = file
	cmd.Stderr = file

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return errors.New(constant.ErrCmdTimeout)
	case err := <-done:
		if err != nil {
			return fmt.Errorf("command execution failed: %v", err)
		}
	}
	return nil
}

func Execf(cmdStr string, a ...interface{}) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdStr, a...))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return handleErr(stdout, stderr, err)
	}
	return stdout.String(), nil
}

func ExecWithCheck(name string, a ...string) (string, error) {
	cmd := exec.Command(name, a...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return handleErr(stdout, stderr, err)
	}
	return stdout.String(), nil
}

func ExecScript(scriptPath, workDir string) (string, error) {
	cmd := exec.Command("bash", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Dir = workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(10 * time.Minute)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return "", errors.New(constant.ErrCmdTimeout)
	case err := <-done:
		if err != nil {
			return handleErr(stdout, stderr, err)
		}
	}

	return stdout.String(), nil
}

func ExecCmd(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error : %v, output: %s", err, output)
	}
	return nil
}

func ExecCmdWithDir(cmdStr, workDir string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error : %v, output: %s", err, output)
	}
	return nil
}

func CheckIllegal(args ...string) bool {
	if args == nil {
		return false
	}
	for _, arg := range args {
		if strings.Contains(arg, "&") || strings.Contains(arg, "|") || strings.Contains(arg, ";") ||
			strings.Contains(arg, "$") || strings.Contains(arg, "'") || strings.Contains(arg, "`") ||
			strings.Contains(arg, "(") || strings.Contains(arg, ")") || strings.Contains(arg, "\"") ||
			strings.Contains(arg, "\n") || strings.Contains(arg, "\r") || strings.Contains(arg, ">") || strings.Contains(arg, "<") {
			return true
		}
	}
	return false
}

func HasNoPasswordSudo() bool {
	cmd2 := exec.Command("sudo", "-n", "ls")
	err2 := cmd2.Run()
	return err2 == nil
}

func SudoHandleCmd() string {
	cmd := exec.Command("sudo", "-n", "ls")
	if err := cmd.Run(); err == nil {
		return "sudo "
	}
	return ""
}

func Which(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ValidateHostname 验证主机名格式
// 主机名规则：
// - 只能包含字母、数字、连字符、下划线、点
// - 长度 1-253 字符
// - 不能以点或连字符开头/结尾
// - 不能有连续的点
func ValidateHostname(hostname string) error {
	if len(hostname) == 0 {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}
	if hostname[0] == '-' || hostname[0] == '.' || hostname[len(hostname)-1] == '-' || hostname[len(hostname)-1] == '.' {
		return fmt.Errorf("hostname cannot start or end with '-' or '.'")
	}
	if strings.Contains(hostname, "..") {
		return fmt.Errorf("hostname cannot contain consecutive dots")
	}
	// 检查是否只包含允许的字符
	for _, r := range hostname {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return fmt.Errorf("hostname contains invalid character: %c", r)
		}
	}
	return nil
}

// ExecCmdSafe 安全执行命令，使用参数化方式，避免命令注入
// 第一个参数是命令名，后续参数是命令参数
func ExecCmdSafe(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, output)
	}
	return nil
}

// ExecCmdSafeWithSudo 安全执行需要 sudo 的命令
func ExecCmdSafeWithSudo(name string, args ...string) error {
	// 检查是否有无密码 sudo
	hasSudo := HasNoPasswordSudo()
	if !hasSudo {
		return fmt.Errorf("passwordless sudo is required")
	}
	// 使用 sudo 执行命令
	cmdArgs := append([]string{name}, args...)
	cmd := exec.Command("sudo", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, output)
	}
	return nil
}

// WriteToFileSafe 安全地将内容写入文件（使用 sudo tee）
func WriteToFileSafe(content, filePath string) error {
	hasSudo := HasNoPasswordSudo()
	if !hasSudo {
		return fmt.Errorf("passwordless sudo is required")
	}
	cmd := exec.Command("sudo", "tee", filePath)
	cmd.Stdin = strings.NewReader(content)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, output)
	}
	// 忽略输出，只检查错误
	_ = output
	return nil
}

// ValidateIPAddress 验证 IP 地址格式（支持 IPv4 和 IPv6）
func ValidateIPAddress(ip string) error {
	if ip == "" {
		return fmt.Errorf("IP address cannot be empty")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}
	return nil
}

// ValidateDNSServers 验证 DNS 服务器列表
// 限制：最多 10 个服务器，每个必须是有效的 IP 地址
func ValidateDNSServers(servers []string) error {
	if len(servers) == 0 {
		return fmt.Errorf("at least one DNS server is required")
	}
	if len(servers) > 10 {
		return fmt.Errorf("too many DNS servers (max 10)")
	}
	for i, server := range servers {
		if server == "" {
			return fmt.Errorf("DNS server at index %d cannot be empty", i)
		}
		if err := ValidateIPAddress(server); err != nil {
			return fmt.Errorf("DNS server at index %d: %w", i, err)
		}
	}
	return nil
}

// ValidateTimezone 验证时区格式
// 时区规则：
// - 只能包含字母、数字、斜杠、下划线、连字符
// - 不能包含路径遍历字符（..）
// - 不能以斜杠开头或结尾
// - 长度限制
func ValidateTimezone(timezone string) error {
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	if len(timezone) > 100 {
		return fmt.Errorf("timezone too long (max 100 characters)")
	}
	// 防止路径遍历攻击
	if strings.Contains(timezone, "..") {
		return fmt.Errorf("timezone cannot contain '..'")
	}
	// 不能以斜杠开头或结尾
	if strings.HasPrefix(timezone, "/") || strings.HasSuffix(timezone, "/") {
		return fmt.Errorf("timezone cannot start or end with '/'")
	}
	// 检查是否只包含允许的字符
	for _, r := range timezone {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '/' || r == '_' || r == '-') {
			return fmt.Errorf("timezone contains invalid character: %c", r)
		}
	}
	// 验证时区文件是否存在（防止无效时区）
	timezonePath := fmt.Sprintf("/usr/share/zoneinfo/%s", timezone)
	if _, err := os.Stat(timezonePath); err != nil {
		return fmt.Errorf("timezone file not found: %s", timezonePath)
	}
	return nil
}

// ValidateContainerName 验证 Docker 容器名格式
// Docker 容器名规则：
// - 只能包含字母、数字、下划线、连字符、点
// - 不能包含特殊字符和空格
// - 长度限制（Docker 限制为 1-255 字符）
func ValidateContainerName(containerName string) error {
	if containerName == "" {
		return fmt.Errorf("container name cannot be empty")
	}
	if len(containerName) > 255 {
		return fmt.Errorf("container name too long (max 255 characters)")
	}
	// 检查是否只包含允许的字符（Docker 容器名允许的字符）
	for _, r := range containerName {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '_' || r == '-' || r == '.') {
			return fmt.Errorf("container name contains invalid character: %c", r)
		}
	}
	return nil
}
