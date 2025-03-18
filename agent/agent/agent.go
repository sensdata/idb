package agent

import (
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	gonet "github.com/shirou/gopsutil/v4/net"

	"github.com/sensdata/idb/agent/agent/action"
	"github.com/sensdata/idb/agent/agent/ca"
	"github.com/sensdata/idb/agent/agent/docker"
	"github.com/sensdata/idb/agent/agent/file"
	"github.com/sensdata/idb/agent/agent/git"
	"github.com/sensdata/idb/agent/agent/session"
	"github.com/sensdata/idb/agent/agent/ssh"
	"github.com/sensdata/idb/agent/agent/terminal"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/systemctl"
)

var (
	CONFMAN        *config.Manager
	AGENT          IAgent
	FileService    = file.NewIFileService()
	SshService     = ssh.NewISSHService()
	GitService     = git.NewIGitService()
	DockerService  = docker.NewIDockerService()
	CaService      = ca.NewICaService()
	SessionService = session.NewISessionService()
)

type Agent struct {
	unixListener net.Listener
	done         chan struct{}
	resetConn    chan struct{}
	// mu             sync.Mutex // 保护centerConn的互斥锁
	// sessionMap     map[string]*session.Session
	sessionManager terminal.Manager

	rx float64
	tx float64
}

//go:embed screen_install.sh
var installScreenShell []byte

//go:embed screen_clean.sh
var cleanScreenShell []byte

type IAgent interface {
	Start() error
	Stop() error
}

func NewAgent() IAgent {
	return &Agent{
		done:      make(chan struct{}),
		resetConn: make(chan struct{}, 3),
		// sessionMap:     make(map[string]*session.Session),
		sessionManager: terminal.NewManager(),
	}
}

func (a *Agent) Start() error {

	fmt.Println("Agent Starting")

	// 启动 Unix 域套接字监听器
	err := a.listenToUnix()
	if err != nil {
		return err
	}

	// 监听端口
	go a.listenToTcp()

	// 监听流量
	go a.monitorTraffic()

	return nil
}

func (a *Agent) Stop() error {
	close(a.done)

	// 关闭监听
	if a.unixListener != nil {
		a.unixListener.Close()
	}

	//删除sock文件
	sockFile := filepath.Join(constant.AgentRunDir, constant.AgentSock)
	os.Remove(sockFile)

	return nil
}

func (a *Agent) listenToUnix() error {
	//先关闭
	if a.unixListener != nil {
		a.unixListener.Close()
	}

	// 检查sock文件
	sockFile := filepath.Join(constant.AgentRunDir, constant.AgentSock)

	var err error
	a.unixListener, err = net.Listen("unix", sockFile)
	if err != nil {
		global.LOG.Error("Failed to start unix listener: %v", err)
		return err
	}

	// 处理unix连接
	go a.acceptUnixConnections()

	return nil
}

func (a *Agent) acceptUnixConnections() {
	for {
		select {
		case <-a.done:
			global.LOG.Info("Agent is stopping, stop accepting new unix connections.")
			return
		default:
			conn, err := a.unixListener.Accept()
			if err != nil {
				select {
				case <-a.done:
					global.LOG.Info("Agent is stopping, stop accepting new unix connections.")
					return
				default:
					global.LOG.Error("failed to accept unix connection: %v", err)
				}
				continue
			}
			go a.handleUnixConnection(conn)
		}
	}
}

func (a *Agent) handleUnixConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		global.LOG.Error("failed to read from unix connection: %v", err)
		return
	}

	command := string(buf[:n])
	parts := strings.Fields(command)

	if len(parts) == 0 {
		conn.Write([]byte("Unknown command"))
		return
	}

	switch parts[0] {
	case "status":
		conn.Write([]byte("Agent is running"))
	case "stop":
		conn.Write([]byte("Agent stopped"))
		// 发送 SIGTERM 信号以停止 Agent
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	case "config":
		switch len(parts) {
		case 1:
			// 输出当前的配置信息
			config, err := CONFMAN.GetConfigString("")
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Failed to get config: %v", err)))
			} else {
				conn.Write([]byte(fmt.Sprintf("%v", config)))
			}
		case 2:
			// 输出当前的指定key配置信息
			key := parts[1]
			value, err := CONFMAN.GetConfigString(key)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Failed to get %s: %v", key, err)))
			} else {
				conn.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
			}
		case 3:
			// 修改指定key的配置
			key := parts[1]
			value := parts[2]
			err := CONFMAN.SetConfig(key, value)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Failed to set config %s: %v", key, err)))
			} else {
				conn.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
				go systemctl.Restart(constant.AgentService)
			}
		default:
			conn.Write([]byte("Unknown config command format"))
		}
	default:
		conn.Write([]byte("Unknown command"))
	}
}

