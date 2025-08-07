package conn

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/sensdata/idb/center/core/plugin"
	"github.com/sensdata/idb/center/core/plugin/shared"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/logstream/pkg/reader/adapters"
	"github.com/sensdata/idb/core/message"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type Center struct {
	agentConns        map[string]net.Conn // 存储Agent端连接的映射
	done              chan struct{}
	mu                sync.Mutex             // 保护agentConns的互斥锁
	responseChMap     map[string]chan string // 用于接收命令执行结果的动态通道
	fileResponseChMap map[string]chan *message.FileMessage
	awsMap            map[string]*AgentWebSocketSession
	sessionTokenMap   map[string]string // 缓存session是否被占用
}

type ICenter interface {
	Start() error
	Stop() error
	ActivateHost(host *model.Host) error
	ExecuteCommand(req core.Command) (string, error)
	ExecuteCommandGroup(req core.CommandGroup) ([]string, error)
	ExecuteAction(req core.HostAction) (*core.Action, error)
	UploadFile(hostID uint, path string, file *multipart.FileHeader) error
	DownloadFile(ctx *gin.Context, hostID uint, path string) error
	GetAgentConn(host *model.Host) (*net.Conn, error)
	IsAgentConnected(host model.Host) bool
	RegisterAgentSession(aws *AgentWebSocketSession)
	UnregisterAgentSession(session string)
	RegisterSessionToken(session string, token string)
	UnregisterSessionToken(session string)
	GetSessionToken(session string) (string, bool)
	TestAgent(host model.Host, req core.TestAgent) error
	ReleaseAgentConn(host model.Host) error
}

func NewCenter() ICenter {
	return &Center{
		agentConns:        make(map[string]net.Conn),
		done:              make(chan struct{}),
		responseChMap:     make(map[string]chan string),
		fileResponseChMap: make(map[string]chan *message.FileMessage),
		awsMap:            make(map[string]*AgentWebSocketSession),
		sessionTokenMap:   make(map[string]string),
	}
}

func (c *Center) Start() error {

	global.LOG.Info("Center Starting")

	// 启动 Unix 域套接字监听器
	go c.listenToUnix()

	// 保障连接和心跳
	go c.ensureConnections()

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

	return nil
}

func (c *Center) listenToUnix() {
	global.LOG.Info("Start listening to unix")

	// 检查sock文件
	sockFile := filepath.Join(constant.CenterRunDir, constant.CenterSock)

	// 如果sock文件存在，尝试删除
	if _, err := os.Stat(sockFile); err == nil {
		global.LOG.Info("Removing existing sock file")
		if err := os.Remove(sockFile); err != nil {
			global.LOG.Error("Failed to remove existing sock file: %v", err)
			return
		}
	}

	listener, err := net.Listen("unix", sockFile)
	if err != nil {
		global.LOG.Error("Failed to create listener: %v", err)
		return
	}
	global.LOG.Info("Unix listener created on sock file: %s", sockFile)

	defer func() {
		global.LOG.Info("Unix listener closing")
		listener.Close()

		// 只在服务退出时清理 socket 文件
		global.LOG.Info("Removing existing sock file")
		sockFile := filepath.Join(constant.CenterRunDir, constant.CenterSock)
		if err := os.Remove(sockFile); err != nil {
			global.LOG.Error("Failed to remove existing sock file: %v", err)
		}
	}()

	for {
		select {
		case <-c.done:
			global.LOG.Info("Stop accepting new unix connections")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				global.LOG.Error("failed to accept unix connection: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 处理连接
			go c.handleUnixConnection(conn)
		}
	}
}

func (c *Center) handleUnixConnection(conn net.Conn) {
	defer func() {
		// 断连
		global.LOG.Info("Close unix conn")
		conn.Close()

		if r := recover(); r != nil {
			global.LOG.Error("[Panic] in handleUnixConnection: %v", r)
		}
	}()

	buf := make([]byte, 1024)
	for {
		select {
		case <-c.done:
			global.LOG.Info("Stop handle unix connection")
			return
		default:
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
				conn.Write([]byte(fmt.Sprintf("iDB Center (pid %d) is running...", os.Getpid())))
			case "config":
				switch len(parts) {
				case 1:
					// 输出当前的配置信息
					settings, err := c.GetSettingsString("")
					if err != nil {
						conn.Write([]byte(fmt.Sprintf("Failed to get config: %v", err)))
					} else {
						conn.Write([]byte(fmt.Sprintf("%v", settings)))
					}
				case 2:
					// 输出当前的指定key配置信息
					key := parts[1]
					value, err := c.GetSettingsString(key)
					if err != nil {
						conn.Write([]byte(fmt.Sprintf("Failed to get %s: %v", key, err)))
					} else {
						conn.Write([]byte(fmt.Sprintf("%v", value)))
					}
				case 3:
					// 修改指定key的配置
					key := parts[1]
					value := parts[2]
					err := c.UpdateSetting(key, value)
					if err != nil {
						conn.Write([]byte(fmt.Sprintf("Failed to set config %s: %v", key, err)))
					} else {
						value, err := c.GetSettingsString(key)
						if err != nil {
							conn.Write([]byte(fmt.Sprintf("Failed to get %s: %v", key, err)))
						} else {
							conn.Write([]byte(fmt.Sprintf("%v", value)))
						}
					}
				default:
					conn.Write([]byte("Unknown config command format"))
				}
			case "update":
				err := c.Upgrade()
				if err != nil {
					conn.Write([]byte(fmt.Sprintf("Failed to update: %v", err)))
				} else {
					conn.Write([]byte("Upgrade success"))
				}
			case "rst-pass":
				newPass, err := c.ResetAdminPassword()
				if err != nil {
					conn.Write([]byte(fmt.Sprintf("Failed to reset password: %v", err)))
				} else {
					conn.Write([]byte(fmt.Sprintf("Password reset, please remember your new password: %s", newPass)))
				}
			default:
				conn.Write([]byte("Unknown command"))
			}
		}
	}
}

