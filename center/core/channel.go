package core

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
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
	mu          sync.Mutex // 保护agentConns和agentMsgIDs的互斥锁
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
	lis, err := net.Listen("tcp", "0.0.0.0:"+portStr)
	if err != nil {
		log.Error("Failed to listen on port %s: %v", portStr, err)
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return err
	}

	c.listener = lis
	log.Info("Center Started, listening on port %d", c.cfg.Port)

	// 启动接收连接的 goroutine
	go c.acceptConnections()

	// 定期发送心跳消息
	go c.sendHeartbeat()

	return nil
}

func (c *Center) Stop() {
	close(c.done)
	// 关闭所有Agent连接
	c.mu.Lock()
	for _, conn := range c.agentConns {
		conn.Close()
	}
	c.mu.Unlock()
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
			log.Info("Server is shutting down, stop accepting new connections")
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

			// 成功接受连接后记录日志
			now := time.Now().Format(time.RFC3339)
			log.Info("Accepted new connection from %s at %s", conn.RemoteAddr().String(), now)
			// 记录到map中
			c.mu.Lock()
			c.agentConns[conn.RemoteAddr().String()] = conn
			c.mu.Unlock()

			// 处理连接
			go c.handleConnection(conn)
		}
	}
}

func (c *Center) handleConnection(conn net.Conn) {
	defer func() {
		agentID := conn.RemoteAddr().String()
		log.Info("close conn %s for err", agentID)

		c.mu.Lock()
		delete(c.agentConns, agentID)
		c.mu.Unlock()
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
		messages, err := message.ParseMessage(buffer, c.cfg.SecretKey)
		if err != nil {
			log.Error("Error processing message: %v", err)
		} else {
			// 记录Agent端的连接和最后一个消息ID
			agentID := conn.RemoteAddr().String()
			if len(messages) > 0 {
				c.mu.Lock()
				c.agentMsgIDs[agentID] = messages[0].MsgID
				c.mu.Unlock()
			}

			// 处理消息
			for _, msg := range messages {

				switch msg.Type {
				case message.Heartbeat: // 收到心跳
					// TODO: 维护在线状态
					log.Info("Heartbeat from %s", agentID)
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

	log.Info("Connection closed: %s", conn.RemoteAddr().String())
}

func (c *Center) sendHeartbeat() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			log.Info("Stopping heartbeat")
			return
		case <-ticker.C:
			log.Info("Sending heartbeats to %d agents", len(c.agentConns))
			c.mu.Lock()
			for agentID, conn := range c.agentConns {
				heartbeatMsg, err := message.CreateMessage(
					utils.GenerateMsgId(),
					"Heartbeat",
					c.cfg.SecretKey,
					utils.GenerateNonce(16),
					message.Heartbeat,
				)
				if err != nil {
					log.Error("Error creating heartbeat message: %v", err)
					continue
				}

				err = message.SendMessage(conn, heartbeatMsg)
				if err != nil {
					log.Error("Failed to send heartbeat message to %s: %v", agentID, err)
					conn.Close()
					delete(c.agentConns, agentID)
					log.Info("close conn %s for heartbeat", agentID)
				} else {
					log.Info("Heartbeat sent to %s", agentID)
				}
			}
			c.mu.Unlock()
		}
	}
}