func (a *Agent) monitorTraffic() {
	tick := time.NewTicker(time.Second * 1) // 每秒触发一次
	defer tick.Stop()

	var lastRxBytes, lastTxBytes uint64
	var lastTime time.Time // 用于计算时间差

	for {
		select {
		case <-a.done:
			global.LOG.Info("Agent is stopping, stop monitorTraffic")
			return
		case <-tick.C:
			// 获取网络流量统计数据
			ioCounters, err := gonet.IOCounters(true)
			if err != nil {
				log.Println("Error getting network stats:", err)
				continue
			}

			for _, counter := range ioCounters {
				// 如果接收或发送字节数大于零，表示接口有流量
				if counter.BytesRecv > lastRxBytes || counter.BytesSent > lastTxBytes {
					// 计算时间差，得到时间间隔（秒）
					timeDiff := time.Since(lastTime).Seconds()

					// 防止时间差为0，避免除零错误
					if timeDiff > 0 {
						// 计算每秒的流量速率
						rxRate := float64(counter.BytesRecv-lastRxBytes) / timeDiff
						txRate := float64(counter.BytesSent-lastTxBytes) / timeDiff

						// 更新全局的Rx和Tx速率
						a.rx = rxRate
						a.tx = txRate

						// 打印当前的网络速率（单位：字节/秒）
						// global.LOG.Info("Interface: %s - RX: %.2f B/s, TX: %.2f B/s", counter.Name, a.rx, a.tx)
					}

					// 更新上次的字节数和时间
					lastRxBytes = counter.BytesRecv
					lastTxBytes = counter.BytesSent
					lastTime = time.Now() // 更新最后的时间戳
				}
			}
		}
	}
}

func (a *Agent) listenToTcp() {
	config := CONFMAN.GetConfig()

	// 创建 TLS 配置
	cert, err := tls.X509KeyPair(global.CertPem, global.KeyPem)
	if err != nil {
		global.LOG.Error("Failed to create cert: %v", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert}, // 设置服务器证书
		MinVersion:         tls.VersionTLS13,        // 设置最小 TLS 版本
		InsecureSkipVerify: true,
	}

	// 使用 tls.Listen 替代 net.Listen
	listener, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port), tlsConfig)
	if err != nil {
		global.LOG.Error("Failed to listen on port %d: %v", config.Port, err)
		return
	}

	global.LOG.Info("Starting TCP server on port %d", config.Port)

	defer func() {
		global.LOG.Info("Tcp listener closing")
		listener.Close()
	}()
	for {
		select {
		case <-a.done:
			global.LOG.Info("Agent is stopping, stop accepting new connections.")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				global.LOG.Error("Failed to accept connection: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 成功接受连接后记录日志
			now := time.Now().Format(time.RFC3339)
			connAddr := conn.RemoteAddr().String()
			global.LOG.Info("Accepted new connection from %s at %s", connAddr, now)

			// 处理连接
			go a.handleConnection(conn)
			return
		}
	}
}

func (a *Agent) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("[Panic] in handleConnection: %v", r)
		}
	}()

	// 传递给 SessionService
	SessionService.Config(&conn, CONFMAN.GetConfig().SecretKey)

	config := CONFMAN.GetConfig()

	// 缓存区：用来缓存从 conn.Read 读取的数据
	dataBuffer := make([]byte, 0)
	tmpBuffer := make([]byte, 1024)
	for {
		select {
		// 断连并退出
		case <-a.done:
			global.LOG.Info("Agent is stopping, stop handleConnection")
			conn.Close()
			return

		// 重置连接
		case <-a.resetConn:
			global.LOG.Info("Close and Accept")
			conn.Close()
			go a.listenToTcp()
			return

		// 读取数据
		default:
			n, err := conn.Read(tmpBuffer)
			if err != nil {
				global.LOG.Error("Error read from conn: %v", err)
				// if err != io.EOF {
				// 	global.LOG.Error("Error read from conn: %v", err)
				// }
				go a.resetConnection()
				continue
			}
			// 将数据拼接到缓存区
			dataBuffer = append(dataBuffer, tmpBuffer[:n]...)

			// 尝试提取完整消息
			msgType, packet, remainingBuffer, err := message.ExtractCompleteMessagePacket(dataBuffer)
			if err != nil {
				if err == message.ErrIncompleteMessage {
					// 数据不完整，继续读取
					continue
				}

				// 错误，重试重连
				global.LOG.Error("Error extract complete message: %v", err)
				go a.resetConnection()
				continue
			}

			// 处理解析后的消息
			msgData := packet[message.MagicBytesLen+message.MsgLenBytes:]
			msg, err := message.DecodeMessage(msgType, msgData, config.SecretKey)
			if err != nil {
				global.LOG.Error("Error decode message: %v", err)
			}
			switch m := msg.(type) {
			case *message.Message:
				go a.processMessage(conn, m)
			case *message.FileMessage:
				go a.processFileMessage(conn, m)
			case *message.SessionMessage:
				go a.processSessionMessage(conn, m)
			default:
				fmt.Println("Unknown message type")
			}

			// 更新缓存，移除已处理的部分
			dataBuffer = remainingBuffer
		}
	}
}