func (c *Center) ensureConnections() {
	global.LOG.Info("Ensure connections")

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			global.LOG.Info("Stop ensure connections")
			return
		case <-ticker.C:
			//获取所有的host
			hosts, err := HostRepo.GetList()
			if err != nil {
				global.LOG.Error("Failed to get host list: %v", err)
				continue
			}

			// 检查连接
			for _, host := range hosts {
				global.LOG.Info("checkConn for host %d - %s", host.ID, host.Addr)
				// 查找agent conn
				conn, _ := c.getAgentConn(&host)
				if conn == nil {
					// 连接
					resultCh := make(chan error, 1)
					go c.connectToAgent(&host, resultCh)
				} else {
					// 找到conn的，发心跳
					go c.sendHeartbeat(&host, conn)
				}
			}
		}
	}
}

func getAuthPlugin() (shared.Auth, error) {
	// 获取插件客户端
	plugin, err := plugin.PLUGINSERVER.GetPlugin("auth")
	if err != nil {
		return nil, err
	}
	// 类型断言为 gRPC client
	client, ok := plugin.Stub.(shared.Auth)
	if !ok {
		return nil, errors.New("invalid plugin client")
	}
	return client, nil
}

func (c *Center) ActivateHost(host *model.Host) error {
	return c.activateHost(host)
}

