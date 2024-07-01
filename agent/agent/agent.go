package agent

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
)

var AGENT = Agent{
	state:      0,
	centerConn: nil,
	done:       make(chan struct{}),
}

type Agent struct {
	state        int // 状态
	unixListener net.Listener
	tcpListener  net.Listener
	centerID     string   // 存储center地址
	centerConn   net.Conn // 存储center端连接的映射
	centerMsgID  string   // 存储center端连接的最后一个消息ID
	done         chan struct{}
	mu           sync.Mutex // 保护centerConn和centerMsgID的互斥锁
}

type IAgent interface {
	Started() bool
	Start() error
	Stop() error
}

func (a *Agent) Started() bool {
	return a.state == 1
}

func (a *Agent) Start() error {

	fmt.Println("Agent Starting")

	configPath := "config.json"

	manager, err := config.NewManager(configPath)
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v \n", err)
	}

	cfg := manager.GetConfig()
	global.CONF = *cfg
	fmt.Println("Get config:")
	fmt.Printf("%+v \n", *cfg)

	// 初始化日志模块
	l, err := log.InitLogger(cfg.LogPath)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
		panic(err)
	}
	global.LOG = l

	// 启动 Unix 域套接字监听器
	a.unixListener, err = net.Listen("unix", "/tmp/idb-agent.sock")
	if err != nil {
		global.LOG.Error("Failed to start unix listener: %v", err)
		return err
	}

	// 将端口号转换为字符串
	portStr := strconv.Itoa(cfg.Port)
	global.LOG.Info("Agent started, try listen on port %s", portStr)

	// 监听端口
	lis, err := net.Listen("tcp", "0.0.0.0:"+portStr)
	if err != nil {
		global.LOG.Error("Failed to listen on port %s: %v", portStr, err)
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return err
	}

	a.tcpListener = lis
	global.LOG.Info("Starting TCP server on port %s", portStr)

	// 处理unix连接
	go a.acceptUnixConnections()

	// 启动接受连接的 goroutine
	go a.acceptConnections()

	a.state = 1
	return nil
}

func (a *Agent) Stop() error {
	a.stopAgent()
	return nil

}

func (a *Agent) stopAgent() {
	close(a.done)

	// 关闭center连接
	a.mu.Lock()
	if a.centerConn != nil {
		a.centerConn.Close()
	}
	a.mu.Unlock()

	//关闭监听
	if a.unixListener != nil {
		a.unixListener.Close()
	}
	if a.tcpListener != nil {
		a.tcpListener.Close()
	}
	a.state = 0
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
	args := strings.Fields(command)

	if len(args) == 0 {
		conn.Write([]byte("No command provided"))
		return
	}

	// 处理命令，根据命令执行相应操作
	switch args[0] {
	case "status":
		conn.Write([]byte("Agent is running"))
	case "stop":
		conn.Write([]byte("Agent stopped"))
		// 发送 SIGTERM 信号以停止 Agent
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	case "config":
		a.handleConfigCommand(conn, args[1:])
	default:
		conn.Write([]byte("Unknown command"))
	}
}

func (a *Agent) handleConfigCommand(conn net.Conn, args []string) {
	switch len(args) {
	case 0:
		// 读取当前所有配置
		conn.Write([]byte("Current configuration: ...")) // 这里替换为实际的配置读取逻辑
	case 1:
		// 读取指定配置项
		key := args[0]
		// 假设有一个 GetConfig 方法用于获取配置
		value, err := a.GetConfig(key)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("Failed to get config for key %s: %v", key, err)))
		} else {
			conn.Write([]byte(fmt.Sprintf("Current value for %s: %s", key, value)))
		}
	case 2:
		// 设置指定配置项
		key := args[0]
		value := args[1]
		// 假设有一个 SetConfig 方法用于设置配置
		if err := a.SetConfig(key, value); err != nil {
			conn.Write([]byte(fmt.Sprintf("Failed to set config for key %s: %v", key, err)))
		} else {
			conn.Write([]byte(fmt.Sprintf("Config %s set to %s", key, value)))
		}
	default:
		conn.Write([]byte("Invalid config command"))
	}
}

// 假设有 GetConfig 和 SetConfig 方法
func (a *Agent) GetConfig(key string) (string, error) {
	// 在这里实现获取配置的逻辑
	return "dummy_value", nil // 替换为实际值
}

func (a *Agent) SetConfig(key, value string) error {
	// 在这里实现设置配置的逻辑
	return nil // 替换为实际操作
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
		messages, err := message.ParseMessage(buffer, global.CONF.SecretKey)
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
					result, err := shell.ExecuteCommand(msg.Data)
					if err != nil {
						global.LOG.Error("Failed to execute command: %v", err)
						continue
					}
					global.LOG.Info("Command output: %s", result)
					a.sendCmdResult(conn, msg.MsgID, result)
				case message.ActionMessage: // 处理 Action 类型的消息
					global.LOG.Info("Processing action message: %s", msg.Data)
					// TODO: 在这里添加处理 action 消息的逻辑

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
	centerID := conn.RemoteAddr().String()

	heartbeatMsg, err := message.CreateMessage(
		utils.GenerateMsgId(),
		"Heartbeat",
		global.CONF.SecretKey,
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
	global.LOG.Info("send cmd result: %s", result)
	centerID := conn.RemoteAddr().String()

	cmdRspMsg, err := message.CreateMessage(
		msgID, // 使用相同的msgID回复
		result,
		global.CONF.SecretKey,
		utils.GenerateNonce(16),
		message.CmdMessage,
	)
	if err != nil {
		global.LOG.Error("Error creating cmd rsp message: %v", err)
		return
	}

	global.LOG.Info("msg data: %s", cmdRspMsg.Data)

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

// func testSendMessage(cfg *config.Config) {
// 	time.Sleep(time.Second * 5)
// 	// 构造消息
// 	commands := []string{
// 		"echo 'Hello, test 2'",
// 		"ps",
// 	}

// 	for _, cmd := range commands {
// 		msg, err := message.CreateMessage(
// 			utils.GenerateMsgId(),
// 			cmd,
// 			cfg.SecretKey,
// 			utils.GenerateNonce(16),
// 			message.CmdMessage,
// 		)
// 		if err != nil {
// 			fmt.Printf("Error creating message: %v\n", err)
// 			return
// 		}

// 		// 发送消息
// 		err = message.SendMessage("127.0.0.1", cfg.Port, msg)
// 		if err != nil {
// 			fmt.Printf("Failed to send message: %v\n", err)
// 		} else {
// 			fmt.Println("Message sent successfully!")
// 		}
// 	}
// }
