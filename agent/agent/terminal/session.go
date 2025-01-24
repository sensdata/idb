package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils/common"
)

// BaseSession
type BaseSession struct {
	// Session Session
	Session string
	// Session Name
	Name string
	// Session Type
	sessionType message.SessionType
	// Status
	Status string
	// Time
	CreateAt time.Time
	// cols
	cols int
	// rows
	rows int
	// PTY
	pty *os.File
	// Command
	cmd *exec.Cmd
	// Mutex
	mutex sync.Mutex
	// Chan for quiting
	quitChan chan bool
	// Chan for output
	outputChan chan string
}

// New instance
func NewBaseSession(session string, name string, cols, rows int, quitChan chan bool, outputChan chan string) *BaseSession {
	return &BaseSession{
		Session:     session,
		Name:        name,
		sessionType: message.SessionTypeBash,
		cols:        cols,
		rows:        rows,
		quitChan:    quitChan,
		outputChan:  outputChan,
	}
}

func (s *BaseSession) Type() message.SessionType {
	return s.sessionType
}

// Start Session
func (s *BaseSession) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cmd != nil {
		return fmt.Errorf("session already started")
	}

	// create cmd
	s.cmd = exec.Command("bash")
	s.cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
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

	// tracking and wait
	go s.trackOutput()
	go s.wait()

	return nil
}

// Attach session
func (s *BaseSession) Attach() error {
	return nil
}

// Write to session
func (s *BaseSession) Input(data string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.pty == nil {
		return fmt.Errorf("session not started")
	}

	_, err := s.pty.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("session write failed: %v", err)
	}
	return nil
}

// Resize
func (s *BaseSession) Resize(cols int, rows int) error {
	ws := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}
	err := pty.Setsize(s.pty, ws)
	if err != nil {
		return err
	}
	return nil
}

// release
func (s *BaseSession) Release() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 关闭PTY
	if s.pty != nil {
		s.pty.Close()
	}

	// 停止命令
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
}

// Track output
func (s *BaseSession) trackOutput() {
	defer common.SetQuit(s.quitChan)

	buf := make([]byte, 1024)
	tick := time.NewTicker(time.Millisecond * 60)
	defer tick.Stop()
	for {
		select {
		case <-s.quitChan:
			return
		case <-tick.C:
			n, err := s.pty.Read(buf)
			if err != nil {
				if err == io.EOF {
					global.LOG.Info("PTY closed")
					return
				}
				global.LOG.Error("failed to read from PTY: %v", err)
				return
			}

			if n > 0 {
				global.LOG.Info("output: %s", string(buf[:n]))
				s.outputChan <- string(buf[:n])
			}
		}
	}
}

// Wait
func (s *BaseSession) wait() {
	if err := s.cmd.Wait(); err != nil {
		common.SetQuit(s.quitChan)
	}
}