func (c *Center) activateHost(host *model.Host) error {
	global.LOG.Info("activate host %d - %s begin", host.ID, host.Addr)
	// 获取指纹
	var fingerprint core.Fingerprint
	actionRequest := core.HostAction{
		HostID: host.ID,
		Action: core.Action{
			Action: core.Host_Fingerprint,
			Data:   "",
		},
	}
	actionResponse, err := c.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Host_Fingerprint %v", err)
		return fmt.Errorf("Failed to send action Host_Fingerprint %w", err)
	}
	if !actionResponse.Result {
		global.LOG.Error("action Host_Fingerprint failed")
		return errors.New("action Host_Fingerprint failed")
	}
	err = utils.FromJSONString(actionResponse.Data, &fingerprint)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to fingerprint: %v", err)
		return fmt.Errorf("Error unmarshaling data to fingerprint: %w", err)
	}

	// 使用auth插件验证
	auth, err := getAuthPlugin()
	if err != nil {
		global.LOG.Error("Failed to get auth plugin: %v", err)
		return fmt.Errorf("Failed to get auth plugin: %w", err)
	}
	verifyResp, err := auth.VerifyFingerprint(core.VerifyRequest{
		Fingerprint: fingerprint.Fingerprint,
		IP:          fingerprint.IP,
		MAC:         fingerprint.MAC,
	})
	if err != nil {
		global.LOG.Error("Failed to verify fingerprint: %v", err)
		return fmt.Errorf("Failed to verify fingerprint: %w", err)
	}

	// 将验证结果发给agent
	fingerprint.VerifyResult = verifyResp.Result
	fingerprint.VerifyTime = time.Unix(verifyResp.VerifyTime, 0)
	fingerprint.ExpireTime = time.Unix(verifyResp.ExpireTime, 0)
	data, err := utils.ToJSONString(fingerprint)
	if err != nil {
		global.LOG.Error("Error marshaling fingerprint to JSON: %v", err)
		return fmt.Errorf("Error marshaling fingerprint to JSON: %w", err)
	}
	verifyRequest := core.HostAction{
		HostID: host.ID,
		Action: core.Action{
			Action: core.Host_Fingerprint_Verify,
			Data:   data,
		},
	}
	verifyResponse, err := c.ExecuteAction(verifyRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Host_Fingerprint_Verify %v", err)
		return fmt.Errorf("Failed to send action Host_Fingerprint_Verify %w", err)
	}
	if !verifyResponse.Result {
		global.LOG.Error("action Host_Fingerprint_Verify failed")
		return fmt.Errorf("action Host_Fingerprint_Verify failed")
	}
	global.LOG.Info("activate host %d - %s success", host.ID, host.Addr)
	return nil
}

func (c *Center) sendHeartbeat(host *model.Host, conn *net.Conn) error {
	agentID := fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort)

	heartbeatMsg, err := message.CreateMessage(
		utils.GenerateMsgId(),
		"Heartbeat",
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
		message.Heartbeat,
	)
	if err != nil {
		global.LOG.Error("Error creating heartbeat message: %v", err)
		return err
	}

	err = message.SendMessage(*conn, heartbeatMsg)
	if err != nil {
		global.LOG.Error("Failed to send heartbeat message to %s: %v", agentID, err)
		(*conn).Close()
		global.LOG.Info("close conn %s for heartbeat", agentID)
		c.mu.Lock()
		delete(c.agentConns, agentID)
		c.mu.Unlock()
		global.LOG.Info("delete conn %s for heartbeat", agentID)
		return err
	} else {
		global.LOG.Info("Heartbeat sent to %s", agentID)
	}
	return nil
}

func (c *Center) connectToAgent(host *model.Host, resultCh chan<- error) {
	global.LOG.Info("try connect to agent %s:%d", host.AgentAddr, host.AgentPort)
	// conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort))

	// 创建证书池并添加自签名证书
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(global.CaCertPem)

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,       // 使用自定义的证书池
		MinVersion:         tls.VersionTLS13, // 设置最小 TLS 版本
		InsecureSkipVerify: true,
	}

	// // 建立 TLS 连接
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort), tlsConfig)
	if err != nil {
		global.LOG.Error("Failed to connect to Agent: %v", err)
		resultCh <- err
		return
	}

	// 记录连接
	agentID := conn.RemoteAddr().String()
	c.mu.Lock()
	c.agentConns[agentID] = conn
	c.mu.Unlock()

	global.LOG.Info("Successfully connected to Agent %s", agentID)
	resultCh <- nil

	// 处理连接
	go c.handleConnection(host, conn)

	// 如果是default设备, 检查是否已激活
	if host.IsDefault {
		go c.activateHost(host)
	}
}

