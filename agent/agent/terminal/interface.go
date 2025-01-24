package terminal

import (
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
)

// Session
type Session interface {
	// Get session type
	Type() message.SessionType
	// Release
	Release()
	// Start session
	Start() error
	// Attach session
	Attach() error
	// Write to session
	Input(data string) error
	// Resize
	Resize(cols int, rows int) error
}

// Manager
type Manager interface {
	// Store session
	StoreSession(session Session)
	// Remove session
	RemoveSession(session string)
	// Start session
	StartSession(sessionType message.SessionType, name string, cols, rows int, quitChan chan bool, outputChan chan string) (Session, error)
	// Attach session
	AttachSession(sessionType message.SessionType, session string, cols, rows int, quitChan chan bool, outputChan chan string) (Session, error)
	// List sessions
	ListSessions(sessionType message.SessionType) (*model.PageResult, error)
	// Detach session
	DetachSession(sessionType message.SessionType, session string) error
	// Quit session
	QuitSession(sessionType message.SessionType, session string) error
	// Write to session
	InputSession(sessionType message.SessionType, session string, data string) error
	// ResizeSession
	ResizeSession(sessionType message.SessionType, session string, cols int, rows int) error
	// RenameSession
	RenameSession(sessionType message.SessionType, session string, data string) error
}
