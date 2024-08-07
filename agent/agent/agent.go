package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sensdata/idb/agent/action"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
)

var (
	CONFMAN *config.Manager
	AGENT   IAgent
)

type Agent struct {
	unixListener net.Listener
	tcpListener  net.Listener
	centerID     string   // 存储center地址
	centerConn   net.Conn // 存储center端连接的映射
	centerMsgID  string   // 存储center端连接的最后一个消息ID
	done         chan struct{}
	mu           sync.Mutex // 保护centerConn和centerMsgID的互斥锁
}

type IAgent interface {
	Start() error
	Stop() error
}

func NewAgent() IAgent {
	return &Agent{
		centerConn: nil,
		done:       make(chan struct{}),
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
	err = a.listenToTcp()
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) Stop() error {
	close(a.done)

	// 关闭center连接
	a.mu.Lock()
	if a.centerConn != nil {
		a.centerConn.Close()
	}
	a.mu.Unlock()

	// 关闭监听
	if a.unixListener != nil {
		a.unixListener.Close()
	}
	if a.tcpListener != nil {
		a.tcpListener.Close()
	}

	//删除sock文件
	sockFile := filepath.Join(constant.AgentDataDir, constant.AgentSock)
	os.Remove(sockFile)

	return nil
}

func (a *Agent) listenToUnix() error {
	//先关闭
	if a.unixListener != nil {
		a.unixListener.Close()
	}

	// 检查sock文件
	sockFile := filepath.Join(constant.AgentDataDir, constant.AgentSock)

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
				a.listenToTcp()
			}
		default:
			conn.Write([]byte("Unknown config command format"))
		}
	default:
		conn.Write([]byte("Unknown command"))
	}
}

func (a *Agent) listenToTcp() error {
	//先关闭
	a.mu.Lock()
	if a.centerConn != nil {
		a.centerConn.Close()
	}
	a.mu.Unlock()

	if a.tcpListener != nil {
		a.tcpListener.Close()
	}

	config := CONFMAN.GetConfig()
	global.LOG.Info("Try listen on port %d", config.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port))
	if err != nil {
		global.LOG.Error("Failed to listen on port %d: %v", config.Port, err)
		fmt.Printf("Failed to listen on port %d, quit \n", config.Port)
		return err
	}

	a.tcpListener = lis
	go a.acceptConnections()

	global.LOG.Info("Starting TCP server on port %d", config.Port)

	return nil
}

func (a *Agent) acceptConnections() {
	for {
		select {
		case <-a.done:
			global.LOG.Info("Agent is stopping, stop accepting new connections.")
			return
		default:
			conn, err := a.tcpListener.Accept()
			if err != nil {
				select {
				case <-a.done:
					global.LOG.Info("Agent is stopping, stop accepting new connections.")
					return
				default:
					global.LOG.Error("Failed to accept connection: %v", err)
				}
				continue
			}

			// 成功接受连接后记录日志
			now := time.Now().Format(time.RFC3339)
			centerID := conn.RemoteAddr().String()
			global.LOG.Info("Accepted new connection from %s at %s", centerID, now)
			// 记录连接
			a.mu.Lock()
			a.centerID = centerID
			a.centerConn = conn
			a.mu.Unlock()

			// 处理连接
			go a.handleConnection(conn)
		}
	}
}

func (a *Agent) handleConnection(conn net.Conn) {
	defer func() {
		a.mu.Lock()
		a.centerID = ""
		a.centerConn = nil
		a.mu.Unlock()

		conn.Close()
	}()

	config := CONFMAN.GetConfig()
	var buffer []byte
	tmp := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				global.LOG.Error("Read error: %v", err)
			}
			break
		}

		buffer = append(buffer, tmp[:n]...)

		// 尝试解析消息
		messages, err := message.ParseMessage(buffer, config.SecretKey)
		if err != nil {
			if err == message.ErrIncompleteMessage {
				global.LOG.Info("not enough data, continue to read")
				continue // 数据不完整，继续读取
			} else {
				global.LOG.Error("Error processing message: %v", err)
			}
		} else {
			// 记录center端的链接和最后一个消息ID
			centerID := conn.RemoteAddr().String()
			if len(messages) > 0 {
				a.mu.Lock()
				a.centerMsgID = messages[0].MsgID
				a.mu.Unlock()
			}

			// 处理消息
			for _, msg := range messages {
				switch msg.Type {
				case message.Heartbeat: // 回复心跳
					global.LOG.Info("Heartbeat from %s", centerID)
					if a.centerID != "" && centerID == a.centerID {
						a.sendHeartbeat(conn)
					} else {
						global.LOG.Error("%s is a unknown center", centerID)
					}
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
						a.sendActionResult(conn, msg.MsgID, &model.Action{Action: "", Data: ""})
					} else {
						a.sendActionResult(conn, msg.MsgID, result)
					}
				default: // 不支持的消息
					global.LOG.Error("Unknown message type: %s", msg.Type)
				}
			}
		}

		// 清空缓冲区
		buffer = buffer[:0]
	}

	global.LOG.Info("Connection closed: %s", conn.RemoteAddr().String())
}