func (c *Center) handleConnection(host *model.Host, conn net.Conn) {
	agentID := conn.RemoteAddr().String()

	defer func() {
		// 在defer中关闭连接
		global.LOG.Info("Close agent conn %s", agentID)
		conn.Close()

		if r := recover(); r != nil {
			global.LOG.Error("[Panic] in handleUnixConnection: %v", r)
		}
	}()

	// 缓存区：用来缓存从 conn.Read 读取的数据
	dataBuffer := make([]byte, 0)
	tmpBuffer := make([]byte, 1024)
	for {
		select {
		case <-c.done:
			global.LOG.Info("Stop handle connection %s", agentID)
			return
		default:
			// 读取数据
			n, err := conn.Read(tmpBuffer)
			if err != nil {
				if err != io.EOF {
					global.LOG.Error("Error read from conn: %v", err)
				}
				return
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
				msg, err := message.DecodeMessage(msgType, msgData, host.AgentKey)
				if err != nil {
					global.LOG.Error("Error decode message: %v", err)
				}
				switch m := msg.(type) {
				case *message.Message:
					c.processMessage(host, m)
				case *message.FileMessage:
					c.processFileMessage(m)
				case *message.SessionMessage:
					c.processSessionMessage(m)
				case *message.LogStreamMessage:
					c.processLogStreamMessage(m)
				default:
					fmt.Println("Unknown message type")
				}

				// 更新缓存，移除已处理的部分
				dataBuffer = remainingBuffer
			}
		}
	}
}

func (c *Center) checkAgentUpdate(host *model.Host, agentVersion string) {
	global.LOG.Info("Check agent update for host %s", host.AgentAddr)
	// latestVersion 通过读取文件 /var/lib/idb/agent/idb-agent.version 来获得
	latestPath := filepath.Join(constant.CenterAgentDir, constant.AgentLatest)
	var latestVersion string
	version, err := os.ReadFile(latestPath)
	if err != nil {
		global.LOG.Error("Failed to read latest version: %v", err)
		latestVersion = ""
	} else {
		latestVersion = strings.TrimSpace(string(version))
	}
	if agentVersion == latestVersion {
		global.LOG.Info("Agent is up to date")
		return
	}
	SSH.InstallAgent(*host, "", true)
}

func (c *Center) removeAgent(host *model.Host) {
	global.LOG.Info("Remove agent %s", host.AgentAddr)
	SSH.UninstallAgent(*host, "")
}

