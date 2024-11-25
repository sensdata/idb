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
	"github.com/sensdata/idb/agent/agent/docker"
	"github.com/sensdata/idb/agent/agent/file"
	"github.com/sensdata/idb/agent/agent/git"
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
	CONFMAN       *config.Manager
	AGENT         IAgent
	FileService   = file.NewIFileService()
	SshService    = ssh.NewISSHService()
	GitService    = git.NewIGitService()
	DockerService = docker.NewIDockerService()
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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

		// TODO: 上传文件
	case model.File_Upload:
		return actionSuccessResult(actionData.Action, "")

		// TODO: 下载文件
	case model.File_Download:
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, "")

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
		return actionSuccessResult(actionData.Action, result)

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
		return actionSuccessResult(actionData.Action, result)

		// git 初始化
	case model.Git_Init:
		var req model.GitInit
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.InitRepo(req.RepoPath, req.IsBare)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 文件列表
	case model.Git_File_List:
		var req model.GitQuery
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		files, err := GitService.GetFileList(req.RepoPath, req.RelativePath, req.Extension, req.Page, req.PageSize)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(files)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 文件信息
	case model.Git_File:
		var req model.GitGetFile
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		file, err := GitService.GetFile(req.RepoPath, req.RelativePath)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(file)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 创建文件
	case model.Git_Create:
		var req model.GitCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Create(req.RepoPath, req.RelativePath, req.Content)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 更新文件
	case model.Git_Update:
		var req model.GitUpdate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Update(req.RepoPath, req.RelativePath, req.Content)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 删除文件
	case model.Git_Delete:
		var req model.GitDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Delete(req.RepoPath, req.RelativePath)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 恢复文件
	case model.Git_Restore:
		var req model.GitRestore
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := GitService.Restore(req.RepoPath, req.RelativePath, req.CommitHash)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// git 文件历史
	case model.Git_Log:
		var req model.GitLog
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		logs, err := GitService.Log(req.RepoPath, req.RelativePath, req.Page, req.PageSize)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(logs)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// git 文件差异
	case model.Git_Diff:
		var req model.GitDiff
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		diffs, err := GitService.Diff(req.RepoPath, req.RelativePath, req.CommitHash)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(diffs)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 执行脚本
	case model.Script_Exec:
		var req model.ScriptExec
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		scriptResult := shell.ExecuteScript(req)
		result, err := utils.ToJSONString(scriptResult)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// docker 状态
	case model.Docker_Status:
		status, err := DockerService.DockerStatus()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(status)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// docker 配置
	case model.Docker_Conf:
		conf, err := DockerService.DockerConf()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(conf)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// 更新 docker 配置
	case model.Docker_Upd_Conf:
		var req model.KeyValue
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateConf(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 更新 docker 配置
	case model.Docker_Upd_Conf_File:
		var req model.DaemonJsonUpdateByFile
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateConfByFile(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 更新 docker log配置
	case model.Docker_Upd_Log:
		var req model.LogOption
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateLogOption(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// 更新 docker ipv6配置
	case model.Docker_Upd_Ipv6:
		var req model.Ipv6Option
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerUpdateIpv6Option(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// docker 操作
	case model.Docker_Operation:
		var req model.DockerOperation
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.DockerOperation(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

		// inspect
	case model.Docker_Inspect:
		var req model.Inspect
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.Inspect(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

		// prune
	case model.Docker_Prune:
		var req model.Prune
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		report, err := DockerService.Prune(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(report)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Query:
		var req model.QueryContainer
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ContainerQuery(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Names:
		names, err := DockerService.ContainerNames()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(names)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Create:
		var req model.ContainerOperate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerCreate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Update:
		var req model.ContainerOperate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerUpdate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Upgrade:
		var req model.ContainerUpgrade
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerUpgrade(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Info:
		containerID := actionData.Data
		info, err := DockerService.ContainerInfo(containerID)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Resource_Usage:
		info, err := DockerService.ContainerResourceUsage()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Resource_Limit:
		info, err := DockerService.ContainerResourceLimit()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Stats:
		containerID := actionData.Data
		info, err := DockerService.ContainerStats(containerID)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Container_Rename:
		var req model.Rename
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerRename(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Log_Clean:
		containerID := actionData.Data
		err := DockerService.ContainerLogClean(containerID)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Operation:
		var req model.ContainerOperation
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ContainerOperation(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Container_Logs:
		// var req model.ContainerOperation
		// if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
		// 	return nil, err
		// }
		// err := DockerService.ContainerLogs(req)
		// if err != nil {
		// 	return nil, err
		// }
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Image_Page:
		var req model.SearchPageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImagePage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_List:
		info, err := DockerService.ImageList()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Build:
		var req model.ImageBuild
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImageBuild(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Pull:
		var req model.ImagePull
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImagePull(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Load:
		var req model.ImageLoad
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImageLoad(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Save:
		var req model.ImageSave
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ImageSave(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Image_Push:
		var req model.ImagePush
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.ImagePush(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Image_Remove:
		var req model.BatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ImageRemove(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Image_Tag:
		var req model.ImageTag
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.ImageTag(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Volume_Page:
		var req model.SearchPageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.VolumePage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Volume_List:
		info, err := DockerService.VolumeList()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Volume_Delete:
		var req model.BatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.VolumeDelete(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Volume_Create:
		var req model.VolumeCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.VolumeCreate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Network_Page:
		var req model.SearchPageInfo
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		info, err := DockerService.NetworkPage(req)
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Network_List:
		info, err := DockerService.NetworkList()
		if err != nil {
			return nil, err
		}
		result, err := utils.ToJSONString(info)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, result)

	case model.Docker_Network_Delete:
		var req model.BatchDelete
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.NetworkDelete(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Network_Create:
		var req model.NetworkCreate
		if err := json.Unmarshal([]byte(actionData.Data), &req); err != nil {
			return nil, err
		}
		err := DockerService.NetworkCreate(req)
		if err != nil {
			return nil, err
		}
		return actionSuccessResult(actionData.Action, "")

	case model.Docker_Compose_Page:
		return actionSuccessResult(actionData.Action, "")
	case model.Docker_Compose_Create:
		return actionSuccessResult(actionData.Action, "")
	case model.Docker_Compose_Operation:
		return actionSuccessResult(actionData.Action, "")
	case model.Docker_Compose_Test:
		return actionSuccessResult(actionData.Action, "")
	case model.Docker_Compose_Update:
		return actionSuccessResult(actionData.Action, "")

	default:
		return nil, nil
	}
}

func actionSuccessResult(action string, data string) (*model.Action, error) {
	return &model.Action{
		Action: action,
		Result: true,
		Data:   data,
	}, nil
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
