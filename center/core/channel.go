package core

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
)

type Center struct {
	cfg         config.CenterConfig
	listener    net.Listener
	agentConns  map[string]net.Conn // 存储Agent端连接的映射
	agentMsgIDs map[string]string   // 存储Agent端连接的最后一个消息ID
	done        chan struct{}
}

func NewCenter(cfg config.CenterConfig) *Center {
	return &Center{
		cfg:         cfg,
		agentConns:  make(map[string]net.Conn),
		agentMsgIDs: make(map[string]string),
		done:        make(chan struct{}),
	}
}

func (c *Center) Start() error {
	// 将端口号转换为字符串
	portStr := strconv.Itoa(c.cfg.Port)

	// 监听端口
	lis, err := net.Listen("tcp", ":"+portStr)
	if err != nil {
		log.Error("Failed to listen on port %s: %v", portStr, err)
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return err
	}

	c.listener = lis
	log.Info("Center Started, listening on port %d", c.cfg.Port)

	// 启动接收连接的goroutine
	go c.acceptConnections()

	// 定期发送心跳消息
	go c.sendHeartbeat()

	return nil
}

func (c *Center) Stop() {
	close(c.done)
	// 关闭所有Agent连接
	for _, conn := range c.agentConns {
		conn.Close()
	}
	//关闭监听
	if c.listener != nil {
		log.Info("Stopping listening")
		c.listener.Close()
	}
}

func (c *Center) acceptConnections() {
	for {
		select {
		case <-c.done:
			log.Info("Shutting down server...")
			return
		default:
			conn, err := c.listener.Accept()
			if err != nil {
				select {
				case <-c.done:
					log.Info("Server is shutting down, stop accepting new connections.")
					return
				default:
					log.Error("Failed to accept connection: %v", err)
				}
				continue
			}
			// 记录Agent端的连接
			agentID := conn.RemoteAddr().String()
			c.agentConns[agentID] = conn
			// 处理连接
			go c.handleConnection(conn)
		}
	}
}

func (c *Center) handleConnection(conn net.Conn) {
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
		messages, err := message.ParseMessage(buffer, c.cfg.SecretKey)
		if err != nil {
			log.Error("Error processing message: %v", err)
		} else {
			// 记录Agent端的连接和最后一个消息ID
			if len(messages) > 1 {
				agentID := conn.RemoteAddr().String()
				c.agentMsgIDs[agentID] = messages[0].MsgID
			}

			// 处理消息
			for _, msg := range messages {
				log.Info("Received message: %s", msg.Data)

				switch msg.Type {
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

func (c *Center) sendHeartbeat() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		// 发送心跳消息给所有Agent端
		for agentID := range c.agentMsgIDs {
			heartbeatMsg, err := message.CreateMessage(
				utils.GenerateMsgId(),
				"Heartbeat",
				c.cfg.SecretKey,
				utils.GenerateNonce(16),
				message.Heartbeat,
			)
			if err != nil {
				fmt.Printf("Error creating heartbeat message: %v\n", err)
				continue
			}

			// 发送心跳消息
			conn := c.agentConns[agentID]
			err = message.SendMessage(conn, heartbeatMsg)
			if err != nil {
				fmt.Printf("Failed to send heartbeat message: %v\n", err)
			}
		}
	}
}
