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

	"github.com/sensdata/idb/agent/agent/action"
	"github.com/sensdata/idb/agent/agent/file"
	"github.com/sensdata/idb/agent/agent/ssh"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/shell"
	"github.com/sensdata/idb/core/utils"
)

var (
	CONFMAN     *config.Manager
	AGENT       IAgent
	FileService = file.NewIFileService()
	SshService  = ssh.NewISSHService()
)

type Agent struct {
	unixListener net.Listener
	tcpListener  net.Listener
	centerID     string   // 存储center地址
	centerConn   net.Conn // 存储center端连接的映射
	done         chan struct{}
	mu           sync.Mutex // 保护centerConn的互斥锁
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
	sockFile := filepath.Join(constant.AgentRunDir, constant.AgentSock)
	os.Remove(sockFile)

	return nil
}

func (a *Agent) listenToUnix() error {
	//先关闭
	if a.unixListener != nil {
		a.unixListener.Close()
	}

	// 检查sock文件
	sockFile := filepath.Join(constant.AgentRunDir, constant.AgentSock)

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
				a.processMessage(conn, m)
			case *message.FileMessage:
				a.processFileMessage(conn, m)
			default:
				fmt.Println("Unknown message type")
			}

			// 更新缓存，移除已处理的部分
			dataBuffer = remainingBuffer
		}
	}

	global.LOG.Info("Connection closed: %s", conn.RemoteAddr().String())
}

func (a *Agent) processMessage(conn net.Conn, msg *message.Message) {
	global.LOG.Info("Message: %v", msg)

	centerID := conn.RemoteAddr().String()

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
			a.sendActionResult(conn, msg.MsgID, &model.Action{Action: "", Result: false, Data: ""})
		} else {
			a.sendActionResult(conn, msg.MsgID, result)
		}
	default: // 不支持的消息
		global.LOG.Error("Unknown message type: %s", msg.Type)
	}
}

func (a *Agent) processFileMessage(conn net.Conn, msg *message.FileMessage) {
	global.LOG.Info("FileMessage: %s, %d, %d", msg.FileName, msg.Offset, msg.ChunkSize)

	switch msg.Type {
	case message.Upload: //上传
		err := files.NewFileOp().WriteChunkToFile(
			msg.Path,
			msg.FileName,
			msg.Offset,
			msg.ChunkSize,
			msg.Chunk,
		)
		if err != nil {
			global.LOG.Error("Failed to process upload: %v", err)
			a.sendUploadResult(conn, msg, message.FileErr)
		} else {
			status := message.FileOk
			//如果是最后一次传输，则设置为Done
			if msg.Offset+int64(msg.ChunkSize) == msg.TotalSize {
				status = message.FileDone
			}
			a.sendUploadResult(conn, msg, status)
		}
	case message.Download: //下载
		totalSize, bytesRead, chunk, err := files.NewFileOp().ReadChunkFromFile(
			filepath.Join(msg.Path, msg.FileName),
			msg.Offset,
			msg.ChunkSize,
		)
		global.LOG.Info("read from file: %d %d", totalSize, bytesRead)
		if err != nil {
			global.LOG.Error("Failed to process download: %v", err)
			msg.Status = message.FileErr
		} else {
			msg.TotalSize = totalSize
			msg.ChunkSize = bytesRead
			msg.Chunk = chunk
			msg.Status = message.FileOk
			//如果是最后一次传输，则设置为Done
			if msg.Offset+int64(msg.ChunkSize) == msg.TotalSize {
				msg.Status = message.FileDone
			}
		}
		a.sendDownloadResult(conn, msg)
	}
}

