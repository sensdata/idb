package terminal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

// ScreenSession
type ScreenSession struct {
	*BaseSession
}

// New instance
func NewScreenSession(session string, name string, cols, rows int) Session {
	return &ScreenSession{
		BaseSession: NewBaseSession(session, name, cols, rows),
	}
}

// Start session
func (s *ScreenSession) Start() error {
	// check screen command
	if _, err := exec.LookPath("screen"); err != nil {
		return errors.New(constant.ErrNotInstalled)
	}

	// 是否传了名字
	var sessionName string
	if s.Name != "" {
		sessionName = s.Name
	} else {
		// 枚举会话，并创建一个 idb-n
		name, err := s.genSessionName()
		if err != nil {
			// 达到限制
			if err.Error() == constant.ErrSessionLimit {
				return fmt.Errorf("%s", "iDB terminal session limit reached. Please clear inactive sessions.")
			}
			return fmt.Errorf("failed to gen session name: %v", err)
		}
		sessionName = name
		global.LOG.Info("gen session name: %s", sessionName)
	}
	s.Name = sessionName

	// create cmd
	tmpRcPath := filepath.Join(os.TempDir(), "idb.screenrc")
	s.cmd = exec.Command("screen", "-c", tmpRcPath, "-S", s.Name, "-s", "bash")
	s.cmd.Env = append(os.Environ(),
		"TERM=screen-256color",
		"SHELL=/bin/bash",
	)

	// win size
	rows := 24
	if s.rows > 0 {
		rows = s.rows
	}
	cols := 80
	if s.cols > 0 {
		cols = s.cols
	}
	ws := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}

	// create pty
	ptyFile, err := pty.StartWithSize(s.cmd, ws)
	if err != nil {
		return fmt.Errorf("failed to start pty: %v", err)
	}
	s.pty = ptyFile

	// find session id
	var sessionID string
	found := false
	for i := 0; i < 5; i++ {
		// delay
		time.Sleep(200 * time.Millisecond)
		id, err := s.getSessionID(s.Name)
		if err != nil {
			global.LOG.Error("failed to get session %s id: %v", s.Name, err)
			continue
		}
		sessionID = id
		found = true
		global.LOG.Info("Found session ID %s for session name %s", id, s.Name)
	}
	if !found {
		return fmt.Errorf("SessionID not found")
	}
	s.Session = sessionID

	// tracking and wait
	go s.trackOutput()
	go s.wait()

	return nil
}

// Attach session
func (s *ScreenSession) Attach() error {
	// check screen command
	if _, err := exec.LookPath("screen"); err != nil {
		return errors.New(constant.ErrNotInstalled)
	}

	// get all detached sessions
	sessions, _ := listScreenSessions(true)
	global.LOG.Info("found %d detached sessions", len(sessions))
	// no detached session
	if len(sessions) == 0 {
		// create one
		global.LOG.Error("No sessions")
		return fmt.Errorf("no session to attach")
	}

	// ID isn't specified
	if s.Session == "" {
		// find latest detached session
		latestSession := sessions[0]
		for _, session := range sessions {
			if session.Time.After(latestSession.Time) && session.Status == "Detached" {
				latestSession = session
			}
		}
		if latestSession.Status != "Detached" {
			global.LOG.Error("No detached sessions")
			return fmt.Errorf("no detached session to attach")
		} else {
			s.Session = latestSession.Session
			s.Name = latestSession.Name
		}
	} else {
		// 请求传入了session，找到该会话
		var toAttach *model.SessionInfo
		for _, session := range sessions {
			if session.Session == s.Session {
				toAttach = &session
			}
		}
		// 没找到
		if toAttach == nil {
			global.LOG.Error("session %s not found", s.Session)
			return fmt.Errorf("session %s not found", s.Session)
		}

		s.Session = toAttach.Session
		s.Name = toAttach.Name
	}
	global.LOG.Info("Find session: %s.%s", s.Session, s.Name)

	// create cmd
	s.cmd = exec.Command("screen", "-d", "-r", s.Session)
	s.cmd.Env = append(os.Environ(),
		"TERM=xterm",
		"SHELL=/bin/bash",
	)

	// win size
	rows := 24
	if s.rows > 0 {
		rows = s.rows
	}
	cols := 80
	if s.cols > 0 {
		cols = s.cols
	}
	ws := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}

	// create pty
	ptyFile, err := pty.StartWithSize(s.cmd, ws)
	if err != nil {
		return fmt.Errorf("failed to start pty: %v", err)
	}
	s.pty = ptyFile

	global.LOG.Info("pty started")

	// tracking and wait
	go s.trackOutput()
	go s.wait()

	return nil
}

func (s *ScreenSession) genSessionName() (string, error) {
	var sessionName string

	// 执行命令以列出所有的 screen 会话
	cmd := exec.Command("bash", "-c", "LC_TIME=en_US.UTF-8 screen -ls")
	output, err := cmd.CombinedOutput()
	if strings.Contains(string(output), "No Sockets found") {
		sessionName = fmt.Sprintf("idb-%d", 1)
		return sessionName, nil
	}
	// 注意: 某些版本的screen -ls在有会话时会返回退出码1，这是正常行为
	// 记录警告日志但继续尝试处理输出
	if err != nil {
		global.LOG.Warn("screen -ls returned non-zero exit code (common behavior in some versions): %v, output: %s", err, string(output))
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
	return sessionName, nil
}

func (s *ScreenSession) getSessionID(sessionName string) (string, error) {
	// 执行 screen -ls 命令获取所有会话列表
	cmd := exec.Command("bash", "-c", "LC_TIME=en_US.UTF-8 screen -ls")
	output, err := cmd.CombinedOutput()
	if strings.Contains(string(output), "No Sockets found") {
		return "", fmt.Errorf("no session found")
	}
	// 注意: 某些版本的screen -ls在有会话时会返回退出码1，这是正常行为
	// 记录警告日志但继续尝试处理输出
	if err != nil {
		global.LOG.Warn("screen -ls returned non-zero exit code (common behavior in some versions): %v, output: %s", err, string(output))
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

	return "", fmt.Errorf("session %s id not found", sessionName)
}
