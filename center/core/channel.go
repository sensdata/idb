package core

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils"
)

type Center struct {
	cfg           config.CenterConfig
	listener      net.Listener
	agentConns    map[string]net.Conn // 存储Agent端连接的映射
	agentMsgIDs   map[string]string   // 存储Agent端连接的最后一个消息ID
	done          chan struct{}
	mu            sync.Mutex             // 保护agentConns和agentMsgIDs的互斥锁
	responseChMap map[string]chan string // 用于接收命令执行结果的动态通道
}

func NewCenter(cfg config.CenterConfig) *Center {
	return &Center{
		cfg:           cfg,
		agentConns:    make(map[string]net.Conn),
		agentMsgIDs:   make(map[string]string),
		done:          make(chan struct{}),
		responseChMap: make(map[string]chan string),
	}
}

func (c *Center) Start() error {
	// 将端口号转换为字符串
	portStr := strconv.Itoa(c.cfg.Port)

	// 监听端口
	lis, err := net.Listen("tcp", "0.0.0.0:"+portStr)
	if err != nil {
		global.LOG.Error("Failed to listen on port %s: %v", portStr, err)
		fmt.Printf("Failed to listen on port %s, quit \n", portStr)
		return err
	}

	c.listener = lis
	global.LOG.Info("Center Started, listening on port %d", c.cfg.Port)

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
		global.LOG.Info("Stopping listening")
		c.listener.Close()
	}
}

func (c *Center) acceptConnections() {
	for {
		select {
		case <-c.done:
			global.LOG.Info("Server is shutting down, stop accepting new connections")
			return
		default:
			conn, err := c.listener.Accept()
			if err != nil {
				select {
				case <-c.done:
					global.LOG.Info("Server is shutting down, stop accepting new connections.")
					return
				default:
					global.LOG.Error("Failed to accept connection: %v", err)
				}
				continue
			}

			// 成功接受连接后记录日志
			now := time.Now().Format(time.RFC3339)
			global.LOG.Info("Accepted new connection from %s at %s", conn.RemoteAddr().String(), now)
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
		global.LOG.Info("close conn %s for err", agentID)

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
				global.LOG.Error("Read error: %v", err)
			}
			break
		}

		buffer = append(buffer, tmp[:n]...)

		// 尝试解析消息
		messages, err := message.ParseMessage(buffer, c.cfg.SecretKey)
		if err != nil {
			if err == message.ErrIncompleteMessage {
				global.LOG.Info("not enough data, continue to read")
				continue // 数据不完整，继续读取
			} else {
				global.LOG.Error("Error processing message: %v", err)
			}
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
					global.LOG.Info("Heartbeat from %s", agentID)
				case message.CmdMessage: // 收到Cmd 类型的回复
					// TODO: 回调给发送者或者关注者
					global.LOG.Info("Processing cmd message: %s", msg.Data)
					// 获取响应通道
					c.mu.Lock()
					responseCh, exists := c.responseChMap[msg.MsgID]
					if exists {
						responseCh <- msg.Data
						close(responseCh)
						delete(c.responseChMap, msg.MsgID)
					}
					c.mu.Unlock()
				case message.ActionMessage: // 处理 Action 类型的消息
					// TODO: 回调给发送者或者关注者
					global.LOG.Info("Processing action message: %s", msg.Data)
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

func (c *Center) sendHeartbeat() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			global.LOG.Info("Stopping heartbeat")
			return
		case <-ticker.C:
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
					global.LOG.Error("Error creating heartbeat message: %v", err)
					continue
				}

				err = message.SendMessage(conn, heartbeatMsg)
				if err != nil {
					global.LOG.Error("Failed to send heartbeat message to %s: %v", agentID, err)
					conn.Close()
					delete(c.agentConns, agentID)
					global.LOG.Info("close conn %s for heartbeat", agentID)
				} else {
					global.LOG.Info("Heartbeat sent to %s", agentID)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Center) ExecuteCommand(cmd string) (string, error) {
	// 创建一个等待通道
	responseCh := make(chan string)

	// 假设只需要发送给一个 Agent
	c.mu.Lock()
	var conn net.Conn
	for _, agentConn := range c.agentConns {
		conn = agentConn
		break
	}
	c.mu.Unlock()

	if conn == nil {
		return "", fmt.Errorf("no agent connected")
	}

	// 创建消息
	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		cmd,
		c.cfg.SecretKey,
		utils.GenerateNonce(16),
		message.CmdMessage,
	)
	if err != nil {
		return "", err
	}

	// 将通道和msgID映射存储在map中
	c.mu.Lock()
	c.responseChMap[msgID] = responseCh
	c.mu.Unlock()

	go func() {
		err = message.SendMessage(conn, msg)
		if err != nil {
			global.LOG.Error("Failed to send command message: %v", err)
			responseCh <- ""
		}
	}()

	// 等待响应
	select {
	case response := <-responseCh:
		return response, nil
	case <-time.After(10 * time.Second): // 设置一个超时时间
		c.mu.Lock()
		delete(c.responseChMap, msgID)
		c.mu.Unlock()
		return "", fmt.Errorf("timeout waiting for response from agent")
	}
}
