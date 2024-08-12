package session

import (
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

// 会话上下文结构体
type SessionContext struct {
	ID        string
	PTY       *os.File
	Command   *exec.Cmd
	Conn      net.Conn
	CreatedAt time.Time
	Active    bool
	mu        sync.Mutex
}

// 创建新的会话上下文
func NewSessionContext(id string, conn net.Conn, cmd *exec.Cmd, ptyFile *os.File) *SessionContext {
	return &SessionContext{
		ID:        id,
		Conn:      conn,
		Command:   cmd,
		PTY:       ptyFile,
		CreatedAt: time.Now(),
		Active:    true,
	}
}

// 更新会话状态
func (s *SessionContext) UpdateActiveStatus(status bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = status
}

// 终止会话
func (s *SessionContext) Terminate() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = false
	if s.PTY != nil {
		s.PTY.Close()
	}
	if s.Command != nil {
		return s.Command.Process.Kill()
	}
	return nil
}