func (a *Agent) sendHeartbeat(conn net.Conn) {
	config := CONFMAN.GetConfig()

	centerID := conn.RemoteAddr().String()

	heartbeatMsg, err := message.CreateMessage(
		utils.GenerateMsgId(),
		"Heartbeat",
		config.SecretKey,
		utils.GenerateNonce(16),
		message.Heartbeat,
	)
	if err != nil {
		global.LOG.Error("Error creating heartbeat message: %v", err)
		return
	}

	err = message.SendMessage(conn, heartbeatMsg)
	if err != nil {
		global.LOG.Error("Failed to send heartbeat message: %v", err)
		a.mu.Lock()
		conn.Close()
		global.LOG.Info("close conn %s for heartbeat", a.centerID)
		a.centerConn = nil
		a.centerID = ""
		a.mu.Unlock()
		//关闭后，尝试重新连接
		// go a.connectToCenter()
	} else {
		global.LOG.Info("Heartbeat sent to %s", centerID)
	}
}

func (a *Agent) sendCmdResult(conn net.Conn, msgID string, result string) {
	config := CONFMAN.GetConfig()

	centerID := conn.RemoteAddr().String()

	cmdRspMsg, err := message.CreateMessage(
		msgID, // 使用相同的msgID回复
		result,
		config.SecretKey,
		utils.GenerateNonce(16),
		message.CmdMessage,
	)
	if err != nil {
		global.LOG.Error("Error creating cmd rsp message: %v", err)
		return
	}

	global.LOG.Info("send msg data: %s", cmdRspMsg.Data)

	err = message.SendMessage(conn, cmdRspMsg)
	if err != nil {
		global.LOG.Error("Failed to send cmd rsp message: %v", err)
		a.mu.Lock()
		conn.Close()
		global.LOG.Info("close conn %s for cmd rsp", a.centerID)
		a.centerConn = nil
		a.centerID = ""
		a.mu.Unlock()
		//关闭后，尝试重新连接
		// go a.connectToCenter()
	} else {
		global.LOG.Info("Cmd rsp sent to %s", centerID)
	}
}

func (a *Agent) sendActionResult(conn net.Conn, msgID string, action *model.Action) {
	config := CONFMAN.GetConfig()

	centerID := conn.RemoteAddr().String()

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
		message.ActionMessage,
	)
	if err != nil {
		global.LOG.Error("Error creating cmd rsp message: %v", err)
		return
	}

	global.LOG.Info("send msg data: %s", cmdRspMsg.Data)

	err = message.SendMessage(conn, cmdRspMsg)
	if err != nil {
		global.LOG.Error("Failed to send cmd rsp message: %v", err)
		a.mu.Lock()
		conn.Close()
		global.LOG.Info("close conn %s for cmd rsp", a.centerID)
		a.centerConn = nil
		a.centerID = ""
		a.mu.Unlock()
		//关闭后，尝试重新连接
		// go a.connectToCenter()
	} else {
		global.LOG.Info("Cmd rsp sent to %s", centerID)
	}
}

func (a *Agent) processAction(data string) (*model.Action, error) {
	var actionData model.Action
	if err := json.Unmarshal([]byte(data), &actionData); err != nil {
		return nil, err
	}

	switch actionData.Action {
	// 获取overview
	case model.Action_SysInfo_OverView:
		overview, err := action.GetOverview()
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Data:   overview,
		}, nil
	// 获取network
	case model.Action_SysInfo_Network:
		network, err := action.GetNetwork()
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Data:   network,
		}, nil
	// 获取SystemInfo
	case model.Action_SysInfo_System:
		systemInfo, err := action.GetSystemInfo()
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Data:   systemInfo,
		}, nil
	default:
		return nil, nil
	}
}
