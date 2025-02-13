package service

import (
	"encoding/base64"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type HostService struct{}

type IHostService interface {
	ListGroup(req core.PageInfo) (*core.PageResult, error)
	List(req core.ListHost) (*core.PageResult, error)
	Create(req core.CreateHost) (*core.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
	Status(id uint) (*core.HostStatus, error)
	UpdateSSH(id uint, req core.UpdateHostSSH) error
	UpdateAgent(id uint, req core.UpdateHostAgent) error
	TestSSH(id uint, req core.TestSSH) error
	TestAgent(id uint, req core.TestAgent) error
	InstallAgent(id uint) error
	AgentStatus(id uint) (*core.AgentStatus, error)
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

	// 私钥用 base64 编码一下
	var encodedPrivateKey string
	if req.AuthMode == "password" {
		encodedPrivateKey = ""
	} else {
		encodedPrivateKey = base64.StdEncoding.EncodeToString([]byte(req.PrivateKey))
	}
	host.PrivateKey = encodedPrivateKey

	//Agent参数设置为默认的先
	host.AgentAddr = req.Addr
	host.AgentPort = 9919                   //TODO 从设置中获取
	host.AgentKey = utils.GenerateNonce(32) //TODO 添加以后，如何给到Agent端？
	host.AgentMode = "http"                 //TODO https连接，需要调整实现

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
		PrivateKey: encodedPrivateKey,
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

func (s *HostService) Delete(ids []uint) error {
	return HostRepo.Delete(CommonRepo.WithIdsIn(ids))
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
		// Encode private key
		encodedPrivateKey := base64.StdEncoding.EncodeToString([]byte(req.PrivateKey))
		global.LOG.Info("private key content: \n %s", encodedPrivateKey)

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

func (s *HostService) TestSSH(id uint, req core.TestSSH) error {
	var host model.Host
	if err := copier.Copy(&host, &req); err != nil {
		return err
	}
	host.ID = id

	// 私钥用 base64 编码一下
	var encodedPrivateKey string
	if req.AuthMode == "password" {
		encodedPrivateKey = ""
	} else {
		encodedPrivateKey = base64.StdEncoding.EncodeToString([]byte(req.PrivateKey))
	}
	host.PrivateKey = encodedPrivateKey

	if err := conn.SSH.TestConnection(host); err != nil {
		return err
	}
	return nil
}

func (s *HostService) TestAgent(id uint, req core.TestAgent) error {
	if err := conn.CENTER.TestAgent(id, req); err != nil {
		return err
	}
	return nil
}

func (s *HostService) InstallAgent(id uint) error {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(id))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 安装
	if err := conn.SSH.InstallAgent(host); err != nil {
		return err
	}
	return nil
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
