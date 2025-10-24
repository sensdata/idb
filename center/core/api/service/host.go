package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/logstream/pkg/types"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type HostService struct{}

type IHostService interface {
	ListGroup(req core.PageInfo) (*core.PageResult, error)
	CreateGroup(req core.CreateGroup) (*core.GroupInfo, error)
	UpdateGroup(id uint, upMap map[string]interface{}) error
	DeleteGroup(ids []uint) error
	List(req core.ListHost) (*core.PageResult, error)
	Create(req core.CreateHost) (*core.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(id uint) error
	Info(id uint) (*core.HostInfo, error)
	Status(id uint) (*core.HostStatus, error)
	StatusFollow(c *gin.Context) error
	ActivateHost(id uint) error
	UpdateSSH(id uint, req core.UpdateHostSSH) error
	UpdateAgent(id uint, req core.UpdateHostAgent) error
	TestSSH(req core.TestSSH) error
	TestAgent(id uint, req core.TestAgent) error
	InstallAgent(id uint, req core.InstallAgent) (*core.LogInfo, error)
	UninstallAgent(id uint) (*core.LogInfo, error)
	AgentStatus(id uint) (*core.AgentStatus, error)
	AgentStatusFollow(c *gin.Context) error
	RestartAgent(id uint) error
}

func NewIHostService() IHostService {
	return &HostService{}
}

// List host group
func (s *HostService) ListGroup(req core.PageInfo) (*core.PageResult, error) {
	total, groups, err := HostGroupRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &core.PageResult{Total: total, Items: groups}, nil
}

func (s *HostService) CreateGroup(req core.CreateGroup) (*core.GroupInfo, error) {
	var group model.HostGroup
	if err := copier.Copy(&group, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	//检查创建的名称是default，如果是则不允许
	if req.GroupName == "default" {
		return nil, errors.WithMessage(constant.ErrInternalServer, "can't create group with name default")
	}
	if err := HostGroupRepo.Create(&group); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}
	return &core.GroupInfo{ID: group.ID, GroupName: group.GroupName, CreatedAt: group.CreatedAt}, nil
}

func (s *HostService) UpdateGroup(id uint, upMap map[string]interface{}) error {
	// 检查更新的名称是否是default，如果则不允许
	if upMap["group_name"] == "default" {
		return errors.WithMessage(constant.ErrInternalServer, "can't update group with name default")
	}

	return HostGroupRepo.Update(id, upMap)
}

func (s *HostService) DeleteGroup(ids []uint) error {
	defaultGroup, err := HostGroupRepo.Get(HostGroupRepo.WithByName("default"))
	if err != nil {
		return errors.WithMessage(constant.ErrInternalServer, err.Error())
	}
	//判断如果ids中包含 defaultGroup.ID
	for _, id := range ids {
		if id == defaultGroup.ID {
			return errors.WithMessage(constant.ErrInternalServer, "can't delete default group")
		}
	}

	return HostGroupRepo.Delete(CommonRepo.WithIdsIn(ids))
}

// List host
func (s *HostService) List(req core.ListHost) (*core.PageResult, error) {
	var opts []repo.DBOption
	opts = append(opts, HostRepo.WithByGroupID(req.GroupID))
	if req.Keyword != "" {
		opts = append(opts, HostRepo.WithByName(req.Keyword))
	}

	_, hosts, err := HostRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	// 最新的agent版本
	latestVersion := getAgentLatest()

	var hostsInfos []core.HostInfo
	for _, host := range hosts {
		//找组
		var group core.GroupInfo
		g, err := HostGroupRepo.Get(HostGroupRepo.WithByID(host.GroupID))
		if err != nil {
			group = core.GroupInfo{}
		} else {
			group = core.GroupInfo{ID: g.ID, GroupName: g.GroupName, CreatedAt: g.CreatedAt}
		}

		// 查询状态
		var agentStatus = global.GetAgentStatus(host.ID)
		if agentStatus == nil {
			agentStatus = &core.AgentStatus{
				Status:    "unknown",
				Connected: "unknown",
			}
		}

		hostsInfos = append(
			hostsInfos,
			core.HostInfo{
				ID:           host.ID,
				CreatedAt:    host.CreatedAt,
				Default:      host.IsDefault,
				GroupInfo:    group,
				Name:         host.Name,
				Addr:         host.Addr,
				Port:         host.Port,
				User:         host.User,
				AuthMode:     host.AuthMode,
				Password:     host.Password,
				PrivateKey:   host.PrivateKey,
				PassPhrase:   host.PassPhrase,
				AgentAddr:    host.AgentAddr,
				AgentPort:    host.AgentPort,
				AgentKey:     host.AgentKey,
				AgentMode:    host.AgentMode,
				AgentVersion: host.AgentVersion,
				AgentStatus:  *agentStatus,
				AgentLatest:  latestVersion,
			},
		)
	}

	return &core.PageResult{Total: int64(len(hostsInfos)), Items: hostsInfos}, nil
}

func getAgentLatest() string {
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
	return latestVersion
}

func (s *HostService) Create(req core.CreateHost) (*core.HostInfo, error) {
	//找组
	group, err := HostGroupRepo.Get(HostGroupRepo.WithByID(req.GroupID))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	var host model.Host
	if err := copier.Copy(&host, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	host.IsDefault = false

	//Agent参数设置为默认的先
	host.AgentAddr = req.Addr
	host.AgentPort = 9919
	host.AgentKey = utils.GenerateNonce(24) // 随机生成
	host.AgentMode = "https"
	host.AgentVersion = ""

	if err := HostRepo.Create(&host); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	return &core.HostInfo{
		ID:           host.ID,
		CreatedAt:    host.CreatedAt,
		Default:      host.IsDefault,
		GroupInfo:    core.GroupInfo{ID: host.GroupID, GroupName: group.GroupName, CreatedAt: group.CreatedAt},
		Name:         host.Name,
		Addr:         host.Addr,
		Port:         host.Port,
		User:         host.User,
		AuthMode:     host.AuthMode,
		Password:     host.Password,
		PrivateKey:   host.PrivateKey,
		PassPhrase:   host.PassPhrase,
		AgentAddr:    host.AgentAddr,
		AgentPort:    host.AgentPort,
		AgentKey:     host.AgentKey,
		AgentMode:    host.AgentMode,
		AgentVersion: host.AgentVersion,
	}, nil
}

func (s *HostService) Update(id uint, upMap map[string]interface{}) error {
	return HostRepo.Update(id, upMap)
}

func (s *HostService) Delete(id uint) error {
	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// default host不可以删除
	if host.IsDefault {
		return errors.WithMessage(constant.ErrInternalServer, "can't delete default host")
	}

	// 断开agent conn
	err = conn.CENTER.ReleaseAgentConn(host)
	if err != nil {
		global.LOG.Error("failed to release agent conn: %v", err)
	}

	return HostRepo.Delete(CommonRepo.WithIdsIn([]uint{host.ID}))
}

func (s *HostService) Info(id uint) (*core.HostInfo, error) {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		global.LOG.Error("host %d not found: %v", id, err)
		return nil, constant.ErrInternalServer
	}

	//找组
	group, err := HostGroupRepo.Get(HostGroupRepo.WithByID(host.GroupID))
	if err != nil {
		global.LOG.Error("group %d not found: %v", host.GroupID, err)
		return nil, constant.ErrInternalServer
	}

	// 查询状态
	var agentStatus = global.GetAgentStatus(host.ID)
	if agentStatus == nil {
		agentStatus = &core.AgentStatus{
			Status:    "unknown",
			Connected: "unknown",
		}
	}

	return &core.HostInfo{
		ID:           host.ID,
		CreatedAt:    host.CreatedAt,
		Default:      host.IsDefault,
		GroupInfo:    core.GroupInfo{ID: host.GroupID, GroupName: group.GroupName, CreatedAt: group.CreatedAt},
		Name:         host.Name,
		Addr:         host.Addr,
		Port:         host.Port,
		User:         host.User,
		AuthMode:     host.AuthMode,
		Password:     host.Password,
		PrivateKey:   host.PrivateKey,
		PassPhrase:   host.PassPhrase,
		AgentAddr:    host.AgentAddr,
		AgentPort:    host.AgentPort,
		AgentKey:     host.AgentKey,
		AgentMode:    host.AgentMode,
		AgentVersion: host.AgentVersion,
		AgentStatus:  *agentStatus,
	}, nil
}

func (s *HostService) Status(id uint) (*core.HostStatus, error) {
	hostStatus := global.GetHostStatus(id)
	if hostStatus == nil {
		return &core.HostStatus{}, nil
	}
	return hostStatus, nil
}

func (s *HostService) StatusFollow(c *gin.Context) error {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		return errors.New("invalid host")
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 使用 context 来控制超时和客户端断开
	ctx := c.Request.Context()
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	defer func() {
		if r := recover(); r != nil {
			global.LOG.Warn("Recovered in SSE loop: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			global.LOG.Info("SSE DONE")
			return nil
		default:
			status, err := s.Status(uint(hostID))
			if err != nil {
				global.LOG.Error("get status failed: %v", err)
				c.SSEvent("error", err.Error())
			} else {
				statusJson, err := utils.ToJSONString(status)
				if err != nil {
					global.LOG.Error("json err: %v", err)
					c.SSEvent("error", err.Error())
				} else {
					c.SSEvent("status", statusJson)
				}
			}
			flusher.Flush()
		}
		time.Sleep(2 * time.Second)
	}
}

func (s *HostService) ActivateHost(id uint) error {
	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 激活host
	return conn.CENTER.ActivateHost(&host)
}

func (s *HostService) UpdateSSH(id uint, req core.UpdateHostSSH) error {
	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	//更新字段
	upMap := make(map[string]interface{})
	upMap["addr"] = req.Addr
	upMap["port"] = req.Port
	upMap["user"] = req.User
	upMap["auth_mode"] = req.AuthMode

	// 校验模式
	if req.AuthMode == "password" {
		upMap["password"] = req.Password
	} else {
		upMap["private_key"] = req.PrivateKey
		upMap["pass_phrase"] = req.PassPhrase
	}

	// agent连接地址也改掉
	upMap["agent_addr"] = req.Addr

	return HostRepo.Update(host.ID, upMap)
}

func (s *HostService) UpdateAgent(id uint, req core.UpdateHostAgent) error {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 更新字段
	upMap := make(map[string]interface{})
	upMap["agent_addr"] = req.AgentAddr
	upMap["agent_port"] = req.AgentPort
	// if req.AgentKey != "" {
	// 	upMap["agent_key"] = req.AgentKey
	// }
	// upMap["agent_mode"] = req.AgentMode

	return HostRepo.Update(host.ID, upMap)
}

func (s *HostService) TestSSH(req core.TestSSH) error {
	var host model.Host
	if err := copier.Copy(&host, &req); err != nil {
		return err
	}

	if err := conn.SSH.TestConnection(host); err != nil {
		return err
	}
	return nil
}

func (s *HostService) TestAgent(id uint, req core.TestAgent) error {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	if err := conn.CENTER.TestAgent(host, req); err != nil {
		return err
	}
	return nil
}

func (s *HostService) InstallAgent(id uint, req core.InstallAgent) (*core.LogInfo, error) {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrHostNotFound
	}

	defaultHost, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		return nil, err
	}

	// 生成任务
	task, err := global.LogStream.CreateTask(types.TaskTypeFile, nil)
	if err != nil {
		return nil, err
	}

	// 异步安装
	go conn.SSH.InstallAgent(host, task.ID, req.Upgrade)

	// 先返回task信息
	return &core.LogInfo{LogHost: defaultHost.ID, LogPath: task.LogPath}, nil
}

