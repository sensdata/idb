package terminal

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
)

var (
	//go:embed no_alt_buffer.screenrc
	noAltBufferScreenRC []byte
)

// DefaultManager
type DefaultManager struct {
	// sessionMap map[string]Session
	sessions sync.Map
}

func NewManager() Manager {
	return &DefaultManager{}
}

// Close all sessions
func (m *DefaultManager) ReleaseAllSessions() error {
	m.sessions.Range(func(key, value interface{}) bool {
		if session, ok := value.(Session); ok {
			if err := session.Release(); err != nil {
				global.LOG.Error("failed to close session: %v", err)
			}
		}
		return true
	})

	return nil
}

// Store session
func (m *DefaultManager) StoreSession(session Session) {
	m.sessions.Store(session.GetSession(), session)
}

// Remove session
func (m *DefaultManager) RemoveSession(id string) {
	m.sessions.Delete(id)
}

// Get session
func (m *DefaultManager) GetSession(id string) (Session, error) {
	// 查找会话
	if session, ok := m.sessions.Load(id); ok {
		return session.(Session), nil
	}

	return nil, fmt.Errorf("session not found: %s", id)
}

// Start session
func (m *DefaultManager) StartSession(sessionType message.SessionType, id string, name string, cols, rows int) (Session, error) {

	var session Session
	switch sessionType {
	case message.SessionTypeScreen:
		// check screenrc
		tmpRcPath := filepath.Join(os.TempDir(), "idb.screenrc")
		if err := os.WriteFile(tmpRcPath, noAltBufferScreenRC, 0644); err != nil {
			global.LOG.Error("failed to write screenrc: %v", err)
		}

		session = NewScreenSession(
			id,
			name,
			cols,
			rows,
		)
	case message.SessionTypeTmux:
		// not support
	case message.SessionTypeDocker:
		global.LOG.Info("container id: %s, shell: %s", id, name)
		session = NewContainerSession(id, "", name, cols, rows)
	default:
		session = NewBaseSession(
			id,
			name,
			cols,
			rows,
		)
	}

	// start session
	if err := session.Start(); err != nil {
		global.LOG.Error("failed to start session, %v", err)
		return nil, err
	}

	// stroe session
	m.StoreSession(session)
	global.LOG.Info("session %s.%s started", session.GetSession(), session.GetName())

	return session, nil
}

// Attach session
func (m *DefaultManager) AttachSession(sessionType message.SessionType, id string, cols, rows int) (Session, error) {
	var session Session
	switch sessionType {
	case message.SessionTypeScreen:
		session = NewScreenSession(
			id,
			"",
			cols,
			rows,
		)
	case message.SessionTypeTmux:
		// not support
	case message.SessionTypeDocker:
		// not support
	default:
		session = NewBaseSession(
			id,
			"",
			cols,
			rows,
		)
	}
	global.LOG.Info("attaching session")

	// attach session
	err := session.Attach()
	if err != nil {
		global.LOG.Error("failed to attach session, %v", err)
		if err.Error() == constant.ErrNotInstalled {
			return nil, err
		}
		err = session.Start()
		if err != nil {
			global.LOG.Error("failed to start session, %v", err)
			return nil, err
		}
	}

	// stroe session
	m.StoreSession(session)
	global.LOG.Info("session %s.%s attached", session.GetSession(), session.GetName())

	return session, nil
}

func (m *DefaultManager) ListSessions(sessionType message.SessionType) (*model.PageResult, error) {
	var result model.PageResult

	switch sessionType {
	case message.SessionTypeScreen:
		// get all sessions
		sessions, _ := listScreenSessions(false)
		result.Total = int64(len(sessions))
		result.Items = sessions

	case message.SessionTypeTmux:
		// get all sessions
		sessions, _ := listTmuxSession(false)
		result.Total = int64(len(sessions))
		result.Items = sessions

	case message.SessionTypeDocker:
		// not support

	default:
		// get all sessions
		sessions, _ := m.listBaseSessions(false)
		result.Total = int64(len(sessions))
		result.Items = sessions
	}
	return &result, nil
}

func (m *DefaultManager) DetachSession(sessionType message.SessionType, id string) error {
	switch sessionType {
	case message.SessionTypeScreen:
		return detachScreenSession(id)
	case message.SessionTypeTmux:
		return detachTmuxSession(id)
	case message.SessionTypeDocker:
		return m.quitContainerSession(id)
	default:
		return nil
	}
}

func (m *DefaultManager) QuitSession(sessionType message.SessionType, id string) error {
	switch sessionType {
	case message.SessionTypeScreen:
		return quitScreenSession(id)
	case message.SessionTypeTmux:
		return quitTmuxSession(id)
	case message.SessionTypeDocker:
		return m.quitContainerSession(id)
	default:
		return nil
	}
}

func (m *DefaultManager) InputSession(sessionType message.SessionType, id string, data string) error {
	// find session
	sessionInterface, ok := m.sessions.Load(id)
	if !ok {
		global.LOG.Error("session %s not found", id)
		return fmt.Errorf("session %s not found", id)
	}
	sessionInstance := sessionInterface.(Session)
	// input
	if err := sessionInstance.Input(data); err != nil {
		global.LOG.Error("failed to input session: %v", err)
		return fmt.Errorf("failed to input session: %v", err)
	}

	return nil
}

