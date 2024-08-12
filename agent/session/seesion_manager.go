package session

import "sync"

type SessionManager struct {
	sessions map[string]*SessionContext
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionContext),
	}
}

// 添加会话
func (sm *SessionManager) AddSession(session *SessionContext) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.ID] = session
}

// 获取会话
func (sm *SessionManager) GetSession(id string) (*SessionContext, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session, exists := sm.sessions[id]
	return session, exists
}

// 删除会话
func (sm *SessionManager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}
