package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
)

// ScreenSession
type ScreenSession struct {
	*BaseSession
}

// New instance
func NewScreenSession(session string, name string, cols, rows int, quitChan chan bool, outputChan chan string) Session {
	return &ScreenSession{
		BaseSession: NewBaseSession(session, name, cols, rows, quitChan, outputChan),
	}
}

// Start session
func (s *ScreenSession) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// check screen command
	if _, err := exec.LookPath("screen"); err != nil {
		return fmt.Errorf("screen is not installed")
	}

	// create cmd
	s.cmd = exec.Command("screen", "-S", s.Name, "-s", "bash")
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

	// delay
	time.Sleep(200 * time.Millisecond)

	// find session id
	sessionID, err := s.findSessionId(s.Name)
	if err != nil {
		return fmt.Errorf("failed to find session: %v", err)
	}
	s.Session = sessionID

	// tracking and wait
	go s.trackOutput()
	go s.wait()

	return nil
}

// Attach session
func (s *ScreenSession) Attach() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// check screen command
	if _, err := exec.LookPath("screen"); err != nil {
		return fmt.Errorf("screen is not installed")
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

func (s *ScreenSession) findSessionId(name string) (string, error) {
	command := fmt.Sprintf("screen -ls | grep -oP '\\d+\\.\\S+' | grep '%s' | awk -F. '{print $1}'", name)
	output, err := exec.Command(command).Output()
	if err != nil {
		global.LOG.Error("failed to find session id: %v", err)
		return "", fmt.Errorf("failed to find session id: %v", err)
	}
	return string(output), nil
}
