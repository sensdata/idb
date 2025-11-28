package pkg

import (
	"context"
	"fmt"
	"os"
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
		parts = append(parts, "-p", fmt.Sprintf("%d", t.RemotePort)) // 修复拼接
	}

	// 根据认证模式处理
	switch t.AuthMode {
	case AuthModePrivateKey:
		if t.SSHPrivateKey != "" {
			parts = append(parts, "-i", t.SSHPrivateKey)
		}
	case AuthModePassword:
		if t.Password != "" {
			// 使用sshpass包装器，密码安全引号
			wrapper = fmt.Sprintf("sshpass -p %s", shQuote(t.Password))
		}
	}

	parts = append(parts,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	)

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
func buildRsyncCommand(t *RsyncTask) ([]string, string, string, error) {
	// 使用优化的基础参数
	args := append([]string{}, getRsyncBaseArgs()...)
	var src, dst string
	var sshCmd, wrapper string

	switch t.RemoteType {
	case RemoteTypeSSH:
		// 使用优化的SSH命令构建
		sshCmd, wrapper = pickSSHCommand(t)

		if sshCmd != "" {
			// 直接传递 ssh 命令字符串（不要加外层引号）——exec.Command 不需要 shell 引号
			args = append(args, "-e", sshCmd)
		}

		// 构建源和目标路径
		if t.Direction == DirectionLocalToRemote {
			src = t.LocalPath
			dst = fmt.Sprintf("%s@%s:%s", t.Username, t.RemoteHost, t.RemotePath)
		} else {
			src = fmt.Sprintf("%s@%s:%s", t.Username, t.RemoteHost, t.RemotePath)
			dst = t.LocalPath
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
			return nil, "", "", fmt.Errorf("unsupported authentication mode for rsync daemon: %s", t.AuthMode)
		}

		if t.Direction == DirectionLocalToRemote {
			src = t.LocalPath
			dst = remoteURI
		} else {
			src = remoteURI
			dst = t.LocalPath
		}

	default:
		return nil, "", "", fmt.Errorf("unsupported remote type: %s", t.RemoteType)
	}

	args = append(args, src, dst)
	return args, sshCmd, wrapper, nil
}

// StartRsync starts rsync as a subprocess, returns ExecProcess and exit/error
func StartRsync(t *RsyncTask, logHandler *LogHandler) (*ExecProcess, error) {
	args, sshCmd, wrapper, err := buildRsyncCommand(t)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to build rsync command for task %s: %v", t.ID, err)
		return nil, err
	}
	global.LOG.Info("[rsyncmgr] args: %s, sshCmd: %s, wrapper: %s", strings.Join(args, " "), sshCmd, wrapper)

	// prepare command
	var cmd *exec.Cmd
	ctx, cancel := context.WithCancel(context.Background())

	// 根据远程类型决定执行方式
	if t.RemoteType == RemoteTypeSSH && sshCmd != "" {
		// SSH模式：需要特殊处理
		if wrapper != "" {
			// 密码认证模式：使用sshpass包装器
			fullCmd := fmt.Sprintf("%s rsync %s", wrapper, strings.Join(args, " "))
			global.LOG.Info("[rsyncmgr] fullCmd: %s", fullCmd)
			cmd = exec.CommandContext(ctx, "bash", "-c", fullCmd)
		} else {
			// 私钥认证模式：直接执行rsync命令
			quoted := make([]string, 0, len(args)+1)
			quoted = append(quoted, "rsync")
			for _, a := range args {
				// 使用 %q 进行安全转义，可准确打印空格、引号等
				quoted = append(quoted, fmt.Sprintf("%q", a))
			}
			global.LOG.Info("[rsyncmgr] fullCmd: %s", strings.Join(quoted, " "))

			cmd = exec.CommandContext(ctx, "rsync", args...)
		}
	} else {
		// Rsync守护进程模式：直接执行rsync命令
		cmd = exec.CommandContext(ctx, "rsync", args...)
	}

	// 设置环境变量，确保PATH和SHELL可用
	cmd.Env = append(os.Environ(), "SHELL=/bin/bash")

	// 使用日志处理器直接写入文件
	logWriter, err := logHandler.GetExecutionLogWriter()
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to get log writer for task %s: %v", t.ID, err)
		return nil, err
	}
	// 同时重定向标准输出和标准错误到同一个日志文件
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	// 设置进程组属性，便于进程管理
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		cancel()
		global.LOG.Error("[rsyncmgr] failed to start rsync for task %s: %v", t.ID, err)
		return nil, err
	}

	// 记录进程启动信息
	global.LOG.Error("[rsyncmgr] rsync process started for task %s, PID: %d", t.ID, cmd.Process.Pid)

	// 输出捕获器已经通过cmd.Stdout和cmd.Stderr关联，runTask可以直接访问
	proc := &ExecProcess{cmd: cmd, cancel: cancel}
	return proc, nil
}

// TestRsync 执行测试同步（dry-run），支持直接写入日志文件
func TestRsync(t *RsyncTask, logHandler *LogHandler) (string, error) {
	args, sshCmd, wrapper, err := buildRsyncCommand(t)
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to build rsync command for test task %s: %v", t.ID, err)
		return "", err
	}

	// 添加dry-run参数
	testArgs := append([]string{"--dry-run"}, args...)

	global.LOG.Error("[rsyncmgr] starting test rsync for task %s: %s", t.ID, strings.Join(testArgs, " "))

	var cmd *exec.Cmd
	ctx := context.Background()

	// 根据远程类型决定执行方式
	if t.RemoteType == RemoteTypeSSH && sshCmd != "" {
		if wrapper != "" {
			// 密码认证模式：使用sshpass包装器
			fullCmd := fmt.Sprintf("%s rsync %s", wrapper, strings.Join(testArgs, " "))
			global.LOG.Error("[rsyncmgr] test fullCmd: %s", fullCmd)
			cmd = exec.CommandContext(ctx, "bash", "-c", fullCmd)
		} else {
			// 私钥认证模式：直接执行rsync命令
			cmd = exec.CommandContext(ctx, "rsync", testArgs...)
		}
	} else {
		// Rsync守护进程模式：直接执行rsync命令
		cmd = exec.CommandContext(ctx, "rsync", testArgs...)
	}

	// 设置环境变量，确保PATH和SHELL可用
	cmd.Env = append(os.Environ(), "SHELL=/bin/bash")

	// 使用日志处理器直接写入文件
	logWriter, err := logHandler.GetTestLogWriter()
	if err != nil {
		global.LOG.Error("[rsyncmgr] failed to get test log writer for task %s: %v", t.ID, err)
		return "", err
	}
	// 同时重定向标准输出和标准错误到同一个日志文件
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	logPath := logHandler.GetTestLogPath()

	// 执行命令
	if err := cmd.Run(); err != nil {
		global.LOG.Error("[rsyncmgr] test rsync failed for task %s: %v", t.ID, err)
		return logPath, err
	}

	global.LOG.Error("[rsyncmgr] test rsync completed for task %s", t.ID)
	return logPath, nil
}
