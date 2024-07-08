package conn

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/utils"
)

type Center struct {
	confManager   *config.Manager
	unixListener  net.Listener
	agentConns    map[string]net.Conn // 存储Agent端连接的映射
	agentMsgIDs   map[string]string   // 存储Agent端连接的最后一个消息ID
	done          chan struct{}
	mu            sync.Mutex             // 保护agentConns和agentMsgIDs的互斥锁
	responseChMap map[string]chan string // 用于接收命令执行结果的动态通道
}

type ICenter interface {
	Start() error
	Stop() error
	ExecuteCommand(cmd string) (string, error)
}

func NewCenter() ICenter {
	return &Center{
		agentConns:    make(map[string]net.Conn),
		agentMsgIDs:   make(map[string]string),
		done:          make(chan struct{}),
		responseChMap: make(map[string]chan string),
	}
}

func (c *Center) Start() error {

	fmt.Printf("Center Starting")

	// 启动 Unix 域套接字监听器
	err := c.listenToUnix()
	if err != nil {
		return err
	}

	// 启动接收连接的 goroutine
	err = c.ensureAgentConnections()
	if err != nil {
		return nil
	}

	return nil
}

func (c *Center) Stop() error {
	close(c.done)

	// 关闭所有Agent连接
	c.mu.Lock()
	for _, conn := range c.agentConns {
		conn.Close()
	}
	c.mu.Unlock()

	//关闭监听
	if c.unixListener != nil {
		c.unixListener.Close()
	}

	//删除sock文件
	sockFile := filepath.Join(constant.BaseDir, constant.CenterSock)
	os.Remove(sockFile)

	return nil
}

func (a *Center) listenToUnix() error {
	//先关闭
	if a.unixListener != nil {
		a.unixListener.Close()
	}

	// 检查sock文件
	sockFile := filepath.Join(constant.BaseDir, constant.CenterSock)
	if err := utils.EnsureFile(sockFile); err != nil {
		global.LOG.Error("Failed to create sock file: %v", err)
		return err
	}

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

func (a *Center) acceptUnixConnections() {
	for {
		select {
		case <-a.done:
			global.LOG.Info("Center is stopping, stop accepting new unix connections.")
			return
		default:
			conn, err := a.unixListener.Accept()
			if err != nil {
				select {
				case <-a.done:
					global.LOG.Info("Center is stopping, stop accepting new unix connections.")
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

func (a *Center) handleUnixConnection(conn net.Conn) {
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
		conn.Write([]byte("Center is running"))
	case "stop":
		conn.Write([]byte("Center stopped"))
		// 发送 SIGTERM 信号以停止 Center
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	case "config":
		switch len(parts) {
		case 1:
			// 输出当前的配置信息
			config, err := a.getConfig()
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Failed to get config: %v", err)))
			} else {
				conn.Write([]byte(fmt.Sprintf("%v", config)))
			}
		case 2:
			// 输出当前的指定key配置信息
			key := parts[1]
			value, err := a.getConfigValue(key)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Failed to get %s: %v", key, err)))
			} else {
				conn.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
			}
		case 3:
			// 修改指定key的配置
			key := parts[1]
			value := parts[2]
			err := a.setConfigValue(key, value)
			if err != nil {
				conn.Write([]byte(fmt.Sprintf("Failed to set config %s: %v", key, err)))
			} else {
				conn.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
			}
		default:
			conn.Write([]byte("Unknown config command format"))
		}
	default:
		conn.Write([]byte("Unknown command"))
	}
}

// 返回当前的所有配置信息
func (a *Center) getConfig() (string, error) {
	return a.confManager.GetConfigString("")
}

// 返回当前的指定配置信息
func (a *Center) getConfigValue(key string) (string, error) {
	return a.confManager.GetConfigString(key)
}

// 设置指定的配置
func (a *Center) setConfigValue(key, value string) error {
	err := a.confManager.SetConfig(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Center) ensureAgentConnections() error {
	global.LOG.Info("ensureAgentConnections")

	//获取所有的host
	hosts, err := HostRepo.GetList()
	if err != nil {
		global.LOG.Error("Failed to get host list: %v", err)
		return err
	}

	// 挨个确认是否已经建立连接
	for _, host := range hosts {
		addr := host.AgentAddr
		// 判断sshClients中是否包含addr的数据
		_, exists := c.agentConns[addr]
		if exists {
			continue
		} else {
			resultCh := make(chan error, 1)
			go c.connectToAgent(&host, resultCh)
			// handle the result if needed
			err := <-resultCh
			if err != nil {
				global.LOG.Error("Failed to connect to agent %s: %v", host.Addr, err)
			}
		}
	}

	// 定期发送心跳消息
	go c.sendHeartbeat()

	return nil
}

func (c *Center) connectToAgent(host *model.Host, resultCh chan<- error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	global.LOG.Info("try connect to agent %s:%d", host.AgentAddr, host.AgentPort)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort))
	if err != nil {
		global.LOG.Error("Failed to connect to Agent: %v", err)
		resultCh <- err
		return
	}

	// 记录连接
	agentID := conn.RemoteAddr().String()
	c.agentConns[agentID] = conn

	global.LOG.Info("Successfully connected to Agent %s", agentID)
	resultCh <- nil

	// 处理连接
	go c.handleConnection(conn)
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

	config := c.confManager.GetConfig()
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
	config := c.confManager.GetConfig()

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
					config.SecretKey,
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

	config := c.confManager.GetConfig()

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
		config.SecretKey,
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
