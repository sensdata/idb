package terminal

import (
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
)

// Session
type Session interface {
	// Get session type
	GetType() message.SessionType
	// Get session id
	GetSession() string
	// Get session name
	GetName() string
	// Get outputChan
	GetOutputChan() <-chan []byte
	// Get doneChan
	GetDoneChan() <-chan struct{}
	// Release
	Release() error
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
	// Release All sessions
	ReleaseAllSessions() error
	// Store session
	StoreSession(session Session)
	// Remove session
	RemoveSession(id string)
	// Get session
	GetSession(id string) (Session, error)
	// Start session
	StartSession(sessionType message.SessionType, id string, name string, cols, rows int) (Session, error)
	// Attach session
	AttachSession(sessionType message.SessionType, id string, cols, rows int) (Session, error)
	// List sessions
	ListSessions(sessionType message.SessionType) (*model.PageResult, error)
	// Detach session
	DetachSession(sessionType message.SessionType, id string) error
	// Quit session
	QuitSession(sessionType message.SessionType, id string) error
	// Write to session
	InputSession(sessionType message.SessionType, id string, data string) error
	// ResizeSession
	ResizeSession(sessionType message.SessionType, id string, cols int, rows int) error
	// RenameSession
	RenameSession(sessionType message.SessionType, id string, data string) error
}
