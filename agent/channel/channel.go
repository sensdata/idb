package channel

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
)

type Agent struct {
	cfg      config.Config
	listener net.Listener
	done     chan struct{}
}

func NewAgent(cfg config.Config) *Agent {
	return &Agent{
		cfg:  cfg,
		done: make(chan struct{}),
	}
}

func (a *Agent) Start() error {
	// 将端口号转换为字符串
	portStr := strconv.Itoa(a.cfg.Port)
	log.Info("Agent started, try listen on port %s", portStr)

	// 监听端口
	lis, err := net.Listen("tcp", ":"+portStr)
	if err != nil {
		log.Error("Failed to listen on port %s: %v", portStr, err)
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return err
	}

	a.listener = lis
	log.Info("Starting TCP server on port %d", a.cfg.Port)

	// 启动接受连接的 goroutine
	go a.acceptConnections()

	return nil
}

func (a *Agent) Stop() {
	close(a.done)
	if a.listener != nil {
		log.Info("Stopping listening")
		a.listener.Close()
	}
}

func (a *Agent) acceptConnections() {
	for {
		select {
		case <-a.done:
			log.Info("Shutting down server...")
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
			// 处理连接
			go a.handleConnection(conn)
		}
	}
}

func (a *Agent) handleConnection(conn net.Conn) {
	defer conn.Close()

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
			log.Error("Error processing message: %v", err)
		} else {
			// 处理消息
			for _, msg := range messages {
				log.Info("Received message: %s", msg.Data)

				switch msg.Type {
				case message.Heartbeat: // 回复心跳
					a.sendHeartbeat(conn)
				case message.CmdMessage: // 处理 Cmd 类型的消息
					result, err := shell.ExecuteCommand(msg.Data)
					if err != nil {
						log.Error("Failed to execute command: %v", err)
						continue
					}
					log.Info("Command output: %s", result)

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
}

func (a *Agent) sendHeartbeat(conn net.Conn) {
	heartbeatMsg, err := message.CreateMessage(
		utils.GenerateMsgId(),
		"Heartbeat",
		a.cfg.SecretKey,
		utils.GenerateNonce(16),
		message.Heartbeat,
	)
	if err != nil {
		fmt.Printf("Error creating heartbeat message: %v\n", err)
		return
	}

	err = message.SendMessage(conn, heartbeatMsg)
	if err != nil {
		fmt.Printf("Failed to send heartbeat message: %v\n", err)
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