func (c *Center) processMessage(host *model.Host, msg *message.Message) {
	switch msg.Type {
	case message.Heartbeat: // 收到心跳
		// 处理心跳消息
		global.LOG.Info("Received heartbeat from agent: %s", msg.Data)
		// 写入agent版本号
		if err := HostRepo.Update(host.ID, map[string]interface{}{"agent_version": msg.Version}); err != nil {
			global.LOG.Error("Failed to update agent version: %v", err)
		}
		// 看是不是需要检查升级
		switch msg.Data {
		case "Update":
			// 检查升级
			go c.checkAgentUpdate(host, msg.Version)
		case "Remove":
			// 移除agent
			go c.removeAgent(host)
		}

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

func (c *Center) processSessionMessage(msg *message.SessionMessage) {
	global.LOG.Info("Process session message: %v", msg)
	c.mu.Lock()
	defer c.mu.Unlock()

	switch msg.Type {
	case message.WsMessageStart:
		// start的时候，通过msgID找aws
		aws, exists := c.awsMap[msg.MsgID]
		if exists {
			// 替换成session作为key
			aws.Session = msg.Data.Session
			aws.Name = msg.Data.Data
			c.awsMap[msg.Data.Session] = aws
			// 删除原来的msgID对应的记录
			delete(c.awsMap, msg.MsgID)
			global.LOG.Info("replace aws msgID %s with session %s", msg.MsgID, msg.Data.Session)

			aws.SessionMessageChan <- msg
		} else {
			global.LOG.Info("no response session")
		}
		// 找 session - token
		token, exists := c.sessionTokenMap[msg.MsgID]
		if exists {
			// 替换成session作为key
			c.sessionTokenMap[msg.Data.Session] = token
			// 删除原 msgID 对应记录
			delete(c.sessionTokenMap, msg.MsgID)
			global.LOG.Info("replace token msgID %s with session %s", msg.MsgID, msg.Data.Session)
		} else {
			global.LOG.Info("no token - session")
		}
	case message.WsMessageAttach:
		// attach的时候，通过msgID找aws
		aws, exists := c.awsMap[msg.MsgID]
		if exists {
			// 替换成session作为key
			aws.Session = msg.Data.Session
			aws.Name = msg.Data.Data
			c.awsMap[msg.Data.Session] = aws
			// 删除原来的msgID对应的记录
			delete(c.awsMap, msg.MsgID)
			global.LOG.Info("replace msgID %s with session %s", msg.MsgID, msg.Data.Session)

			aws.SessionMessageChan <- msg
		} else {
			global.LOG.Info("no response session")
		}
		// 找 session - token
		token, exists := c.sessionTokenMap[msg.MsgID]
		if exists {
			// 替换成session作为key
			c.sessionTokenMap[msg.Data.Session] = token
			// 删除原 msgID 对应记录
			delete(c.sessionTokenMap, msg.MsgID)
			global.LOG.Info("replace token msgID %s with session %s", msg.MsgID, msg.Data.Session)
		} else {
			global.LOG.Info("no token - session")
		}
	case message.WsMessageCmd:
		// command的时候，通过session找aws
		aws, exists := c.awsMap[msg.Data.Session]
		if exists {
			aws.SessionMessageChan <- msg
		} else {
			global.LOG.Info("no response session")
		}
	default: // 不支持的消息
		global.LOG.Error("Unknown sesssion message type: %s", msg.Type)
	}
}

func (c *Center) processLogStreamMessage(msg *message.LogStreamMessage) {
	if msg == nil {
		global.LOG.Error("received nil log stream message")
		return
	}

	ls := global.LogStream

	switch msg.Type {
	case message.LogStreamStart:
		global.LOG.Info("log stream started for file: %s", msg.LogPath)
		return

	case message.LogStreamStop:
		global.LOG.Info("log stream stopped for file: %s", msg.LogPath)
		return

	case message.LogStreamData, message.LogStreamError:
		// 获取已存在的 reader
		reader, err := ls.GetExistingReader(msg.TaskID)
		if err != nil {
			global.LOG.Error("get existing reader failed for task %s: %v", msg.TaskID, err)
			return
		}

		// 类型转换
		remoteReader, ok := reader.(*adapters.RemoteReader)
		if !ok {
			global.LOG.Error("invalid reader type for task %s", msg.TaskID)
			return
		}

		// 根据消息类型准备内容
		var content []byte
		if msg.Type == message.LogStreamData {
			content = []byte(msg.Content)
		} else {
			content = []byte(fmt.Sprintf("Error: %s", msg.Error))
		}

		// 发送日志内容
		if err := remoteReader.SendLog(content); err != nil {
			global.LOG.Error("send log failed for task %s: %v", msg.TaskID, err)
			return
		}

	default:
		global.LOG.Error("unknown log stream message type: %s", msg.Type)
	}
}

func (c *Center) UploadFile(hostID uint, path string, file *multipart.FileHeader) error {

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(hostID))
	if err != nil || host.ID == 0 {
		return errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(&host)
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
			err := message.SendFileMessage(*conn, msg)
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
	conn, err := c.getAgentConn(&host)
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

			err = message.SendFileMessage(*conn, msg)
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

func (c *Center) GetAgentConn(host *model.Host) (*net.Conn, error) {
	return c.getAgentConn(host)
}

func (c *Center) RegisterAgentSession(aws *AgentWebSocketSession) {
	c.mu.Lock()
	defer c.mu.Unlock()

	global.LOG.Info("Session %s registered", aws.Session)
	c.awsMap[aws.Session] = aws
}

func (c *Center) UnregisterAgentSession(session string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	global.LOG.Info("Session %s unregistered", session)
	delete(c.awsMap, session)
}

func (c *Center) RegisterSessionToken(session string, token string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	global.LOG.Info("Session %s token registered", session)
	c.sessionTokenMap[session] = token
}

func (c *Center) UnregisterSessionToken(session string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	global.LOG.Info("Session %s token unregistered", session)
	delete(c.sessionTokenMap, session)
}

func (c *Center) GetSessionToken(session string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	token, exists := c.sessionTokenMap[session]
	return token, exists
}

func (c *Center) ExecuteAction(req core.HostAction) (*core.Action, error) {

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil || host.ID == 0 {
		return nil, errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(&host)
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
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
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
		err = message.SendMessage(*conn, msg)
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

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil {
		return "", errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(&host)
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
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
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
		err = message.SendMessage(*conn, msg)
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

func (c *Center) IsAgentConnected(host model.Host) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	agentID := fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort)
	conn, exists := c.agentConns[agentID]
	return exists && conn != nil
}

func (c *Center) getAgentConn(host *model.Host) (*net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	agentID := fmt.Sprintf("%s:%d", host.AgentAddr, host.AgentPort)
	conn, exists := c.agentConns[agentID]
	if !exists || conn == nil {
		return nil, errors.WithMessage(constant.ErrAgent, "not connected")
	}
	return &conn, nil
}

func (c *Center) ExecuteCommandGroup(req core.CommandGroup) ([]string, error) {

	if len(req.Commands) < 1 {
		return []string{}, constant.ErrInvalidParams
	}

	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil || host.ID == 0 {
		return []string{}, errors.WithMessage(constant.ErrHost, err.Error())
	}

	// 查找agent conn
	conn, err := c.getAgentConn(&host)
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
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
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
		err = message.SendMessage(*conn, msg)
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

func (c *Center) TestAgent(host model.Host, req core.TestAgent) error {
	// 查找agent conn
	conn, _ := c.getAgentConn(&host)
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

func (c *Center) ReleaseAgentConn(host model.Host) error {
	// 查找agent conn
	conn, _ := c.getAgentConn(&host)
	if conn != nil {
		(*conn).Close()
	}

	return nil
}

func (c *Center) GetSettingsString(item string) (string, error) {
	settings, err := c.getServerSettings()
	if err != nil {
		return "", err
	}
	var result strings.Builder

	if item == "" {
		result.WriteString(fmt.Sprintf("bind_ip         : %s\n", settings.BindIP))
		result.WriteString(fmt.Sprintf("bind_port       : %d\n", settings.BindPort))
		result.WriteString(fmt.Sprintf("bind_domain     : %s\n", settings.BindDomain))

		protocal := "http"
		if settings.Https == "yes" {
			protocal = "https"
		}
		result.WriteString(fmt.Sprintf("protocal        : %s\n", protocal))
		result.WriteString(fmt.Sprintf("https_cert_type : %s\n", settings.HttpsCertType))
		result.WriteString(fmt.Sprintf("https_cert_path : %s\n", settings.HttpsCertPath))
		result.WriteString(fmt.Sprintf("https_key_path  : %s\n", settings.HttpsKeyPath))
	} else {
		switch item {
		case "bind_ip":
			result.WriteString(fmt.Sprintf("bind_ip         : %s\n", settings.BindIP))
		case "bind_port":
			result.WriteString(fmt.Sprintf("bind_port       : %d\n", settings.BindPort))
		case "bind_domain":
			result.WriteString(fmt.Sprintf("bind_domain     : %s\n", settings.BindDomain))
		case "protocal":
			protocal := "http"
			if settings.Https == "yes" {
				protocal = "https"
			}
			result.WriteString(fmt.Sprintf("protocal        : %s\n", protocal))
		case "https_cert_type":
			result.WriteString(fmt.Sprintf("https_cert_type : %s\n", settings.HttpsCertType))
		case "https_cert_path":
			result.WriteString(fmt.Sprintf("https_cert_path : %s\n", settings.HttpsCertPath))
		case "https_key_path":
			result.WriteString(fmt.Sprintf("https_key_path  : %s\n", settings.HttpsKeyPath))
		}
	}

	return result.String(), nil
}

func (c *Center) getServerSettings() (*core.SettingInfo, error) {
	settingRepo := repo.NewSettingsRepo()
	bindIP, err := settingRepo.Get(settingRepo.WithByKey("BindIP"))
	if err != nil {
		return nil, err
	}
	bindPort, err := settingRepo.Get(settingRepo.WithByKey("BindPort"))
	if err != nil {
		return nil, err
	}
	bindPortValue, err := strconv.Atoi(bindPort.Value)
	if err != nil {
		return nil, err
	}
	bindDomain, err := settingRepo.Get(settingRepo.WithByKey("BindDomain"))
	if err != nil {
		return nil, err
	}
	https, err := settingRepo.Get(settingRepo.WithByKey("Https"))
	if err != nil {
		return nil, err
	}
	httpsCertType, err := settingRepo.Get(settingRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return nil, err
	}
	httpsCertPath, err := settingRepo.Get(settingRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return nil, err
	}
	httpsKeyPath, err := settingRepo.Get(settingRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return nil, err
	}

	return &core.SettingInfo{
		BindIP:        bindIP.Value,
		BindPort:      bindPortValue,
		BindDomain:    bindDomain.Value,
		Https:         https.Value,
		HttpsCertType: httpsCertType.Value,
		HttpsCertPath: httpsCertPath.Value,
		HttpsKeyPath:  httpsKeyPath.Value,
	}, nil
}

func (c *Center) UpdateSetting(key string, value string) error {
	return c.updateServerSetting(key, value)
}

func (c *Center) updateServerSetting(key string, value string) error {

	// 开始事务
	var err error
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			global.LOG.Error("Transaction failed: %v  - rollback", r)
		} else if err != nil {
			tx.Rollback() // 如果发生错误，回滚事务
			global.LOG.Error("Error Happend - rollback")
		}
	}()

	switch key {
	case "bind_ip":
		err = c.updateBindIP(value)
		if err != nil {
			global.LOG.Error("update bind ip failed: %v", err)
			return err
		}
	case "bind_port":
		port, err := strconv.Atoi(value)
		if err != nil {
			global.LOG.Error("update bind port failed: %v", err)
			return err
		}
		err = c.updateBindPort(port)
		if err != nil {
			global.LOG.Error("update bind port failed: %v", err)
			return err
		}
	case "bind_domain":
		err = c.updateBindDomain(value)
		if err != nil {
			global.LOG.Error("update bind domain failed: %v", err)
			return err
		}
	case "protocal":
		https := "no"
		if value == "https" {
			https = "yes"
		}
		err = c.updateHttps(https)
		if err != nil {
			global.LOG.Error("update protocal failed: %v", err)
			return err
		}
	case "https_cert_type":
		err = c.updateHttpsCertType(value)
		if err != nil {
			global.LOG.Error("update https cert type failed: %v", err)
			return err
		}
	case "https_cert_path":
		err = c.updateHttpsCertPath(value)
		if err != nil {
			global.LOG.Error("update https cert path failed: %v", err)
			return err
		}
	case "https_key_path":
		err = c.updateHttpsKeyPath(value)
		if err != nil {
			global.LOG.Error("update https key path failed: %v", err)
			return err
		}
	default:
		return errors.New("invalid key")
	}

	// 提交事务
	tx.Commit()

	go func() {
		time.Sleep(2 * time.Second)
		// 发送 SIGTERM 信号给主进程，触发容器重启
		if err := syscall.Kill(1, syscall.SIGTERM); err != nil {
			global.LOG.Error("Failed to send termination signal: %v", err)
		}
	}()

	return nil
}

func (c *Center) updateBindIP(newIP string) error {
	if len(newIP) == 0 {
		return errors.New("invalid bind ip")
	}

	settingsRepo := repo.NewSettingsRepo()
	oldIP, err := settingsRepo.Get(settingsRepo.WithByKey("BindIP"))
	if err != nil {
		return err
	}
	if newIP == oldIP.Value {
		return nil
	}

	return settingsRepo.Update("BindIP", newIP)
}

func (c *Center) updateBindPort(newPort int) error {
	if newPort <= 0 || newPort > 65535 {
		return errors.New("server port must between 1 - 65535")
	}
	settingsRepo := repo.NewSettingsRepo()
	oldPort, err := settingsRepo.Get(settingsRepo.WithByKey("BindPort"))
	if err != nil {
		return err
	}
	newPortStr := strconv.Itoa(newPort)
	if newPortStr == oldPort.Value {
		return nil
	}

	if common.ScanPort(newPort) {
		return errors.New(constant.ErrPortInUsed)
	}

	// TODO: 处理port的更换（调用nftables）

	return settingsRepo.Update("BindPort", newPortStr)
}

func (c *Center) updateBindDomain(newDomain string) error {
	domain := newDomain
	if newDomain == "empty" {
		domain = ""
	}
	settingsRepo := repo.NewSettingsRepo()
	oldDomain, err := settingsRepo.Get(settingsRepo.WithByKey("BindDomain"))
	if err != nil {
		return err
	}
	if domain == oldDomain.Value {
		return nil
	}
	return settingsRepo.Update("BindDomain", domain)
}

func (c *Center) updateHttps(https string) error {
	if len(https) == 0 {
		return nil
	}
	settingsRepo := repo.NewSettingsRepo()
	oldHttps, err := settingsRepo.Get(settingsRepo.WithByKey("Https"))
	if err != nil {
		return err
	}
	if https == oldHttps.Value {
		return nil
	}
	return settingsRepo.Update("Https", https)
}

func (c *Center) updateHttpsCertType(certType string) error {
	if len(certType) == 0 {
		return nil
	}
	settingsRepo := repo.NewSettingsRepo()
	oldCertType, err := settingsRepo.Get(settingsRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return err
	}
	if certType == oldCertType.Value {
		return nil
	}
	return settingsRepo.Update("HttpsCertType", certType)
}

func (c *Center) updateHttpsCertPath(certPath string) error {
	if len(certPath) == 0 {
		return nil
	}
	settingsRepo := repo.NewSettingsRepo()
	oldCertPath, err := settingsRepo.Get(settingsRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return err
	}
	if certPath == oldCertPath.Value {
		return nil
	}
	return settingsRepo.Update("HttpsCertPath", certPath)
}

func (c *Center) updateHttpsKeyPath(keyPath string) error {
	if len(keyPath) == 0 {
		return nil
	}
	settingsRepo := repo.NewSettingsRepo()
	oldKeyPath, err := settingsRepo.Get(settingsRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return err
	}
	if keyPath == oldKeyPath.Value {
		return nil
	}
	return settingsRepo.Update("HttpsKeyPath", keyPath)
}

func (c *Center) Upgrade() error {
	return c.upgrade()
}

func (c *Center) ResetAdminPassword() (string, error) {
	userRepo := repo.NewUserRepo()
	user, err := userRepo.Get(userRepo.WithByName("admin"))
	if err != nil {
		return "", errors.New("failed to get admin user")
	}

	salt := utils.GenerateNonce(8)
	newPass := utils.GeneratePassword(8)
	passwordHash := utils.HashPassword(newPass, salt)
	upMap := make(map[string]interface{})
	upMap["password"] = passwordHash
	upMap["salt"] = salt
	if err := UserRepo.Update(user.ID, upMap); err != nil {
		return "", errors.New("failed to reset admin password")
	}

	return newPass, nil
}

func (c *Center) upgrade() error {
	newVersion := c.getLatestVersion()
	if len(newVersion) == 0 {
		return errors.New("failed to get latest version")
	}

	if global.Version == newVersion {
		return errors.New("already up to date")
	}

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return errors.New("failed to get default host")
	}

	agentConn, err := c.getAgentConn(&host)
	if err != nil {
		global.LOG.Error("Failed to get agent connection")
		return errors.New("failed to get agent connection")
	}

	// 创建消息
	cmd := fmt.Sprintf("curl -sSL https://static.sensdata.com/idb/release/upgrade.sh -o /tmp/upgrade.sh && bash /tmp/upgrade.sh %s", newVersion)

	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		cmd,
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
		message.CmdMessage,
	)
	err = message.SendMessage(*agentConn, msg)
	if err != nil {
		global.LOG.Error("Failed to send command message: %v", err)
		return errors.New("failed to notify agent")
	}

	return nil
}

func (c *Center) getLatestVersion() string {
	cmd := fmt.Sprintf("curl -sSL %s", CONFMAN.GetConfig().Latest)
	global.LOG.Info("Getting latest version: %s", cmd)
	latest, err := utils.Exec(cmd)
	if err != nil {
		global.LOG.Error("Failed to get latest version: %v", err)
		return ""
	}
	global.LOG.Info("Got latest version: %s", latest)
	return strings.TrimSpace(latest)
}
