package session

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/creack/pty"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type Session struct {
	ID        string
	Name      string
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

	// 是否传了名字
	var sessionName string
	if sessionData.Data != "" {
		sessionName = sessionData.Data
	} else {
		// 枚举会话，并创建一个 idb-n
		name, err := s.genSessionName()
		if err != nil {
			// 达到限制
			if err.Error() == constant.ErrSessionLimit {
				return nil, fmt.Errorf("%s", "iDB terminal session limit reached. Please clear inactive sessions.")
			}
			global.LOG.Error("failed to gen session name: %v", err)
			return nil, fmt.Errorf("failed to gen session name: %v", err)
		}
		sessionName = name
	}

	screenCmd := exec.Command("screen", "-S", sessionName, "-s", "bash")

	// 设置环境变量
	// homeDir := os.Getenv("HOME")
	// path := os.Getenv("PATH")
	// global.LOG.Info("HOME: %s \n PATH: %s", homeDir, path)
	screenCmd.Env = append(os.Environ(),
		"TERM=screen-256color", // 设置为xterm以兼容xterm.js
		"SHELL=/bin/bash",      // 设置默认shell
		// "HOME=/root",           // 设置用户主目录
		// "PATH="+path,           // 确保PATH包含必要的命令
	)

	// 创建伪终端
	rows := 24
	if sessionData.Rows > 0 {
		rows = sessionData.Rows
	}
	cols := 80
	if sessionData.Cols > 0 {
		cols = sessionData.Cols
	}
	ws := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}
	ptyFile, err := pty.StartWithSize(screenCmd, ws)
	if err != nil {
		global.LOG.Error("failed to start pty: %v", err)
		return nil, fmt.Errorf("failed to start pty: %v", err)
	}

	global.LOG.Info("session started")

	// 延迟一点点
	time.Sleep(100 * time.Millisecond)

	// 找到会话
	sessionID, err := s.getSessionID(sessionName)
	if err != nil {
		global.LOG.Error("failed to get session ID: %v", err)
		return nil, fmt.Errorf("failed to find session: %v", err)
	}
	session := &Session{
		ID:        sessionID,
		Name:      sessionName,
		Cmd:       screenCmd,
		Pty:       ptyFile,
		Conn:      s.centerConn,
		SecretKey: s.secretKey,
	}

	global.LOG.Info("Session %s started", session.ID)
	return session, nil
}