func (a *Agent) resetConnection() {
	a.resetConn <- struct{}{}
}

func (a *Agent) processMessage(conn net.Conn, msg *message.Message) {
	global.LOG.Info("Message: %v", msg)

	switch msg.Type {
	case message.Heartbeat: // 回复心跳
		global.LOG.Info("Heartbeat from %s", conn.RemoteAddr().String())
		a.sendHeartbeat(conn)

	case message.CmdMessage: // 处理 Cmd 类型的消息
		global.LOG.Info("recv cmd message: %s", msg.Data)
		if strings.Contains(msg.Data, message.Separator) {
			commands := strings.Split(msg.Data, message.Separator)
			results, err := shell.ExecuteCommands(commands)
			if err != nil {
				global.LOG.Error("Failed to excute multi commands: %v", err)
				a.sendCmdResult(conn, msg.MsgID, "error")
			} else {
				result := strings.Join(results, message.Separator)
				a.sendCmdResult(conn, msg.MsgID, result)
			}
		} else {
			result, err := shell.ExecuteCommand(msg.Data)
			if err != nil {
				global.LOG.Error("Failed to execute command: %v", err)
				a.sendCmdResult(conn, msg.MsgID, "error")
			} else {
				a.sendCmdResult(conn, msg.MsgID, result)
			}
		}

	case message.ActionMessage: // 处理 Action 类型的消息
		global.LOG.Info("recv action message: %s", msg.Data)
		result, err := a.processAction(msg.Data)
		if err != nil {
			global.LOG.Error("Failed to process action: %v", err)
			a.sendActionResult(conn, msg.MsgID, &model.Action{Action: "", Result: false, Data: err.Error()})
		} else {
			a.sendActionResult(conn, msg.MsgID, result)
		}
	default: // 不支持的消息
		global.LOG.Error("Unknown message type: %s", msg.Type)
	}
}

func (a *Agent) processFileMessage(conn net.Conn, msg *message.FileMessage) {
	global.LOG.Info("FileMessage: %s, %d, %d", msg.FileName, msg.Offset, msg.ChunkSize)

	switch msg.Type {
	case message.Upload: //上传
		err := files.NewFileOp().WriteChunkToFile(
			msg.Path,
			msg.FileName,
			msg.Offset,
			msg.ChunkSize,
			msg.Chunk,
		)
		if err != nil {
			global.LOG.Error("Failed to process upload: %v", err)
			a.sendUploadResult(conn, msg, message.FileErr)
		} else {
			status := message.FileOk
			//如果是最后一次传输，则设置为Done
			if msg.Offset+int64(msg.ChunkSize) == msg.TotalSize {
				status = message.FileDone
			}
			a.sendUploadResult(conn, msg, status)
		}
	case message.Download: //下载
		totalSize, bytesRead, chunk, err := files.NewFileOp().ReadChunkFromFile(
			filepath.Join(msg.Path, msg.FileName),
			msg.Offset,
			msg.ChunkSize,
		)
		global.LOG.Info("read from file: %d %d", totalSize, bytesRead)
		if err != nil {
			global.LOG.Error("Failed to process download: %v", err)
			msg.Status = message.FileErr
		} else {
			msg.TotalSize = totalSize
			msg.ChunkSize = bytesRead
			msg.Chunk = chunk
			msg.Status = message.FileOk
			//如果是最后一次传输，则设置为Done
			if msg.Offset+int64(msg.ChunkSize) == msg.TotalSize {
				msg.Status = message.FileDone
			}
		}
		a.sendDownloadResult(conn, msg)
	}
}

// func (a *Agent) registerSession(session *session.Session) {
// 	a.mu.Lock()
// 	defer a.mu.Unlock()

// 	a.sessionMap[session.ID] = session
// 	global.LOG.Info("Session %s registered", session.ID)
// }

// func (a *Agent) unregisterSession(sessionID string) {
// 	a.mu.Lock()
// 	defer a.mu.Unlock()

// 	delete(a.sessionMap, sessionID)
// 	global.LOG.Info("Session %s unregistered", sessionID)
// }

// func (a *Agent) getSession(sessionID string) (*session.Session, bool) {
// 	a.mu.Lock()
// 	defer a.mu.Unlock()

// 	session, exists := a.sessionMap[sessionID]
// 	return session, exists
// }

