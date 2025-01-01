package session

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/message"
)

type Session struct {
	End     chan struct{}  // 用于结束跟踪输出的信号
	ID      string         // 会话ID（比如 screen 会话名）
	Cmd     *exec.Cmd      // 执行的命令
	Stdin   io.WriteCloser // stdin
	Stdout  io.ReadCloser  // stdout
	Scanner *bufio.Scanner // 用于读取输出
}

type SessionService struct {
	SessionMap map[string]*Session
}

type ISessionServie interface {
	Start(sessionData message.SessionData, outputCallback func(string)) (*Session, error)
	Finish(sessionID string) error
	Detach(sessionID string) error
	Attach(sessionID string, outputCallback func(string)) (*Session, error)
	Rename(sessionID string, newSessionID string) error
	Input(sessionData message.SessionData) error
}

func NewISessionService() ISessionServie {
	return &SessionService{}
}

func (s *SessionService) Start(sessionData message.SessionData, outputCallback func(string)) (*Session, error) {
	if s.SessionMap[sessionData.SessionID] != nil {
		return nil, fmt.Errorf("session already exist")
	}

	// 启动 screen 会话
	screenCmd := exec.Command("screen", "-S", sessionData.SessionID, "-d", "-m", sessionData.Data)
	stdout, err := screenCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stdin, err := screenCmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	// 启动命令
	if err := screenCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start screen command: %v", err)
	}

	// 创建会话
	session := &Session{
		End:     make(chan struct{}),
		ID:      sessionData.SessionID,
		Cmd:     screenCmd,
		Stdin:   stdin,
		Stdout:  stdout,
		Scanner: bufio.NewScanner(stdout),
	}
	s.SessionMap[sessionData.SessionID] = session

	// 启动 goroutine 以跟踪输出
	go s.trackOutput(session, outputCallback)

	return session, nil
}

// 跟踪输出
func (s *SessionService) trackOutput(session *Session, outputCallback func(string)) {
	for {
		select {
		case <-session.End: // 检查是否收到结束信号
			global.LOG.Info("Stopping output tracking for session %s", session.ID)
			return
		default:
			if session.Scanner.Scan() {
				output := session.Scanner.Text()
				// 调用回调函数
				outputCallback(output)
			} else {
				// 检查是否发生错误
				if err := session.Scanner.Err(); err != nil {
					global.LOG.Error("Error reading from session output: %v", err)
					return // 发生错误时退出
				}
				// 如果没有更多输出，继续循环以等待新的输出
			}
		}
	}
}

func (s *SessionService) Finish(sessionID string) error {
	// 关闭会话
	s.closeSession(sessionID)
	// 从会话映射中删除会话
	delete(s.SessionMap, sessionID)

	// 通过命令行关闭 Screen 会话
	if err := exec.Command("screen", "-S", sessionID, "-X", "quit").Run(); err != nil {
		global.LOG.Error("Error stopping screen session %s: %v", sessionID, err)
		return err
	}
	global.LOG.Info("Session %s finished", sessionID)

	return nil
}

func (s *SessionService) closeSession(sessionID string) {
	if session, exists := s.SessionMap[sessionID]; exists {
		close(session.End) // 关闭 End 通道以停止跟踪

		// 安全关闭 Stdin 和 Stdout
		if err := session.Stdin.Close(); err != nil {
			global.LOG.Error("Error closing stdin for session %s: %v", sessionID, err)
		}
		if err := session.Stdout.Close(); err != nil {
			global.LOG.Error("Error closing stdout for session %s: %v", sessionID, err)
		}

	}
}

func (s *SessionService) Detach(sessionID string) error {
	// 关闭会话
	s.closeSession(sessionID)

	// 通过命令行分离 Screen 会话
	if err := exec.Command("screen", "-S", sessionID, "-X", "detach").Run(); err != nil {
		global.LOG.Error("Error detaching screen session %s: %v", sessionID, err)
		return err
	}
	global.LOG.Info("Session %s detached", sessionID)

	return nil
}

func (s *SessionService) Attach(sessionID string, outputCallback func(string)) (*Session, error) {
	if session, exists := s.SessionMap[sessionID]; exists {
		// 重新启动跟踪输出
		session.End = make(chan struct{}) // 重新创建 End 通道
		s.SessionMap[sessionID] = session // 重新添加到会话映射中

		// 启动命令以恢复 Screen 会话
		screenCmd := exec.Command("screen", "-r", sessionID)
		stdout, err := screenCmd.StdoutPipe()
		if err != nil {
			global.LOG.Error("failed to get stdout pipe: %v", err)
			return nil, err
		}

		stdin, err := screenCmd.StdinPipe()
		if err != nil {
			global.LOG.Error("failed to get stdin pipe: %v", err)
			return nil, err
		}

		// 启动命令
		if err := screenCmd.Start(); err != nil {
			global.LOG.Error("failed to start screen command: %v", err)
			return nil, err
		}

		// 更新会话的 Stdin, Stdout 和 Scanner
		session.Stdin = stdin
		session.Stdout = stdout
		session.Scanner = bufio.NewScanner(stdout)

		// 启动 goroutine 以跟踪输出
		go s.trackOutput(session, outputCallback)

		global.LOG.Info("Session %s attached", sessionID)
		return session, nil
	}
	return nil, fmt.Errorf("session %s not found", sessionID)
}

func (s *SessionService) Rename(sessionID string, newSessionID string) error {
	if session, exists := s.SessionMap[sessionID]; exists {
		// 从会话映射中删除会话
		delete(s.SessionMap, sessionID)

		session.ID = newSessionID
		s.SessionMap[newSessionID] = session
	}
	// 执行 screen 重命名操作
	if err := exec.Command("screen", "-S", sessionID, "-X", "sessionname", newSessionID).Run(); err != nil {
		global.LOG.Error("Error renaming screen session %s to %s: %v", sessionID, newSessionID, err)
		return err
	}
	return fmt.Errorf("session %s not found", sessionID)
}

func (s *SessionService) Input(sessionData message.SessionData) error {
	if session, exists := s.SessionMap[sessionData.SessionID]; exists {
		_, err := session.Stdin.Write([]byte(sessionData.Data + "\n")) // 向 Stdin 写入输入内容
		return err
	}
	return fmt.Errorf("session %s not found", sessionData.SessionID)
}