func (s *SessionService) Attach(sessionData message.SessionData) (*Session, error) {
	var (
		sessionID   string
		sessionName string
	)

	// 获取当前存在的 Detached 会话
	sessions, _ := s.page(true)
	global.LOG.Info("found sessions: \n %v", sessions)
	// 不存在任何会话
	if len(sessions) == 0 {
		// 创建新会话
		global.LOG.Info("No exist sessions, create one")
		return s.Start(sessionData)
	}

	// 请求没有传入session时，看是否需要创建会话
	if sessionData.Session == "" {
		// 查找时间最近，且已经Detached的会话
		latestSession := sessions[0]
		for _, session := range sessions {
			if session.Time.After(latestSession.Time) && session.Status == "Detached" {
				latestSession = session
			}
		}
		if latestSession.Status != "Detached" {
			global.LOG.Info("No detached sessions, create one")
			return s.Start(sessionData)
		} else {
			// sessionID
			sessionID = latestSession.Session
			sessionName = latestSession.Name
		}
	} else {
		// 请求传入了session，找到该会话
		var toAttach *model.SessionInfo
		for _, session := range sessions {
			if session.Session == sessionData.Session {
				toAttach = &session
			}
		}
		// 没找到
		if toAttach == nil {
			global.LOG.Error("session %s not found", sessionData.Session)
			return nil, fmt.Errorf("session %s not found", sessionData.Session)
		}
		// 已经 attached
		// if toAttach.Status == "Attached" {
		// 	global.LOG.Error("session %s already attached", toAttach.Session)
		// 	return nil, fmt.Errorf("session %s already attached", toAttach.Name)
		// }

		sessionID = toAttach.Session
		sessionName = toAttach.Name
	}
	global.LOG.Info("Find session: %s.%s", sessionID, sessionName)

	// 恢复会话
	screenCmd := exec.Command("screen", "-d", "-r", sessionID)
	global.LOG.Info("Attaching to session: %s", sessionID)

	// 设置环境变量
	homeDir, _ := os.UserHomeDir()
	global.LOG.Info("homedir: %s", homeDir)
	screenCmd.Env = append(os.Environ(),
		"TERM=xterm",              // 设置为xterm以兼容xterm.js
		"SHELL=/bin/bash",         // 设置默认shell
		"HOME="+homeDir,           // 设置用户主目录
		"PATH="+os.Getenv("PATH"), // 确保PATH包含必要的命令
	)

	// 创建伪终端
	rows := 24
	if sessionData.Rows > 0 {
		rows = sessionData.Rows
	}
	cols := 80
	if sessionData.Cols > 0 {
		cols = sessionData.Cols
	}
	ws := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}
	ptyFile, err := pty.StartWithSize(screenCmd, ws)
	if err != nil {
		global.LOG.Error("failed to start pty: %v", err)
		return nil, fmt.Errorf("failed to start pty: %v", err)
	}

	global.LOG.Info("session started")

	// 创建会话
	session := &Session{
		ID:        sessionID,
		Name:      sessionName,
		Cmd:       screenCmd,
		Pty:       ptyFile,
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

func (s *Session) Input(data string) error {
	_, err := s.Pty.Write([]byte(data))
	if err != nil {
		global.LOG.Error("Error writing to PTY for session %s: %v", s.ID, err)
		return err
	}
	global.LOG.Info("Input sent to session %s", s.ID)
	return nil
}

func (s *Session) Resize(cols, rows int) error {
	ws := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}
	err := pty.Setsize(s.Pty, ws)
	if err != nil {
		global.LOG.Error("Error resize session %s: %v", s.ID, err)
		return err
	}
	global.LOG.Info("Resize session %s", s.ID)
	return nil
}

// 跟踪输出
func (s *Session) trackOutput(quitChan chan bool) {
	defer common.SetQuit(quitChan)

	// reader := bufio.NewReader(s.Pty)
	buf := make([]byte, 1024)
	tick := time.NewTicker(time.Millisecond * 60)
	defer tick.Stop()
	for {
		select {
		case <-quitChan:

			return
		case <-tick.C:
			n, err := s.Pty.Read(buf)
			if err != nil {
				if err == io.EOF {
					global.LOG.Info("PTY closed")
					return
				}
				global.LOG.Error("failed to read from PTY: %v", err)
				return
			}

			if n > 0 {
				// 输出数据
				global.LOG.Info("output: %s", string(buf[:n]))
				go s.sendSessionResult(string(buf[:n]))
			}
		}
	}

}

func (s *Session) sendSessionResult(data string) {
	rspMsg, err := message.CreateSessionMessage(
		utils.GenerateMsgId(),
		message.WsMessageCmd,
		message.SessionData{Code: constant.CodeSuccess, Msg: "", Session: s.ID, Data: data},
		s.SecretKey,
		utils.GenerateNonce(16),
		global.Version,
	)
	if err != nil {
		global.LOG.Error("[Session] Error creating session rsp message: %v", err)
		return
	}

	err = message.SendSessionMessage(*s.Conn, rspMsg)
	if err != nil {
		global.LOG.Error("[Session] Failed to send session rsp : %v", err)
		return
	}
	global.LOG.Info("[Session] Session rsp sent")
}

func (s *SessionService) Page() (*model.PageResult, error) {
	var result model.PageResult

	sessions, _ := s.page(false)

	result.Total = int64(len(sessions))
	result.Items = sessions

	return &result, nil
}

