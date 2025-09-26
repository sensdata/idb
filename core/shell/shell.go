package shell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sensdata/idb/core/model"
)

func ExecuteCommandIgnore(command string) error {
	output, err := ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("error : %v, output: %s", err, output)
	}
	return nil
}

func ExecuteCommand(command string) (string, error) {
	return executeCommand("/bin/bash", []string{"-c", command})
}

func ExecuteCommands(commands []string) (results []string, err error) {
	for _, command := range commands {
		result, err := ExecuteCommand(command)
		if err != nil {
			results = append(results, "error")
		} else {
			results = append(results, result)
		}
	}
	return results, nil
}

func executeCommand(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	return runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func ExecuteScript(req model.ScriptExec) *model.ScriptResult {
	//执行脚本
	result := executeScript(req)

	// 记录日志
	scriptName := filepath.Base(req.ScriptPath)
	scriptLog(req.LogPath, scriptName, result)

	// 移除脚本
	if req.Remove {
		os.Remove(req.ScriptPath)
	}

	return result
}

func executeScript(req model.ScriptExec) *model.ScriptResult {
	result := model.ScriptResult{
		LogPath: req.LogPath,
		Start:   time.Now(),
		End:     time.Now(),
		Out:     "",
		Err:     "",
	}

	// 检查脚本文件是否存在及权限
	fileInfo, err := os.Stat(req.ScriptPath)
	if os.IsNotExist(err) {
		result.Err = "script not found"
		return &result
	}
	if err != nil {
		result.Err = fmt.Sprintf("failed to access script: %v", err)
		return &result
	}
	if fileInfo.Mode().Perm()&(1<<(uint(7))) == 0 {
		result.Err = "script file is not executable"
		return &result
	}

	// 设置执行上下文
	var ctx context.Context
	var cancel context.CancelFunc
	if req.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background()) // 不设限
	}
	defer cancel()

	// 定义命令和缓冲区
	cmd := exec.CommandContext(ctx, "/bin/bash", req.ScriptPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// 记录开始时间
	result.Start = time.Now()

	// 启动脚本并等待完成
	if err := cmd.Start(); err != nil {
		result.End = time.Now()
		result.Err = fmt.Sprintf("failed to start script: %v", err)
		return &result
	}

	err = cmd.Wait()
	result.End = time.Now()
	result.Out = out.String()

	// 检查超时和其他错误情况
	if ctx.Err() == context.DeadlineExceeded {
		result.Err = fmt.Sprintf("script execution timed out after %ds", req.Timeout)
	} else if err != nil {
		result.Err = fmt.Sprintf("script execution failed: %v", err)
	}

	return &result
}

func scriptLog(logPath string, tag string, result *model.ScriptResult) error {
	// 确保日志目录存在
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 创建日志文件
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := logFile.Close(); closeErr != nil {
			fmt.Printf("failed to close log file: %v\n", closeErr)
		}
	}()

	// 记录开始
	if _, err := fmt.Fprintf(logFile, "[%s]EXECUTION START [%s]\n", tag, result.Start.Format("2006-01-02 15:04:05")); err != nil {
		return err
	}

	// 记录输出
	if result.Out != "" {
		if _, err := fmt.Fprintf(logFile, "\n[%s]EXECUTION OUTPUT:\n%s\n", tag, result.Out); err != nil {
			return err
		}
	}

	// 记录错误
	if result.Err != "" {
		if _, err := fmt.Fprintf(logFile, "\n[%s]EXECUTION FAILED: %v\n", tag, result.Err); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(logFile, "\n[%s]EXECUTION SUCCESS\n", tag); err != nil {
			return err
		}
	}

	// 记录结束
	if _, err := fmt.Fprintf(logFile, "[%s]EXECUTION END [%s]\n\n", tag, result.End.Format("2006-01-02 15:04:05")); err != nil {
		return err
	}

	return nil
}
