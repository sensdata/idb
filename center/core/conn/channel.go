package conn

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type Center struct {
	unixListener      net.Listener
	agentConns        map[string]net.Conn // 存储Agent端连接的映射
	done              chan struct{}
	mu                sync.Mutex             // 保护agentConns的互斥锁
	responseChMap     map[string]chan string // 用于接收命令执行结果的动态通道
	fileResponseChMap map[string]chan *message.FileMessage
}

type ICenter interface {
	Start() error
	Stop() error
	ExecuteCommand(req core.Command) (string, error)
	ExecuteCommandGroup(req core.CommandGroup) ([]string, error)
	ExecuteAction(req core.HostAction) (*core.Action, error)
	UploadFile(hostID uint, path string, file *multipart.FileHeader) error
	DownloadFile(ctx *gin.Context, hostID uint, path string) error
	TestAgent(id uint, req core.TestAgent) error
}

func NewCenter() ICenter {
	return &Center{
		agentConns:        make(map[string]net.Conn),
		done:              make(chan struct{}),
		responseChMap:     make(map[string]chan string),
		fileResponseChMap: make(map[string]chan *message.FileMessage),
	}
}

func (c *Center) Start() error {

	global.LOG.Info("Center Starting")

	// 启动 Unix 域套接字监听器
	err := c.listenToUnix()
	if err != nil {
		return err
	}

	// 连接
	err = c.ensureAgentConnections()
	if err != nil {
		return err
	}
	// 连接并发送心跳
	go c.ensureConnectionsAndHeartbeat()

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
	sockFile := filepath.Join(constant.CenterDataDir, constant.CenterSock)
	os.Remove(sockFile)

	return nil
}

func (a *Center) listenToUnix() error {
	//先关闭
	if a.unixListener != nil {
		a.unixListener.Close()
	}

	// 检查sock文件
	sockFile := filepath.Join(constant.CenterDataDir, constant.CenterSock)

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
			}
		default:
			conn.Write([]byte("Unknown config command format"))
		}
	default:
		conn.Write([]byte("Unknown command"))
	}
}

func (c *Center) ensureConnectionsAndHeartbeat() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			global.LOG.Info("Stopping heartbeat")
			return
		case <-ticker.C:
			// 连接
			err := c.ensureAgentConnections()
			if err != nil {
				continue
			}
			// 心跳
			c.sendHeartbeat()
		}
	}
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
		// 查找agent conn
		conn, _ := c.getAgentConn(host)
		if conn != nil {
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

	config := CONFMAN.GetConfig()

	// 缓存区：用来缓存从 conn.Read 读取的数据
	dataBuffer := make([]byte, 0)

	for {
		// 读取数据
		tmpBuffer := make([]byte, 1024)
		n, err := conn.Read(tmpBuffer)
		if err != nil {
			if err != io.EOF {
				global.LOG.Error("Error read from conn: %v", err)
			}
			break
		}
		// 将数据拼接到缓存区
		dataBuffer = append(dataBuffer, tmpBuffer[:n]...)

		// 尝试解析消息
		for {
			// 提取完整消息
			msgType, packet, remainingBuffer, err := message.ExtractCompleteMessagePacket(dataBuffer)
			if err != nil {
				if err == message.ErrIncompleteMessage {
					// 数据不完整，继续读取
					break
				} else {
					global.LOG.Error("Error extract complete message: %v", err)
					break
				}
			}

			// 处理解析后的消息
			msgData := packet[message.MagicBytesLen+message.MsgLenBytes:]
			msg, err := message.DecodeMessage(msgType, msgData, config.SecretKey)
			if err != nil {
				global.LOG.Error("Error decode message: %v", err)
			}
			switch m := msg.(type) {
			case *message.Message:
				c.processMessage(m)
			case *message.FileMessage:
				c.processFileMessage(m)
			default:
				fmt.Println("Unknown message type")
			}

			// 更新缓存，移除已处理的部分
			dataBuffer = remainingBuffer
		}
	}

	global.LOG.Info("Connection closed: %s", conn.RemoteAddr().String())
}

func (c *Center) processMessage(msg *message.Message) {
	switch msg.Type {
	case message.Heartbeat: // 收到心跳
		// TODO: 维护在线状态
	case message.CmdMessage: // 收到Cmd 类型的回复
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
		global.LOG.Info("Processing action message: %s", msg.Data)
		//获取响应通道
		c.mu.Lock()
		responseCh, exists := c.responseChMap[msg.MsgID]
		if exists {
			responseCh <- msg.Data // msg.Data 是 model.Action
			close(responseCh)
			delete(c.responseChMap, msg.MsgID)
		}
		c.mu.Unlock()
	default: // 不支持的消息
		global.LOG.Error("Unknown message type: %s", msg.Type)
	}
}