func (s *SessionService) page(filterDetached bool) ([]model.SessionInfo, error) {
	var sessions []model.SessionInfo

	// 执行命令以列出所有的 screen 会话
	cmd := exec.Command("screen", "-ls")
	output, err := cmd.Output()
	if strings.Contains(string(output), "No Sockets found") {
		global.LOG.Info("no session found")
		return sessions, nil
	}
	if err != nil {
		global.LOG.Error("failed to list sessions: %v", err)
		return sessions, nil
	}

	// 处理返回的结果字符串
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

			// matches[0]是完整匹配的字符串
			// matches[1]是第一个捕获组,即(\d+\.[^\s]+)匹配的内容
			sessionParts := strings.Split(matches[1], ".")
			if len(sessionParts) != 2 {
				continue
			}
			sessionID := sessionParts[0]
			sessionName := sessionParts[1]
			timeStr := matches[2]
			status := matches[3]

			// 如果只筛选 Detached 会话
			if filterDetached && status != "Detached" {
				continue
			}

			parsedTime, err := time.Parse("01/02/2006 03:04:05 PM", timeStr)
			if err != nil {
				global.LOG.Error("Error parsing time: %v", err)
				continue
			}

			sessionInfo := model.SessionInfo{
				Session: sessionID,
				Name:    sessionName,
				Time:    parsedTime,
				Status:  status,
			}
			sessions = append(sessions, sessionInfo)
		}
	}

	return sessions, nil
}

func (s *SessionService) genSessionName() (string, error) {
	var sessionName string

	// 执行命令以列出所有的 screen 会话
	cmd := exec.Command("screen", "-ls")
	output, err := cmd.Output()
	if strings.Contains(string(output), "No Sockets found") {
		global.LOG.Info("no session found")
		sessionName = fmt.Sprintf("idb-%d", 1)
		global.LOG.Info("generated session name: %s", sessionName)
		return sessionName, nil
	}
	if err != nil {
		global.LOG.Error("failed to list sessions: %v", err)
		return "", err
	}

	// 处理返回的结果字符串
	var numbers []int
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 跳过空行或不包含会话信息的行
		if !strings.Contains(line, ".idb-") {
			continue
		}

		// 使用正则表达式提取会话编号,使用\s+匹配空格
		re := regexp.MustCompile(`idb-(\d+)\s+`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				numbers = append(numbers, num)
			}
		}
	}

	// 如果已经有20个了，不允许再创建新的
	if len(numbers) >= 20 {
		return "", errors.New(constant.ErrSessionLimit)
	}

	// 得到了当前的所有numbers，按从低到高排序
	sort.Ints(numbers)

	// 找到第一个缺失的数字，如果没有缺失就用最大值+1
	nextNum := 1
	for _, num := range numbers {
		if num != nextNum {
			break
		}
		nextNum++
	}

	// 生成新的会话名称
	sessionName = fmt.Sprintf("idb-%d", nextNum)
	global.LOG.Info("generated session name: %s", sessionName)

	return sessionName, nil
}

func (s *SessionService) getSessionID(sessionName string) (string, error) {
	// 执行 screen -ls 命令获取所有会话列表
	output, err := exec.Command("screen", "-ls").Output()
	if strings.Contains(string(output), "No Sockets found") {
		global.LOG.Info("no session found")
		return "", fmt.Errorf("no session found")
	}
	if err != nil {
		global.LOG.Error("failed to list sessions: %v", err)
		return "", fmt.Errorf("failed to list sessions: %v", err)
	}

	// 处理返回的结果字符串
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// 查找包含 .sessionName 的行
		if !strings.Contains(line, "."+sessionName) {
			continue
		}

		// 使用正则表达式提取会话ID
		re := regexp.MustCompile(fmt.Sprintf(`(\d+)\.%s\s+`, sessionName))
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("session %s not found", sessionName)
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
