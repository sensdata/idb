package service

import (
	"encoding/base64"
	"fmt"
	"os"

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
	List(req core.ListHost) (*core.PageResult, error)
	Create(req core.CreateHost) (*core.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(id uint) error
	Info(id uint) (*core.HostInfo, error)
	Status(id uint) (*core.HostStatus, error)
	UpdateSSH(id uint, req core.UpdateHostSSH) error
	UpdateAgent(id uint, req core.UpdateHostAgent) error
	TestSSH(req core.TestSSH) error
	TestAgent(id uint, req core.TestAgent) error
	InstallAgent(id uint) (*core.TaskInfo, error)
	UninstallAgent(id uint) (*core.TaskInfo, error)
	AgentStatus(id uint) (*core.AgentStatus, error)
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

// List host
func (s *HostService) List(req core.ListHost) (*core.PageResult, error) {
	var opts []repo.DBOption
	opts = append(opts, HostRepo.WithByGroupID(req.GroupID))
	if req.Keyword != "" {
		opts = append(opts, HostRepo.WithByName(req.Keyword))
	}

	total, hosts, err := HostRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &core.PageResult{Total: total, Items: hosts}, nil
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

	//Agent参数设置为默认的先
	host.AgentAddr = req.Addr
	host.AgentPort = 9919                      //TODO 从设置中获取
	host.AgentKey = "idbidbidbidbidbidbidbidb" //TODO 添加以后，如何给到Agent端？
	host.AgentMode = "https"                   //TODO https连接，需要调整实现

	if err := HostRepo.Create(&host); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	return &core.HostInfo{
		ID:         host.ID,
		CreatedAt:  host.CreatedAt,
		GroupInfo:  core.GroupInfo{ID: host.GroupID, GroupName: group.GroupName, CreatedAt: group.CreatedAt},
		Name:       host.Name,
		Addr:       host.Addr,
		Port:       host.Port,
		User:       host.User,
		AuthMode:   host.AuthMode,
		Password:   host.Password,
		PrivateKey: host.PrivateKey,
		PassPhrase: host.PassPhrase,
		AgentAddr:  host.AgentAddr,
		AgentPort:  host.AgentPort,
		AgentKey:   host.AgentKey,
		AgentMode:  host.AgentMode,
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

	return &core.HostInfo{
		ID:         host.ID,
		CreatedAt:  host.CreatedAt,
		GroupInfo:  core.GroupInfo{ID: host.GroupID, GroupName: group.GroupName, CreatedAt: group.CreatedAt},
		Name:       host.Name,
		Addr:       host.Addr,
		Port:       host.Port,
		User:       host.User,
		AuthMode:   host.AuthMode,
		Password:   host.Password,
		PrivateKey: host.PrivateKey,
		PassPhrase: host.PassPhrase,
		AgentAddr:  host.AgentAddr,
		AgentPort:  host.AgentPort,
		AgentKey:   host.AgentKey,
		AgentMode:  host.AgentMode,
	}, nil
}

func (s *HostService) Status(id uint) (*core.HostStatus, error) {
	var status core.HostStatus

	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	actionRequest := core.HostAction{
		HostID: host.ID,
		Action: core.Action{
			Action: core.Host_Status,
			Data:   "",
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return &status, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return &status, fmt.Errorf("failed to query sessions")
	}

	err = utils.FromJSONString(actionResponse.Data, &status)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to session list: %v", err)
		return &status, fmt.Errorf("json err: %v", err)
	}

	return &status, nil
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
		// 读取文件
		privateKey, err := os.ReadFile(req.PrivateKey)
		if err != nil {
			global.LOG.Error("failed to read private key file: %v", err)
			return errors.New(constant.ErrFileRead)
		}
		encodedPrivateKey := base64.StdEncoding.EncodeToString(privateKey)

		upMap["private_key"] = encodedPrivateKey
		upMap["pass_phrase"] = req.PassPhrase
	}

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
	if req.AgentKey != "" {
		upMap["agent_key"] = req.AgentKey
	}
	upMap["agent_mode"] = req.AgentMode

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

func (s *HostService) InstallAgent(id uint) (*core.TaskInfo, error) {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrHostNotFound
	}

	// 生成任务
	taskId, err := global.LogStream.CreateTask(types.TaskTypeBuffer, nil)
	if err != nil {
		return nil, err
	}

	// 异步安装
	go conn.SSH.InstallAgent(host, taskId)

	// 先返回task信息
	return &core.TaskInfo{TaskID: taskId}, nil
}

func (s *HostService) UninstallAgent(id uint) (*core.TaskInfo, error) {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrHostNotFound
	}

	// 生成任务
	taskId, err := global.LogStream.CreateTask(types.TaskTypeBuffer, nil)
	if err != nil {
		return nil, err
	}

	// 异步卸载
	go conn.SSH.UninstallAgent(host, taskId)

	// 先返回task信息
	return &core.TaskInfo{TaskID: taskId}, nil
}

func (s *HostService) AgentStatus(id uint) (*core.AgentStatus, error) {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 查询安装状态
	status, err := conn.SSH.AgentStatus(host)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrSsh, err.Error())
	}

	// 查询连接状态
	if conn.CENTER.IsAgentConnected(host) {
		status.Connected = "online"
	} else {
		status.Connected = "offline"
	}

	return status, nil
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
