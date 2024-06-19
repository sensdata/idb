package channel

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
)

type Agent struct {
	cfg         config.Config
	listener    net.Listener
	centerID    string   // 存储center地址
	centerConn  net.Conn // 存储center端连接的映射
	centerMsgID string   // 存储center端连接的最后一个消息ID
	done        chan struct{}
	mu          sync.Mutex // 保护centerConn和centerMsgID的互斥锁
}

func NewAgent(cfg config.Config) *Agent {
	return &Agent{
		cfg:        cfg,
		centerConn: nil,
		done:       make(chan struct{}),
	}
}

func (a *Agent) Start() error {
	// 将端口号转换为字符串
	portStr := strconv.Itoa(a.cfg.Port)
	log.Info("Agent started, try listen on port %s", portStr)

	// 监听端口
	lis, err := net.Listen("tcp", "0.0.0.0:"+portStr)
	if err != nil {
		log.Error("Failed to listen on port %s: %v", portStr, err)
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return err
	}

	a.listener = lis
	log.Info("Starting TCP server on port %d", a.cfg.Port)

	// 启动接受连接的 goroutine
	go a.acceptConnections()

	// 连接server
	go a.connectToCenter()

	return nil
}

func (a *Agent) Stop() {
	close(a.done)
	// 关闭center连接
	a.mu.Lock()
	if a.centerConn != nil {
		a.centerConn.Close()
	}
	a.mu.Unlock()
	//关闭监听
	if a.listener != nil {
		log.Info("Stopping listening")
		a.listener.Close()
	}
}

func (a *Agent) connectToCenter() {
	const maxRetries = 5
	const retryInterval = time.Second * 5

	for retries := 0; retries < maxRetries; retries++ {
		select {
		case <-a.done:
			return
		default:
			if a.centerConn == nil {
				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", a.cfg.CenterIP, a.cfg.CenterPort))
				if err != nil {
					select {
					case <-a.done:
						log.Info("Server is shutting down, stop connect to Center.")
						return
					default:
						log.Error("Failed to connect to Center: %v", err)
						time.Sleep(retryInterval)
					}
				} else {
					// 记录连接
					centerID := conn.RemoteAddr().String()
					a.mu.Lock()
					a.centerID = centerID
					a.centerConn = conn
					a.mu.Unlock()
					log.Info("Successfully connected to Center %s", centerID)

					// 处理连接
					go a.handleConnection(conn)

					return
				}
			}
		}
	}
	log.Error("Max retries reached. Unable to connect to Center.")
}

func (a *Agent) acceptConnections() {
	for {
		select {
		case <-a.done:
			log.Info("Server is shutting down, stop accepting new connections.")
			return
		default:
			conn, err := a.listener.Accept()
			if err != nil {
				select {
				case <-a.done:
					log.Info("Server is shutting down, stop accepting new connections.")
					return
				default:
					log.Error("Failed to accept connection: %v", err)
				}
				continue
			}

			// 成功接受连接后记录日志
			now := time.Now().Format(time.RFC3339)
			centerID := conn.RemoteAddr().String()
			log.Info("Accepted new connection from %s at %s", centerID, now)

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
				log.Error("Read error: %v", err)
			}
			break
		}

		buffer = append(buffer, tmp[:n]...)

		// 尝试解析消息
		messages, err := message.ParseMessage(buffer, a.cfg.SecretKey)
		if err != nil {
			if err == message.ErrIncompleteMessage {
				log.Info("not enough data, continue to read")
				continue // 数据不完整，继续读取
			} else {
				log.Error("Error processing message: %v", err)
			}
		} else {
			// 记录center端的链接和最后一个消息ID
			if len(messages) > 0 {
				a.mu.Lock()
				a.centerMsgID = messages[0].MsgID
				a.mu.Unlock()
			}

			// 处理消息
			for _, msg := range messages {
				centerID := conn.RemoteAddr().String()
				switch msg.Type {
				case message.Heartbeat: // 回复心跳
					log.Info("Heartbeat from %s", centerID)
					if a.centerID != "" && centerID == a.centerID {
						a.sendHeartbeat(conn)
					} else {
						log.Error("%s is a unknown center", centerID)
					}
				case message.CmdMessage: // 处理 Cmd 类型的消息
					result, err := shell.ExecuteCommand(msg.Data)
					if err != nil {
						log.Error("Failed to execute command: %v", err)
						continue
					}
					log.Info("Command output: %s", result)
					a.sendCmdResult(conn, msg.MsgID, result)
				case message.ActionMessage: // 处理 Action 类型的消息
					log.Info("Processing action message: %s", msg.Data)
					// TODO: 在这里添加处理 action 消息的逻辑

				default: // 不支持的消息
					log.Error("Unknown message type: %s", msg.Type)
				}
			}
		}

		// 清空缓冲区
		buffer = buffer[:0]
	}

	log.Info("Connection closed: %s", conn.RemoteAddr().String())
}

func (a *Agent) sendHeartbeat(conn net.Conn) {
	centerID := conn.RemoteAddr().String()

	heartbeatMsg, err := message.CreateMessage(
		utils.GenerateMsgId(),
		"Heartbeat",
		a.cfg.SecretKey,
		utils.GenerateNonce(16),
		message.Heartbeat,
	)
	if err != nil {
		log.Error("Error creating heartbeat message: %v", err)
		return
	}

	err = message.SendMessage(conn, heartbeatMsg)
	if err != nil {
		log.Error("Failed to send heartbeat message: %v", err)
		a.mu.Lock()
		conn.Close()
		log.Info("close conn %s for heartbeat", a.centerID)
		a.centerConn = nil
		a.centerID = ""
		a.mu.Unlock()
		//关闭后，尝试重新连接
		go a.connectToCenter()
	} else {
		log.Info("Heartbeat sent to %s", centerID)
	}
}

func (a *Agent) sendCmdResult(conn net.Conn, msgID string, result string) {
	log.Info("send cmd result: %s", result)
	centerID := conn.RemoteAddr().String()

	cmdRspMsg, err := message.CreateMessage(
		msgID, // 使用相同的msgID回复
		result,
		a.cfg.SecretKey,
		utils.GenerateNonce(16),
		message.CmdMessage,
	)
	if err != nil {
		log.Error("Error creating cmd rsp message: %v", err)
		return
	}

	log.Info("msg data: %s", cmdRspMsg.Data)

	err = message.SendMessage(conn, cmdRspMsg)
	if err != nil {
		log.Error("Failed to send cmd rsp message: %v", err)
		a.mu.Lock()
		conn.Close()
		log.Info("close conn %s for cmd rsp", a.centerID)
		a.centerConn = nil
		a.centerID = ""
		a.mu.Unlock()
		//关闭后，尝试重新连接
		go a.connectToCenter()
	} else {
		log.Info("Cmd rsp sent to %s", centerID)
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
