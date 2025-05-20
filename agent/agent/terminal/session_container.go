package terminal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/sensdata/idb/core/constant"
)

type ContainerSession struct {
	*BaseSession
	Shell string
}

func NewContainerSession(session string, name string, shell string, cols, rows int) Session {
	return &ContainerSession{
		BaseSession: NewBaseSession(session, name, cols, rows),
		Shell:       shell,
	}
}

func (s *ContainerSession) Start() error {
	// 检查 docker 命令
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New(constant.ErrNotInstalled)
	}

	// 是否传了id
	if s.Session == "" {
		return errors.New("container id is empty")
	}

	// create cmd
	s.cmd = exec.Command("docker", "exec", "-it", s.Session, s.Shell)
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

	// delay
	time.Sleep(200 * time.Millisecond)

	// tracking and wait
	go s.trackOutput()
	go s.wait()

	return nil
}