func (a *Agent) processSessionMessage(conn net.Conn, msg *message.SessionMessage) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("[Panic] in processSessionMessage: %v", r)
		}
	}()

	global.LOG.Info("processSessionMessage: %v", msg)

	switch msg.Type {
	case message.WsMessageStart: // 创建会话
		global.LOG.Info("session begin")
		session, err := a.sessionManager.StartSession(
			msg.Data.Type,
			msg.Data.Data,
			msg.Data.Cols,
			msg.Data.Rows,
		)
		if err != nil {
			global.LOG.Error("Failed to start session: %v", err)
			a.sendSessionResult(conn, msg.MsgID, msg.Type, constant.CodeFailed, err.Error(), msg.Data.Type, "", "")
			return
		}

		// 先返回本会话信息
		a.sendSessionResult(
			conn,
			msg.MsgID,
			msg.Type,
			constant.CodeSuccess,
			"",
			session.GetType(),
			session.GetSession(),
			session.GetName(),
		)

		go a.waitForSessionOutput(conn, session)

	case message.WsMessageAttach: // 恢复会话
		// find old session
		oldSession, _ := a.sessionManager.GetSession(msg.Data.Session)
		if oldSession != nil {
			// detach old session
			err := a.sessionManager.DetachSession(oldSession.GetType(), oldSession.GetSession())
			if err != nil {
				global.LOG.Error("Failed to detach session %s for re-attaching", oldSession.GetSession())
			}

			// delay
			time.Sleep(300 * time.Millisecond)
		}

		global.LOG.Info("session begin")
		session, err := a.sessionManager.AttachSession(
			msg.Data.Type,
			msg.Data.Session,
			msg.Data.Cols,
			msg.Data.Rows,
		)
		if err != nil {
			global.LOG.Error("Failed to attach session: %v", err)
			code := constant.CodeFailed
			if err.Error() == constant.ErrNotInstalled {
				code = constant.CodeErrEnvironment
			}
			a.sendSessionResult(conn, msg.MsgID, msg.Type, code, err.Error(), msg.Data.Type, "", "")
			return
		}

		// 先返回本会话信息
		a.sendSessionResult(
			conn,
			msg.MsgID,
			msg.Type,
			constant.CodeSuccess,
			"",
			session.GetType(),
			session.GetSession(),
			session.GetName(),
		)

		go a.waitForSessionOutput(conn, session)

	case message.WsMessageCmd: // 会话输入
		a.sessionInput(msg.Data)

	case message.WsMessageResize: // 调整尺寸
		a.sessionResize(msg.Data)

	default:
		global.LOG.Error("not supported session mesage")
	}

	global.LOG.Info("processSessionMessage end")
}

func (a *Agent) waitForSessionOutput(conn net.Conn, session terminal.Session) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Error("[Panic] in waitForSessionOutput: %v", r)
		}
	}()

	for {
		select {
		case <-session.GetDoneChan():
			a.sessionManager.RemoveSession(session.GetSession())
			session.Release()
			global.LOG.Info("session end")
			return
		case output := <-session.GetOutputChan():
			global.LOG.Info("session output: %s", string(output))
			a.sendSessionResult(
				conn,
				utils.GenerateMsgId(),
				message.WsMessageCmd,
				constant.CodeSuccess,
				"",
				session.GetType(),
				session.GetSession(),
				string(output),
			)
		}
	}
}

func (a *Agent) sessionInput(sessionData message.SessionData) {
	if err := a.sessionManager.InputSession(sessionData.Type, sessionData.Session, sessionData.Data); err != nil {
		global.LOG.Error("Failed to input to session %s: %v", sessionData.Session, err)
	}
}

func (a *Agent) sessionResize(sessionData message.SessionData) {
	if err := a.sessionManager.ResizeSession(sessionData.Type, sessionData.Session, sessionData.Cols, sessionData.Rows); err != nil {
		global.LOG.Error("Failed to resize session %s: %v", sessionData.Session, err)
	}
}

func (a *Agent) isScreenInstalled() bool {
	// 检查 screen 是否安装
	cmd := exec.Command("screen", "-v")
	if err := cmd.Run(); err != nil {
		global.LOG.Error("screen is not installed: %v", err)
		return false
	}
	return true
}

func (a *Agent) installScreen() error {
	// 将installScreenShell保存到 /tmp/iDB_screen_timestamp.sh
	// 生成临时脚本文件名
	timestamp := time.Now().Unix()
	scriptPath := fmt.Sprintf("/tmp/iDB_screen_%d.sh", timestamp)
	logPath := fmt.Sprintf("/tmp/iDB_screen_%d.log", timestamp)

	// 写入脚本内容
	err := os.WriteFile(scriptPath, installScreenShell, 0755)
	if err != nil {
		global.LOG.Error("Failed to prepare installation script, %v", err)
		return fmt.Errorf("failed to prepare script")
	}
	defer os.Remove(scriptPath)

	// 执行安装脚本
	req := model.ScriptExec{ScriptPath: scriptPath, LogPath: logPath}
	scriptResult := shell.ExecuteScript(req)
	if scriptResult.Err != "" {
		return fmt.Errorf("failed to install")
	}

	return nil
}