func (s *HostService) UninstallAgent(id uint) (*core.LogInfo, error) {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrHostNotFound
	}

	defaultHost, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		return nil, err
	}

	// 生成任务
	task, err := global.LogStream.CreateTask(types.TaskTypeFile, nil)
	if err != nil {
		return nil, err
	}

	// 异步卸载
	go conn.SSH.UninstallAgent(host, task.ID)

	// 先返回task信息
	return &core.LogInfo{LogHost: defaultHost.ID, LogPath: task.LogPath}, nil
}

func (s *HostService) AgentStatus(id uint) (*core.AgentStatus, error) {
	agentStatus := global.GetAgentStatus(id)
	if agentStatus == nil {
		return &core.AgentStatus{
			Status:    "unknown",
			Connected: "unknown",
		}, nil
	}
	return agentStatus, nil
}

func (s *HostService) AgentStatusFollow(c *gin.Context) error {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		return errors.New("invalid host")
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 使用 context 来控制超时和客户端断开
	ctx := c.Request.Context()
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	defer func() {
		if r := recover(); r != nil {
			global.LOG.Warn("Recovered in SSE loop: %v", r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			global.LOG.Info("SSE DONE")
			return nil
		default:
			status, err := s.AgentStatus(uint(hostID))
			if err != nil {
				global.LOG.Error("get agent status failed: %v", err)
				c.SSEvent("error", err.Error())
			} else {
				statusJson, err := utils.ToJSONString(status)
				if err != nil {
					global.LOG.Error("json err: %v", err)
					c.SSEvent("error", err.Error())
				} else {
					c.SSEvent("status", statusJson)
				}
			}
			flusher.Flush()
		}
		time.Sleep(1 * time.Second)
	}
}

func (s *HostService) RestartAgent(id uint) error {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return constant.ErrRecordNotFound
	}

	// restart
	err = conn.SSH.RestartAgent(host)
	if err != nil {
		return err
	}

	return nil
}
