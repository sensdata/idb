package pkg

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sensdata/idb/agent/global"
)

var (
	rsyncArgs     []string
	rsyncArgsOnce sync.Once
)

// getRsyncBaseArgs 根据rsync版本选择最佳参数
func getRsyncBaseArgs() []string {
	rsyncArgsOnce.Do(func() {
		cmd := exec.Command("rsync", "--version")
		out, err := cmd.Output()
		if err == nil {
			// 解析版本号
			versionLine := strings.SplitN(string(out), "\n", 2)[0]
			// rsync  version 3.1.3  protocol version 31
			re := regexp.MustCompile(`version\s+(\d+)\.(\d+)\.(\d+)`)
			m := re.FindStringSubmatch(versionLine)
			if len(m) == 4 {
				major, _ := strconv.Atoi(m[1])
				minor, _ := strconv.Atoi(m[2])
				if major > 3 || (major == 3 && minor >= 1) {
					// >= 3.1 支持 --info=progress2
					rsyncArgs = []string{"-azv", "--info=progress2", "--partial", "--stats"}
					return
				}
			}
		}
		// fallback - 对于旧版本使用传统进度显示
		rsyncArgs = []string{"-azv", "--progress", "--partial", "--stats"}
	})
	return rsyncArgs
}

// shQuote 安全的shell引号函数，处理特殊字符
func shQuote(s string) string {
	if s == "" {
		return "''"
	}
	// replace ' with '\'' (close, escaped quote, reopen)
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// pickSSHCommand 构建SSH命令并返回可选的包装器
func pickSSHCommand(t *RsyncTask) (sshCmd string, wrapper string) {
	if t.RemoteType != RemoteTypeSSH {
		return "", ""
	}

	parts := []string{"ssh"}
	if t.RemotePort != 0 {
		parts = append(parts, fmt.Sprintf("-p %d", t.RemotePort))
	}

	// 根据认证模式处理
	switch t.AuthMode {
	case AuthModePrivateKey:
		if t.SSHPrivateKey != "" {
			parts = append(parts, fmt.Sprintf("-i %s", t.SSHPrivateKey))
		}
	case AuthModePassword:
		if t.Password != "" {
			// 使用sshpass包装器，密码安全引号
			wrapper = fmt.Sprintf("sshpass -p %s", shQuote(t.Password))
		}
	}

	// 用户名
	if t.Username != "" {
		parts = append(parts, fmt.Sprintf("-l %s", t.Username))
	}

	// 避免交互式主机密钥提示
	parts = append(parts, "-o StrictHostKeyChecking=no", "-o UserKnownHostsFile=/dev/null")

	sshCmd = strings.Join(parts, " ")
	return sshCmd, wrapper
}

// ExecProcess wraps a running command so we can stop it
type ExecProcess struct {
	cmd    *exec.Cmd
	lock   sync.Mutex
	cancel context.CancelFunc
}

func (p *ExecProcess) Stop() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("no process")
	}

	// 先取消context
	if p.cancel != nil {
		p.cancel()
	}

	// try polite kill then force
	_ = p.cmd.Process.Signal(syscall.SIGINT)
	time.Sleep(200 * time.Millisecond)

	err := p.cmd.Process.Kill()

	// 确保进程完全退出
	if p.cmd.ProcessState == nil {
		go func() {
			_, _ = p.cmd.Process.Wait()
		}()
	}

	return err
}