func (a *Agent) cleanScreen() error {
	// 将cleanScreenShell保存到 /tmp/iDB_screen_clean_timestamp.sh
	// 生成临时脚本文件名
	timestamp := time.Now().Unix()
	scriptPath := fmt.Sprintf("/tmp/iDB_screen_clean_%d.sh", timestamp)
	logPath := fmt.Sprintf("/tmp/iDB_screen_clean_%d.log", timestamp)

	// 写入脚本内容
	err := os.WriteFile(scriptPath, cleanScreenShell, 0755)
	if err != nil {
		global.LOG.Error("Failed to prepare clean script, %v", err)
		return fmt.Errorf("failed to prepare script")
	}
	defer os.Remove(scriptPath)

	// 执行安装脚本
	req := model.ScriptExec{ScriptPath: scriptPath, LogPath: logPath}
	scriptResult := shell.ExecuteScript(req)
	if scriptResult.Err != "" {
		return fmt.Errorf("failed to install")
	}

	return nil
}

func (a *Agent) processAction(data string) (*model.Action, error) {
	var actionData model.Action
	if err := utils.FromJSONString(data, &actionData); err != nil {
		return nil, err
	}

	switch actionData.Action {
	// 获取host status
	case model.Host_Status:
		status, err := action.GetStatus()
		if err != nil {
			return nil, err
		}
		status.Rx = math.Round(a.rx*100) / 100
		status.Tx = math.Round(a.tx*100) / 100

		result, err := utils.ToJSONString(status)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 获取overview
	case model.SysInfo_OverView:
		overview, err := action.GetOverview()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(overview)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 获取network
	case model.SysInfo_Network:
		network, err := action.GetNetwork()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(network)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 获取SystemInfo
	case model.SysInfo_System:
		systemInfo, err := action.GetSystemInfo()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(systemInfo)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 设置时间
	case model.SysInfo_Set_Time:
		var setTimeReq model.SetTimeReq
		if err := json.Unmarshal([]byte(actionData.Data), &setTimeReq); err != nil {
			return nil, err
		}
		if err := action.SetTime(setTimeReq); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 设置时区
	case model.SysInfo_Set_Time_Zone:
		var setTimezoneReq model.SetTimezoneReq
		if err := json.Unmarshal([]byte(actionData.Data), &setTimezoneReq); err != nil {
			return nil, err
		}
		if err := action.SetTimezone(setTimezoneReq); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 同步时间
	case model.SysInfo_Sync_Time:
		if err := action.SyncTime(); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 清理缓存
	case model.Sysinfo_Clear_Mem_Cache:
		if err := action.ClearMemCache(); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 设置自动清理
	case model.SysInfo_Set_Auto_Clear:
		var autoClearReq model.AutoClearMemCacheReq
		if err := json.Unmarshal([]byte(actionData.Data), &autoClearReq); err != nil {
			return nil, err
		}
		if err := action.SetAutoClearInterval(autoClearReq); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.SysInfo_Get_Auto_Clear:
		autoClear, err := action.GetAutoClearInterval()
		if err != nil {
			return nil, err
		}
		conf := model.AutoClearMemCacheConf{Interval: autoClear}
		result, err := utils.ToJSONString(conf)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 创建swap
	case model.SysInfo_Create_Swap:
		var createSwapReq model.CreateSwapReq
		if err := json.Unmarshal([]byte(actionData.Data), &createSwapReq); err != nil {
			return nil, err
		}
		if err := action.CreateSwap(createSwapReq); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 删除swap
	case model.Sysinfo_Delete_Swap:
		if err := action.DeleteSwap(); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 设置dns
	case model.Sysinfo_Update_Dns:
		var dnsReq model.UpdateDnsSettingsReq
		if err := json.Unmarshal([]byte(actionData.Data), &dnsReq); err != nil {
			return nil, err
		}
		if err := action.UpdateDnsSettings(dnsReq); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 获取系统设置
	case model.Sysinfo_Get_Sys_Setting:
		sysSetting, err := action.GetSystemSettings()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(sysSetting)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 更新系统设置
	case model.Sysinfo_Upd_Sys_Setting:
		var sysSetting model.UpdateSystemSettingsReq
		if err := json.Unmarshal([]byte(actionData.Data), &sysSetting); err != nil {
			return nil, err
		}
		if err := action.UpdateSystemSettings(sysSetting); err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 文件树
	case model.File_Tree:
		var fileOption model.FileOption
		if err := json.Unmarshal([]byte(actionData.Data), &fileOption); err != nil {
			return nil, err
		}

		fileInfo, err := FileService.GetFileTree(fileOption)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(fileInfo)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 列出文件
	case model.File_List:
		var fileOption model.FileOption
		if err := json.Unmarshal([]byte(actionData.Data), &fileOption); err != nil {
			return nil, err
		}

		fileInfo, err := FileService.GetFileList(fileOption)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(fileInfo)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 搜索文件
	case model.File_Search:
		var fileOption model.FileOption
		if err := json.Unmarshal([]byte(actionData.Data), &fileOption); err != nil {
			return nil, err
		}

		files, err := FileService.SearchFiles(fileOption)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(files)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// TODO: 上传文件
	case model.File_Upload:
		return actionSuccessResult(actionData.Action, "")

		// TODO: 下载文件
	case model.File_Download:
		return actionSuccessResult(actionData.Action, "")

	// 创建文件
	case model.File_Create:
		var req model.FileCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.Create(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 删除文件
	case model.File_Delete:
		var req model.FileDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.Delete(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 批量删除
	case model.File_Batch_Delete:
		var req model.FileBatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.BatchDelete(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 批量修改用户/组
	case model.File_Batch_Change_Owner:
		var req model.FileRoleReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.BatchChangeOwner(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 批量修改mode
	case model.File_Batch_Change_Mode:
		var req model.FileModeReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.BatchChangeMode(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 修改文件权限
	case model.File_Change_Mode:
		var req model.FileCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.ChangeMode(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 修改文件用户/组
	case model.File_Change_Owner:
		var req model.FileRoleUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.ChangeOwner(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 修改文件名称
	case model.File_Change_Name:
		var req model.FileRename

		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.ChangeName(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 压缩文件
	case model.File_Compress:
		var req model.FileCompress

		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.Compress(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 解压文件
	case model.File_Decompress:
		var req model.FileDeCompress
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.DeCompress(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 获取文件内容
	case model.File_Content:
		var req model.FileContentReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		fileInfo, err := FileService.GetContent(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(fileInfo)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 保存文件内容
	case model.File_Content_Modify:
		var req model.FileEdit
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.SaveContent(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 移动文件
	case model.File_Move:
		var req model.FileMove
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.MvFile(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 目录大小
	case model.File_Dir_Size:
		var req model.DirSizeReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		rsp, err := FileService.DirSize(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(rsp)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 收藏列表
	case model.Favorite_List:
		var req model.PageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		rsp, err := FileService.GetFavoriteList(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(rsp)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 创建收藏
	case model.Favorite_Create:
		var req model.FavoriteCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		rsp, err := FileService.CreateFavorite(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(rsp)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 删除收藏
	case model.Favorite_Delete:
		var req model.FavoriteDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.DeleteFavorite(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 获取ssh配置
	case model.Ssh_Config:
		info, err := SshService.GetConfig()
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 更新ssh配置
	case model.Ssh_Config_Update:
		var req model.SSHUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.UpdateConfig(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 获取ssh配置文件内容
	case model.Ssh_Config_Content:
		content, err := SshService.GetContent()
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(content)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// 更新ssh配置文件内容
	case model.Ssh_Config_Content_Update:
		var req model.ContentUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.UpdateContent(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 操作ssh
	case model.Ssh_Operate:
		var req model.SSHOperate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.OperateSSH(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 创建秘钥
	case model.Ssh_Secret_Create:
		var req model.GenerateKey
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.CreateKey(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	// 枚举秘钥
	case model.Ssh_Secret:
		var req model.ListKey
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		keys, err := SshService.ListKeys(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(keys)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	// ssh日志
	case model.Ssh_Log:
		var req model.SearchSSHLog
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		log, _ := SshService.LoadLog(req)
		result, err := utils.ToJSONString(log)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 初始化
	case model.Git_Init:
		var req model.GitInit
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.InitRepo(req.RepoPath, req.IsBare)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 文件列表
	case model.Git_File_List:
		var req model.GitQuery
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		files, err := GitService.GetFileList(req.RepoPath, req.RelativePath, req.Extension, req.Page, req.PageSize)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(files)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 文件信息
	case model.Git_File:
		var req model.GitGetFile
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		file, err := GitService.GetFile(req.RepoPath, req.RelativePath)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(file)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 创建文件
	case model.Git_Create:
		var req model.GitCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Create(req.RepoPath, req.RelativePath, req.Dir, req.Content)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 更新文件
	case model.Git_Update:
		var req model.GitUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Update(req.RepoPath, req.RelativePath, req.Dir, req.NewName, req.Content)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 删除文件
	case model.Git_Delete:
		var req model.GitDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Delete(req.RepoPath, req.RelativePath, req.Dir)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 恢复文件
	case model.Git_Restore:
		var req model.GitRestore
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Restore(req.RepoPath, req.RelativePath, req.CommitHash)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 文件历史
	case model.Git_Log:
		var req model.GitLog
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		logs, err := GitService.Log(req.RepoPath, req.RelativePath, req.Page, req.PageSize)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(logs)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 文件差异
	case model.Git_Diff:
		var req model.GitDiff
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		diffs, err := GitService.Diff(req.RepoPath, req.RelativePath, req.CommitHash)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(diffs)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 执行脚本
	case model.Script_Exec:
		var req model.ScriptExec
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		scriptResult := shell.ExecuteScript(req)
		result, err := utils.ToJSONString(scriptResult)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// docker 状态
	case model.Docker_Status:
		status, err := DockerService.DockerStatus()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(status)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// docker 配置
	case model.Docker_Conf:
		conf, err := DockerService.DockerConf()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(conf)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 更新 docker 配置
	case model.Docker_Upd_Conf:
		var req model.KeyValue
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateConf(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 更新 docker 配置
	case model.Docker_Upd_Conf_File:
		var req model.DaemonJsonUpdateByFile
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateConfByFile(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 更新 docker log配置
	case model.Docker_Upd_Log:
		var req model.LogOption
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateLogOption(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 更新 docker ipv6配置
	case model.Docker_Upd_Ipv6:
		var req model.Ipv6Option
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateIpv6Option(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// docker 操作
	case model.Docker_Operation:
		var req model.DockerOperation
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerOperation(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// inspect
	case model.Docker_Inspect:
		var req model.Inspect
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.Inspect(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// prune
	case model.Docker_Prune:
		var req model.Prune
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		report, err := DockerService.Prune(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(report)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Query:
		var req model.QueryContainer
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ContainerQuery(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Names:
		names, err := DockerService.ContainerNames()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(names)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Create:
		var req model.ContainerOperate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerCreate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Update:
		var req model.ContainerOperate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerUpdate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Upgrade:
		var req model.ContainerUpgrade
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerUpgrade(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Info:
		containerID := actionData.Data
		info, err := DockerService.ContainerInfo(containerID)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Resource_Usage:
		info, err := DockerService.ContainerResourceUsage()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Resource_Limit:
		info, err := DockerService.ContainerResourceLimit()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Stats:
		containerID := actionData.Data
		info, err := DockerService.ContainerStats(containerID)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Rename:
		var req model.Rename
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerRename(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Log_Clean:
		containerID := actionData.Data
		err := DockerService.ContainerLogClean(containerID)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Operation:
		var req model.ContainerOperation
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerOperation(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Logs:
		// var req model.ContainerOperation
		// if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
		// 	return nil, err
		// }
		// err := DockerService.ContainerLogs(req)
		// if err != nil {
		// 	return nil, err
		// }
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Image_Page:
		var req model.SearchPageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImagePage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_List:
		info, err := DockerService.ImageList()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Build:
		var req model.ImageBuild
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImageBuild(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Pull:
		var req model.ImagePull
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImagePull(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Load:
		var req model.ImageLoad
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImageLoad(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Save:
		var req model.ImageSave
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ImageSave(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Image_Push:
		var req model.ImagePush
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImagePush(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Remove:
		var req model.BatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ImageRemove(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Image_Tag:
		var req model.ImageTag
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ImageTag(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Volume_Page:
		var req model.SearchPageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.VolumePage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Volume_List:
		info, err := DockerService.VolumeList()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Volume_Delete:
		var req model.BatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.VolumeDelete(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Volume_Create:
		var req model.VolumeCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.VolumeCreate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Network_Page:
		var req model.SearchPageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.NetworkPage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Network_List:
		info, err := DockerService.NetworkList()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Network_Delete:
		var req model.BatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.NetworkDelete(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Network_Create:
		var req model.NetworkCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.NetworkCreate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Compose_Page:
		var req model.QueryCompose
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ComposePage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Compose_Test:
		var req model.ComposeCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ComposeTest(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Compose_Create:
		var req model.ComposeCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ComposeCreate(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Compose_Operation:
		var req model.ComposeOperation
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ComposeOperation(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Compose_Update:
		var req model.ComposeUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ComposeUpdate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.CA_Groups:
		page, err := CaService.GetCertificateGroups()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(page)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.CA_Group_Pk:
		var req model.GroupPkRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := CaService.GetPrivateKeyInfo(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.CA_Group_Csr:
		var req model.GroupPkRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := CaService.GetCSRInfo(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.CA_Group_Create:
		var req model.CreateGroupRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := CaService.GenerateCertificate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.CA_Group_Remove:
		var req model.DeleteGroupRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := CaService.RemoveCertificateGroup(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.CA_Self_Sign:
		var req model.SelfSignedRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := CaService.GenerateSelfSignedCertificate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.CA_Info:
		var req model.CertificateInfoRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := CaService.GetCertificateInfo(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.CA_Complete:
		var req model.CertificateInfoRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := CaService.CompleteCertificateChain(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.CA_Remove:
		var req model.DeleteCertificateRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := CaService.RemoveCertificate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.CA_Import:
		var req model.ImportCertificateRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := CaService.ImportCertificate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Terminal_List:
		var req model.TerminalRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		list, err := a.sessionManager.ListSessions(message.SessionType(req.Type))
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(list)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Terminal_Detach:
		var req model.TerminalRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := a.sessionManager.DetachSession(message.SessionType(req.Type), req.Session)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Terminal_Finish:
		var req model.TerminalRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := a.sessionManager.QuitSession(message.SessionType(req.Type), req.Session)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Terminal_Rename:
		var req model.TerminalRequest
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := a.sessionManager.RenameSession(message.SessionType(req.Type), req.Session, req.Data)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Terminal_Install:
		err := a.installScreen()
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Terminal_Prune:
		err := a.cleanScreen()
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	default:
		return nil, nil
	}
}

func actionSuccessResult(action string, data string) (*model.Action, error) {
	return &model.Action{
		Action: action,
		Result: true,
		Data:   data,
	}, nil
}

func (a *Agent) sendHeartbeat(conn net.Conn) {
	config := CONFMAN.GetConfig()

	heartbeatMsg, err := message.CreateMessage(
		utils.GenerateMsgId(),
		"Heartbeat",
		config.SecretKey,
		utils.GenerateNonce(16),
		global.Version,
		message.Heartbeat,
	)
	if err != nil {
		global.LOG.Error("Error creating heartbeat message: %v", err)
		return
	}

	err = message.SendMessage(conn, heartbeatMsg)
	if err != nil {
		global.LOG.Error("Failed to send heartbeat message: %v", err)
		a.resetConnection()
	} else {
		global.LOG.Info("Heartbeat sent to %s", conn.RemoteAddr().String())
	}
}

func (a *Agent) sendCmdResult(conn net.Conn, msgID string, result string) {
	config := CONFMAN.GetConfig()

	cmdRspMsg, err := message.CreateMessage(
		msgID, // 使用相同的msgID回复
		result,
		config.SecretKey,
		utils.GenerateNonce(16),
		global.Version,
		message.CmdMessage,
	)
	if err != nil {
		global.LOG.Error("Error creating cmd rsp message: %v", err)
		return
	}

	global.LOG.Info("send cmd data: %s", cmdRspMsg.Data)

	err = message.SendMessage(conn, cmdRspMsg)
	if err != nil {
		global.LOG.Error("Failed to send cmd rsp message: %v", err)
		a.resetConnection()
	}
}

func (a *Agent) sendActionResult(conn net.Conn, msgID string, action *model.Action) {
	config := CONFMAN.GetConfig()

	data, err := json.Marshal(action)
	if err != nil {
		global.LOG.Error("Error marshal action:: %v", err)
		return
	}

	cmdRspMsg, err := message.CreateMessage(
		msgID, // 使用相同的msgID回复
		string(data),
		config.SecretKey,
		utils.GenerateNonce(16),
		global.Version,
		message.ActionMessage,
	)
	if err != nil {
		global.LOG.Error("Error creating cmd rsp message: %v", err)
		return
	}

	global.LOG.Info("send action data: %s", cmdRspMsg.Data)

	err = message.SendMessage(conn, cmdRspMsg)
	if err != nil {
		global.LOG.Error("Failed to send cmd rsp message: %v", err)
		a.resetConnection()
	}
}

func (a *Agent) sendUploadResult(conn net.Conn, msg *message.FileMessage, status int) {
	rspMsg, err := message.CreateFileMessage(
		msg.MsgID,
		msg.Type,
		status,
		msg.Path,
		msg.FileName,
		msg.TotalSize,
		msg.Offset,
		0,
		nil,
	)
	if err != nil {
		global.LOG.Error("Error creating file rsp message: %v", err)
		return
	}

	err = message.SendFileMessage(conn, rspMsg)
	if err != nil {
		global.LOG.Error("Failed to send file rsp : %v", err)
		a.resetConnection()
	}
}

func (a *Agent) sendDownloadResult(conn net.Conn, msg *message.FileMessage) {
	err := message.SendFileMessage(conn, msg)
	if err != nil {
		global.LOG.Error("Failed to send file rsp : %v", err)
		a.resetConnection()
	}
}

func (a *Agent) sendSessionResult(conn net.Conn, msgID string, msgType string, code int, msg string, sessionType message.SessionType, sessionID string, data string) {
	config := CONFMAN.GetConfig()

	rspMsg, err := message.CreateSessionMessage(
		msgID,
		msgType,
		message.SessionData{Code: code, Msg: msg, Type: sessionType, Session: sessionID, Data: data},
		config.SecretKey,
		utils.GenerateNonce(16),
		global.Version,
	)
	if err != nil {
		global.LOG.Error("Error creating session rsp message: %v", err)
		return
	}

	global.LOG.Info("send session data: %v", rspMsg.Data)

	err = message.SendSessionMessage(conn, rspMsg)
	if err != nil {
		global.LOG.Error("Failed to send session rsp : %v", err)
		a.resetConnection()
	}
}