func (c *Center) processFileMessage(msg *message.FileMessage) {
	switch msg.Type {
	case message.Upload: //上传回复
		global.LOG.Info("Upload: %s, %d, %d, %d", msg.FileName, msg.Status, msg.Offset, msg.ChunkSize)
		// 获取响应通道
		c.mu.Lock()
		responseCh, exists := c.fileResponseChMap[msg.MsgID]
		if exists {
			responseCh <- msg
			//最后一次传输，关闭通道
			if msg.Status == message.FileDone {
				close(responseCh)
				delete(c.fileResponseChMap, msg.MsgID)
			}
		}
		c.mu.Unlock()
	case message.Download: //下载回复
		global.LOG.Info("Download: %s, %d, %d, %d", msg.FileName, msg.Status, msg.Offset, msg.ChunkSize)
		// 获取响应通道
		c.mu.Lock()
		responseCh, exists := c.fileResponseChMap[msg.MsgID]
		if exists {
			responseCh <- msg
			//最后一次传输，关闭通道
			if msg.Status == message.FileDone {
				close(responseCh)
				delete(c.fileResponseChMap, msg.MsgID)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Center) sendHeartbeat() {
	config := CONFMAN.GetConfig()
	c.mu.Lock()
	defer c.mu.Unlock()

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
}

func (c *Center) UploadFile(hostID uint, path string, file *multipart.FileHeader) error {

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(hostID))
	if err != nil || host.ID == 0 {
		return errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(host)
	if err != nil {
		return errors.WithMessage(constant.ErrAgent, err.Error())
	}

	// 打开文件
	srcFile, err := file.Open()
	if err != nil {
		return errors.WithMessage(errors.New(constant.ErrFileOpen), err.Error())
	}
	defer srcFile.Close()

	// 获取文件大小
	fileSize := file.Size

	// 创建等待响应的通道
	responseCh := make(chan *message.FileMessage)

	// 生成消息ID
	msgID := utils.GenerateMsgId()

	// 将通道和msgID映射存储在map中
	c.mu.Lock()
	c.fileResponseChMap[msgID] = responseCh
	c.mu.Unlock()

	// 分块读取文件并发送
	const bufferSize = 256 * 1024 // 256KB 块大小
	buffer := make([]byte, bufferSize)
	var offset int64 = 0
	for {
		// 读取文件块
		n, err := srcFile.Read(buffer)
		if err != nil && err != io.EOF {
			return errors.WithMessage(errors.New(constant.ErrFileRead), err.Error())
		}
		if n == 0 {
			break // 文件读取完毕
		}

		// 构造要发送的消息
		msg, err := message.CreateFileMessage(
			msgID,
			message.Upload,
			0,
			path,
			file.Filename,
			fileSize,
			offset,
			n,
			buffer[:n],
		)
		if err != nil {
			return errors.WithMessage(constant.ErrInternalServer, err.Error())
		}

		// 并发发送消息
		go func() {
			err := message.SendFileMessage(conn, msg)
			if err != nil {
				global.LOG.Error("Failed to send file chunk: %s %d %d, %v", msg.FileName, msg.Offset, msg.ChunkSize, err)
				// 如果发送失败，写入空响应
				msg.Status = message.FileErr
				responseCh <- msg
			}
		}()

		// 等待 agent 响应
		select {
		case response := <-responseCh:
			if response.Status == message.FileErr {
				return errors.New("failed to upload file chunk")
			}
			// 继续下一块
		case <-time.After(10 * time.Second): // 设置超时时间
			c.mu.Lock()
			delete(c.fileResponseChMap, msgID)
			c.mu.Unlock()
			return fmt.Errorf("timeout waiting for response from agent")
		}

		// 更新偏移量，准备发送下一块
		offset += int64(n)
	}

	return nil
}

func (c *Center) DownloadFile(ctx *gin.Context, hostID uint, path string) error {
	dir := filepath.Dir(path)
	// 解析出文件名
	fileName := filepath.Base(path)

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(hostID))
	if err != nil || host.ID == 0 {
		return errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(host)
	if err != nil {
		return errors.WithMessage(constant.ErrAgent, err.Error())
	}

	// 设置 HTTP 响应头，确保下载的是文件
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	// 根据文件扩展名获取 MIME 类型
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		// 如果无法确定 MIME 类型，使用默认的二进制流类型
		mimeType = "application/octet-stream"
	}
	ctx.Header("Content-Type", mimeType)

	// 创建等待响应的通道
	responseCh := make(chan *message.FileMessage)

	// 生成消息ID
	msgID := utils.GenerateMsgId()

	// 将通道和msgID映射存储在map中
	c.mu.Lock()
	c.fileResponseChMap[msgID] = responseCh
	c.mu.Unlock()

	var finished bool = false
	var offset int64 = 0
	for {
		// 完成时，跳出
		if finished {
			break
		}

		go func() {
			// 构造要发送的消息
			msg, err := message.CreateFileMessage(
				msgID,
				message.Download,
				0,
				dir,
				fileName,
				0,
				offset,
				256*1024,
				nil,
			)
			if err != nil {
				global.LOG.Error("Failed to create file message: %s %d %d, %v", msg.FileName, msg.Offset, msg.ChunkSize, err)
				// 如果发送失败，写入空响应
				msg.Status = message.FileErr
				responseCh <- msg
				return
			}

			err = message.SendFileMessage(conn, msg)
			if err != nil {
				global.LOG.Error("Failed to send file chunk: %s %d %d, %v", msg.FileName, msg.Offset, msg.ChunkSize, err)
				// 如果发送失败，写入空响应
				msg.Status = message.FileErr
				responseCh <- msg
			}
		}()

		select {
		case response := <-responseCh:
			if response.Status == message.FileErr {
				return errors.New("failed to download file chunk")
			}
			// 写入response
			if _, err := ctx.Writer.Write(response.Chunk[:response.ChunkSize]); err != nil {
				return errors.WithMessage(constant.ErrInternalServer, err.Error())
			}
			// 如果已经完成
			if response.Status == message.FileDone {
				finished = true
			} else {
				// 继续请求下一块
				offset = response.Offset + int64(response.ChunkSize)
			}
		case <-time.After(10 * time.Second): // 设置超时时间
			c.mu.Lock()
			delete(c.fileResponseChMap, msgID)
			c.mu.Unlock()
			return fmt.Errorf("timeout waiting for response from agent")
		}
	}
	return nil
}

func (c *Center) ExecuteAction(req core.HostAction) (*core.Action, error) {
	config := CONFMAN.GetConfig()

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil || host.ID == 0 {
		return nil, errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(host)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(req.Action)
	if err != nil {
		return nil, err
	}

	// 创建一个等待通道
	responseCh := make(chan string)

	// 创建消息
	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		string(data),
		config.SecretKey,
		utils.GenerateNonce(16),
		message.ActionMessage,
	)
	if err != nil {
		return nil, err
	}

	// 将通道和msgID映射存储在map中
	c.mu.Lock()
	c.responseChMap[msgID] = responseCh
	c.mu.Unlock()

	go func() {
		err = message.SendMessage(conn, msg)
		if err != nil {
			global.LOG.Error("Failed to send action message: %v", err)
			responseCh <- ""
		}
	}()

	// 等待响应
	select {
	case response := <-responseCh:
		var action core.Action
		if err := json.Unmarshal([]byte(response), &action); err != nil {
			return nil, err
		}
		return &action, nil
	case <-time.After(10 * time.Second): // 设置一个超时时间
		c.mu.Lock()
		delete(c.responseChMap, msgID)
		c.mu.Unlock()
		return nil, fmt.Errorf("timeout waiting for response from agent")
	}
}

func (c *Center) ExecuteCommand(req core.Command) (string, error) {

	config := CONFMAN.GetConfig()

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil {
		return "", errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(host)
	if err != nil {
		return "", err
	}

	// 创建一个等待通道
	responseCh := make(chan string)

	// 创建消息
	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		req.Command,
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

func (c *Center) getAgentConn(host model.Host) (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	agentID := fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort)
	conn, exists := c.agentConns[agentID]
	if !exists || conn == nil {
		return nil, errors.WithMessage(constant.ErrAgent, "not connected")
	}
	return conn, nil
}

func (c *Center) ExecuteCommandGroup(req core.CommandGroup) ([]string, error) {
	config := CONFMAN.GetConfig()

	if len(req.Commands) < 1 {
		return []string{}, constant.ErrInvalidParams
	}

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil || host.ID == 0 {
		return []string{}, errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(host)
	if err != nil {
		return []string{}, err
	}

	// 创建一个等待通道
	responseCh := make(chan string)

	// 创建消息
	var data string
	if len(req.Commands) > 1 {
		data = strings.Join(req.Commands, message.Separator)
	} else {
		data = req.Commands[0]
	}
	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		data,
		config.SecretKey,
		utils.GenerateNonce(16),
		message.CmdMessage,
	)
	if err != nil {
		return []string{}, err
	}

	// 将通道和msgID映射存储在map中
	c.mu.Lock()
	c.responseChMap[msgID] = responseCh
	c.mu.Unlock()

	go func() {
		global.LOG.Info("send msg data: %s", msg.Data)
		err = message.SendMessage(conn, msg)
		if err != nil {
			global.LOG.Error("Failed to send command message: %v", err)
			responseCh <- ""
		}
	}()

	// 等待响应
	select {
	case response := <-responseCh:
		global.LOG.Info("recv msg data: %s", response)
		var results []string
		if strings.Contains(response, message.Separator) {
			results = strings.Split(response, message.Separator)
		} else {
			results = append(results, response)
		}
		return results, nil
	case <-time.After(10 * time.Second): // 设置一个超时时间
		c.mu.Lock()
		delete(c.responseChMap, msgID)
		c.mu.Unlock()
		return []string{}, fmt.Errorf("timeout waiting for response from agent")
	}
}

func (c *Center) TestAgent(id uint, req core.TestAgent) error {
	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, _ := c.getAgentConn(host)
	if conn != nil {
		return nil
	} else {
		resultCh := make(chan error, 1)
		go c.connectToAgent(&host, resultCh)
		// handle the result if needed
		err := <-resultCh
		if err != nil {
			global.LOG.Error("Failed to connect to agent %s: %v", host.Addr, err)
			return err
		}
	}

	return nil
}
