package session

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type Session struct {
	ID        string // 会话ID（比如 screen 会话名）
	Cmd       *exec.Cmd
	Pty       *os.File
	Conn      *net.Conn
	SecretKey string
}

type SessionService struct {
	centerConn *net.Conn
	secretKey  string
}

type ISessionServie interface {
	Config(conn *net.Conn, secretKey string)

	Start(sessionData message.SessionData) (*Session, error)
	Attach(sessionData message.SessionData) (*Session, error)
	Input(sessionData message.SessionData) error

	Page() (*model.PageResult, error)
	Finish(sessionID string) error
	Detach(sessionID string) error
	Rename(sessionID string, newSessionID string) error
}

func NewISessionService() ISessionServie {
	return &SessionService{}
}

func (s *SessionService) Config(conn *net.Conn, secretKey string) {
	s.centerConn = conn
	s.secretKey = secretKey
}

func (s *SessionService) Start(sessionData message.SessionData) (*Session, error) {
	// 启动 screen 会话
	global.LOG.Info("starting")

	screenCmd := exec.Command("screen", "-S", sessionData.SessionID)

	// 设置环境变量 TERM
	screenCmd.Env = append(os.Environ(), "TERM=xterm")

	// 创建伪终端
	pty, err := pty.Start(screenCmd)
	if err != nil {
		global.LOG.Error("failed to start pty: %v", err)
		return nil, fmt.Errorf("failed to start pty: %v", err)
	}

	global.LOG.Info("session started")

	// 创建会话
	session := &Session{
		ID:        sessionData.SessionID,
		Cmd:       screenCmd,
		Pty:       pty,
		Conn:      s.centerConn,
		SecretKey: s.secretKey,
	}

	global.LOG.Info("Session %s started", session.ID)
	return session, nil
}

func (s *SessionService) Attach(sessionData message.SessionData) (*Session, error) {
	// 启动命令以恢复 Screen 会话
	screenCmd := exec.Command("screen", "-r", sessionData.SessionID)

	// 设置环境变量 TERM
	screenCmd.Env = append(os.Environ(), "TERM=xterm")

	// 创建伪终端
	pty, err := pty.Start(screenCmd)
	if err != nil {
		global.LOG.Error("failed to start pty: %v", err)
		return nil, fmt.Errorf("failed to start pty: %v", err)
	}

	global.LOG.Info("session started")

	// 创建会话
	session := &Session{
		ID:        sessionData.SessionID,
		Cmd:       screenCmd,
		Pty:       pty,
		Conn:      s.centerConn,
		SecretKey: s.secretKey,
	}

	global.LOG.Info("Session %s attached", session.ID)
	return session, nil
}

func (s *Session) Start(quitChan chan bool) {
	go s.trackOutput(quitChan)
}

func (s *Session) Wait(quitChan chan bool) {
	if err := s.Cmd.Wait(); err != nil {
		quitChan <- true
	}
}

// 跟踪输出
func (s *Session) trackOutput(quitChan chan bool) {
	defer common.SetQuit(quitChan)

	tick := time.NewTicker(time.Millisecond * time.Duration(60))
	defer tick.Stop()
	for {
		select {
		case <-quitChan:

			return
		case <-tick.C:
			if s.Pty == nil {
				global.LOG.Error("no pty")
				return
			}
			// 读取 PTY 的输出
			bs := make([]byte, 1024)
			n, err := s.Pty.Read(bs)
			if err != nil && err.Error() != "EOF" {
				global.LOG.Error("failed to read from PTY: %v", err)
				return
			}

			if n > 0 {
				// 输出数据
				global.LOG.Info("output: %s", string(bs[:n]))
				go s.sendSessionResult(string(bs[:n]))
			}
		}
	}

}

func (s *Session) sendSessionResult(data string) {
	// 处理输出，可能是发送给其他组件
	rspMsg, err := message.CreateSessionMessage(
		utils.GenerateMsgId(),
		message.TerminalCommand,
		message.SessionData{SessionID: s.ID, Data: data},
		s.SecretKey,
		utils.GenerateNonce(16),
	)
	if err != nil {
		global.LOG.Error("Error creating session rsp message: %v", err)
		return
	}

	err = message.SendSessionMessage(*s.Conn, rspMsg)
	if err != nil {
		global.LOG.Error("Failed to send session rsp : %v", err)
		return
	}
	global.LOG.Info("Session rsp sent")
}

func (s *SessionService) Page() (*model.PageResult, error) {
	var result model.PageResult

	// 执行命令以列出所有的 screen 会话
	cmd := exec.Command("screen", "-ls")
	output, err := cmd.Output()
	if err != nil {
		global.LOG.Error("failed to list sessions: %v", err)
		return &result, fmt.Errorf("failed to list sessions: %v", err)
	}
	global.LOG.Info("output: %s", output)

	// 处理返回的结果字符串
	var sessions []model.SessionInfo
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 解析每一行以提取会话信息
		// 假设会话信息格式为 " 12345.session_name (01/02/2025 12:52:58 PM) (Attached)"
		if strings.Contains(line, ".") {
			// 使用正则表达式提取会话信息
			re := regexp.MustCompile(`(\d+\.[^\s]+)\s+\((\d{2}/\d{2}/\d{4}\s+\d{2}:\d{2}:\d{2}\s+[AP]M)\)\s+\((Attached|Detached)\)`)
			matches := re.FindStringSubmatch(line)

			if len(matches) != 4 {
				continue
			}

			sessionParts := strings.Split(matches[1], ".")
			if len(sessionParts) != 2 {
				continue
			}
			sessionID := sessionParts[1]
			timeStr := matches[2]
			status := matches[3]

			parsedTime, err := time.Parse("01/02/2006 03:04:05 PM", timeStr)
			if err != nil {
				global.LOG.Error("Error parsing time: %v", err)
				continue
			}

			sessionInfo := model.SessionInfo{
				Session: sessionID,
				Time:    parsedTime,
				Status:  status,
			}
			sessions = append(sessions, sessionInfo)
		}
	}

	result.Total = int64(len(sessions))
	result.Items = sessions

	return &result, nil
}

func (s *SessionService) Finish(sessionID string) error {
	// 通过命令行关闭 Screen 会话
	if err := exec.Command("screen", "-S", sessionID, "-X", "quit").Run(); err != nil {
		global.LOG.Error("Error stopping screen session %s: %v", sessionID, err)
		return err
	}
	global.LOG.Info("Session %s finished", sessionID)

	return nil
}

func (s *SessionService) Detach(sessionID string) error {
	// 通过命令行分离 Screen 会话
	if err := exec.Command("screen", "-S", sessionID, "-X", "detach").Run(); err != nil {
		global.LOG.Error("Error detaching screen session %s: %v", sessionID, err)
		return err
	}
	global.LOG.Info("Session %s detached", sessionID)

	return nil
}

func (s *SessionService) Rename(sessionID string, newSessionID string) error {
	// 执行 screen 重命名操作
	if err := exec.Command("screen", "-S", sessionID, "-X", "sessionname", newSessionID).Run(); err != nil {
		global.LOG.Error("Error renaming screen session %s to %s: %v", sessionID, newSessionID, err)
		return err
	}
	return nil
}

func (s *SessionService) Input(sessionData message.SessionData) error {
	// 执行 screen 输入
	if err := exec.Command("screen", "-S", sessionData.SessionID, "-X", "stuff", sessionData.Data).Run(); err != nil {
		global.LOG.Error("Error input screen session %s: %v", sessionData.SessionID, err)
		return err
	}
	return nil
}