func (a *Agent) processAction(data string) (*model.Action, error) {
	var actionData model.Action
	if err := utils.FromJSONString(data, &actionData); err != nil {
		return nil, err
	}

	switch actionData.Action {
	// 获取overview
	case model.SysInfo_OverView:
		overview, err := action.GetOverview()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(overview)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 获取network
	case model.SysInfo_Network:
		network, err := action.GetNetwork()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(network)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 获取SystemInfo
	case model.SysInfo_System:
		systemInfo, err := action.GetSystemInfo()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(systemInfo)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

		// 文件树
	case model.File_Tree:
		var fileOption model.FileOption
		if err := json.Unmarshal([]byte(actionData.Data), &fileOption); err != nil {
			return nil, err
		}

		fileInfo, err := FileService.GetFileTree(fileOption)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(fileInfo)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 列出文件
	case model.File_List:
		var fileOption model.FileOption
		if err := json.Unmarshal([]byte(actionData.Data), &fileOption); err != nil {
			return nil, err
		}

		fileInfo, err := FileService.GetFileList(fileOption)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(fileInfo)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

		// TODO: 上传文件
	case model.File_Upload:
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

		// TODO: 下载文件
	case model.File_Download:
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 创建文件
	case model.File_Create:
		var req model.FileCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.Create(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 删除文件
	case model.File_Delete:
		var req model.FileDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.Delete(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

		// 批量删除
	case model.File_Batch_Delete:
		var req model.FileBatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.BatchDelete(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 批量修改用户/组
	case model.File_Batch_Change_Owner:
		var req model.FileRoleReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.BatchChangeModeAndOwner(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

		// 修改文件权限
	case model.File_Change_Mode:
		var req model.FileCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.ChangeMode(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 修改文件用户/组
	case model.File_Change_Owner:
		var req model.FileRoleUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.ChangeOwner(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 修改文件名称
	case model.File_Change_Name:
		var req model.FileRename

		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.ChangeName(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 压缩文件
	case model.File_Compress:
		var req model.FileCompress

		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.Compress(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 解压文件
	case model.File_Decompress:
		var req model.FileDeCompress
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.DeCompress(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 获取文件内容
	case model.File_Content:
		var req model.FileContentReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		fileInfo, err := FileService.GetContent(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(fileInfo)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 保存文件内容
	case model.File_Content_Modify:
		var req model.FileEdit
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.SaveContent(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 移动文件
	case model.File_Move:
		var req model.FileMove
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.MvFile(req)
		if err != nil {
			return nil, err
		}
		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 目录大小
	case model.File_Dir_Size:
		var req model.DirSizeReq
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		rsp, err := FileService.DirSize(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(rsp)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 收藏列表
	case model.Favorite_List:
		var req model.PageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		rsp, err := FileService.GetFavoriteList(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(rsp)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 创建收藏
	case model.Favorite_Create:
		var req model.FavoriteCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		rsp, err := FileService.CreateFavorite(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(rsp)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 删除收藏
	case model.Favorite_Delete:
		var req model.FavoriteDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := FileService.DeleteFavorite(req)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 获取ssh配置
	case model.Ssh_Config:
		info, err := SshService.GetConfig()
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 更新ssh配置
	case model.Ssh_Config_Update:
		var req model.SSHUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.UpdateConfig(req)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 获取ssh配置文件内容
	case model.Ssh_Config_Content:
		content, err := SshService.GetContent()
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(content)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// 更新ssh配置文件内容
	case model.Ssh_Config_Content_Update:
		var req model.ContentUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.UpdateContent(req)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 操作ssh
	case model.Ssh_Operate:
		var req model.SSHOperate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.OperateSSH(req)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 创建秘钥
	case model.Ssh_Secret_Create:
		var req model.GenerateKey
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		err := SshService.CreateKey(req)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   "",
		}, nil

	// 枚举秘钥
	case model.Ssh_Secret:
		var req model.ListKey
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		keys, err := SshService.ListKeys(req)
		if err != nil {
			return nil, err
		}

		result, err := utils.ToJSONString(keys)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	// ssh日志
	case model.Ssh_Log:
		var req model.SearchSSHLog
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}

		log, _ := SshService.LoadLog(req)
		result, err := utils.ToJSONString(log)
		if err != nil {
			return nil, err
		}

		return &model.Action{
			Action: actionData.Action,
			Result: true,
			Data:   result,
		}, nil

	default:
		return nil, nil
	}
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

func (a *Agent) sendUploadResult(conn net.Conn, msg *message.FileMessage, status int) {
	centerID := conn.RemoteAddr().String()

	rspMsg, err := message.CreateFileMessage(
		msg.MsgID,
		msg.Type,
		status,
		msg.Path,
		msg.FileName,
		msg.TotalSize,
		msg.Offset,
		0,
		nil,
	)
	if err != nil {
		global.LOG.Error("Error creating file rsp message: %v", err)
		return
	}

	err = message.SendFileMessage(conn, rspMsg)
	if err != nil {
		global.LOG.Error("Failed to send file rsp : %v", err)
		return
	}
	global.LOG.Info("File rsp send to %s", centerID)
}

func (a *Agent) sendDownloadResult(conn net.Conn, msg *message.FileMessage) {
	centerID := conn.RemoteAddr().String()

	err := message.SendFileMessage(conn, msg)
	if err != nil {
		global.LOG.Error("Failed to send file rsp : %v", err)
		return
	}
	global.LOG.Info("File rsp send to %s", centerID)
}