// buildRsyncCommand builds rsync command according to task
func buildRsyncCommand(t *RsyncTask) ([]string, error) {
	// 使用优化的基础参数
	args := append([]string{}, getRsyncBaseArgs()...)
	var src, dst string

	switch t.RemoteType {
	case RemoteTypeSSH:
		// 使用优化的SSH命令构建
		sshCmd, wrapper := pickSSHCommand(t)

		if sshCmd != "" {
			// 添加SSH命令参数，使用安全引号
			args = append(args, "-e", shQuote(sshCmd))
		}

		// 构建源和目标路径
		if t.Direction == DirectionLocalToRemote {
			src = t.LocalPath
			dst = fmt.Sprintf("%s@%s:%s", t.Username, t.RemoteHost, t.RemotePath)
		} else {
			src = fmt.Sprintf("%s@%s:%s", t.Username, t.RemoteHost, t.RemotePath)
			dst = t.LocalPath
		}

		// 如果存在包装器（sshpass），需要特殊处理
		if wrapper != "" {
			// 对于密码认证，我们使用环境变量方式，这里保持兼容性
			// 实际密码传递在StartRsync函数中通过环境变量处理
		}

	case RemoteTypeRsync:
		// Rsync守护进程模式
		var remoteURI string

		switch t.AuthMode {
		case AuthModePassword:
			// 密码认证：在URI中包含用户名
			if t.RemotePort != 0 {
				remoteURI = fmt.Sprintf("rsync://%s@%s:%d/%s/%s", t.Username, t.RemoteHost, t.RemotePort, strings.Trim(t.Module, "/"), strings.Trim(t.RemotePath, "/"))
			} else {
				remoteURI = fmt.Sprintf("rsync://%s@%s/%s/%s", t.Username, t.RemoteHost, strings.Trim(t.Module, "/"), strings.Trim(t.RemotePath, "/"))
			}

		case AuthModeAnonymous:
			// 匿名认证：不包含用户名
			if t.RemotePort != 0 {
				remoteURI = fmt.Sprintf("rsync://%s:%d/%s/%s", t.RemoteHost, t.RemotePort, strings.Trim(t.Module, "/"), strings.Trim(t.RemotePath, "/"))
			} else {
				remoteURI = fmt.Sprintf("rsync://%s/%s/%s", t.RemoteHost, strings.Trim(t.Module, "/"), strings.Trim(t.RemotePath, "/"))
			}

		default:
			return nil, fmt.Errorf("unsupported authentication mode for rsync daemon: %s", t.AuthMode)
		}

		if t.Direction == DirectionLocalToRemote {
			src = t.LocalPath
			dst = remoteURI
		} else {
			src = remoteURI
			dst = t.LocalPath
		}

	default:
		return nil, fmt.Errorf("unsupported remote type: %s", t.RemoteType)
	}

	args = append(args, src, dst)
	return args, nil
}

// StartRsync starts rsync as a subprocess, returns ExecProcess and exit/error
func StartRsync(t *RsyncTask) (*ExecProcess, error) {
	args, err := buildRsyncCommand(t)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to build rsync command for task %s: %v", t.ID, err)
		return nil, err
	}

	global.LOG.Info("[rsyncmgr] starting rsync for task %s: %s", t.ID, strings.Join(args, " "))

	// prepare command
	ctx, cancel := context.WithCancel(context.Background())

	// 检查是否需要使用sshpass包装器
	sshCmd, wrapper := pickSSHCommand(t)
	global.LOG.Info("[rsyncmgr] sshCmd: %s, wrapper: %s", sshCmd, wrapper)

	var cmd *exec.Cmd
	if wrapper != "" && sshCmd != "" {
		// 使用sshpass包装器执行命令（-p参数方式）
		// 构建完整的shell命令：sshpass -p 'password' rsync [args]
		fullCmd := fmt.Sprintf("%s rsync %s", wrapper, strings.Join(args, " "))
		cmd = exec.CommandContext(ctx, "sh", "-c", fullCmd)
	} else {
		// 直接执行rsync命令（适用于Rsync守护进程模式）
		cmd = exec.CommandContext(ctx, "rsync", args...)
	}

	// optional: redirect stdout/stderr to logs or buffer
	// Here we let stdout/stderr inherit to agent's process for visibility (or you can capture).
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		cancel()
		global.LOG.Error("[rsyncmgr] failed to start rsync for task %s: %v", t.ID, err)
		return nil, err
	}
	proc := &ExecProcess{cmd: cmd, cancel: cancel}
	return proc, nil
}