func (m *DefaultManager) ResizeSession(sessionType message.SessionType, id string, cols int, rows int) error {
	// find session
	sessionInterface, ok := m.sessions.Load(id)
	if !ok {
		global.LOG.Error("session %s not found", id)
		return fmt.Errorf("session %s not found", id)
	}
	sessionInstance := sessionInterface.(Session)
	// input
	if err := sessionInstance.Resize(cols, rows); err != nil {
		global.LOG.Error("failed to resize session: %v", err)
		return fmt.Errorf("failed to resize session: %v", err)
	}

	return nil
}

func (m *DefaultManager) RenameSession(sessionType message.SessionType, id string, data string) error {
	switch sessionType {
	case message.SessionTypeScreen:
		return renameScreenSession(id, data)
	case message.SessionTypeTmux:
		return renameTmuxSession(id, data)
	case message.SessionTypeDocker:
		// not support
		return nil
	default:
		return m.renameBaseSession(id, data)
	}
}

func (m *DefaultManager) listBaseSessions(filterDetached bool) ([]model.SessionInfo, error) {
	var sessions []model.SessionInfo

	m.sessions.Range(func(key, value interface{}) bool {
		if baseSession, ok := value.(*BaseSession); ok {

			// 如果只筛选 Detached 会话
			if filterDetached && baseSession.Status != "Detached" {
				return true
			}

			sessionInfo := model.SessionInfo{
				Session: baseSession.Session,
				Name:    baseSession.Name,
				Time:    baseSession.CreateAt,
				Status:  baseSession.Status,
			}
			sessions = append(sessions, sessionInfo)
		}
		return true
	})

	return sessions, nil
}

func listScreenSessions(filterDetached bool) ([]model.SessionInfo, error) {
	var sessions []model.SessionInfo

	// 执行命令以列出所有的 screen 会话
	cmd := exec.Command("screen", "-ls")
	output, err := cmd.Output()
	if strings.Contains(string(output), "No Sockets found") {
		return sessions, nil
	}
	if err != nil {
		global.LOG.Error("failed to list sessions: %v", err)
		return sessions, nil
	}
	global.LOG.Info("listScreenSessions: %s", string(output))

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

func listTmuxSession(filterDetached bool) ([]model.SessionInfo, error) {
	var sessions []model.SessionInfo

	// 执行命令以列出所有的 tmux 会话
	cmd := exec.Command("tmux", "ls")
	output, err := cmd.Output()
	if strings.Contains(string(output), "no server running") {
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
		// 假设会话信息格式为 "session_name: windows (created time) [attached]"
		if strings.Contains(line, ":") {
			// 使用正则表达式提取会话信息
			re := regexp.MustCompile(`([^:]+):\s+\d+\s+windows\s+\(created\s+(.+?)\)(?:\s+\[(attached|detached)\])?`)
			matches := re.FindStringSubmatch(line)

			if len(matches) < 3 {
				continue
			}

			// matches[1]是会话名称
			// matches[2]是创建时间
			// matches[3]是状态（如果存在）
			sessionName := strings.TrimSpace(matches[1])
			timeStr := matches[2]

			// 解析状态
			status := "Detached"
			if len(matches) > 3 && matches[3] == "attached" {
				status = "Attached"
			}

			// 如果只筛选 Detached 会话
			if filterDetached && status != "Detached" {
				continue
			}

			// 解析时间
			var parsedTime time.Time
			t, _ := parseTmuxTime(timeStr)
			parsedTime = t

			sessionInfo := model.SessionInfo{
				Session: "", // tmux没有数字ID
				Name:    sessionName,
				Time:    parsedTime,
				Status:  status,
			}
			sessions = append(sessions, sessionInfo)
		}
	}

	return sessions, nil
}

func detachScreenSession(id string) error {
	if err := exec.Command("screen", "-S", id, "-X", "detach").Run(); err != nil {
		global.LOG.Error("Error detaching screen session %s: %v", id, err)
		return err
	}
	global.LOG.Info("Session %s detached", id)
	return nil
}

func detachTmuxSession(id string) error {
	if err := exec.Command("tmux", "detach-session", "-t", id).Run(); err != nil {
		global.LOG.Error("Error detaching tmux session %s: %v", id, err)
		return err
	}
	global.LOG.Info("Tmux session %s detached", id)
	return nil
}

func quitScreenSession(id string) error {
	if err := exec.Command("screen", "-S", id, "-X", "quit").Run(); err != nil {
		global.LOG.Error("Error quit screen session %s: %v", id, err)
		return err
	}
	global.LOG.Info("Session %s quit", id)

	return nil
}

func quitTmuxSession(id string) error {
	if err := exec.Command("tmux", "kill-session", "-t", id).Run(); err != nil {
		global.LOG.Error("Error quitting tmux session %s: %v", id, err)
		return err
	}
	global.LOG.Info("Tmux session %s quit", id)
	return nil
}

func (m *DefaultManager) quitContainerSession(id string) error {
	session, err := m.GetSession(id)
	if err != nil {
		global.LOG.Error("failed to get session: %v", err)
		return err
	}
	if err := session.Release(); err != nil {
		global.LOG.Error("failed to close session: %v", err)
		return err
	}
	return nil
}

func (m *DefaultManager) renameBaseSession(_ string, _ string) error {
	return nil
}

func renameScreenSession(id string, name string) error {
	if err := exec.Command("screen", "-S", id, "-X", "sessionname", name).Run(); err != nil {
		global.LOG.Error("Error renaming screen session %s to %s: %v", id, name, err)
		return err
	}
	return nil
}

func renameTmuxSession(id string, name string) error {
	if err := exec.Command("tmux", "rename-session", "-t", id, name).Run(); err != nil {
		global.LOG.Error("Error renaming tmux session %s to %s: %v", id, name, err)
		return err
	}
	return nil
}
