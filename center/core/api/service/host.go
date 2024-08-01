package service

import (
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/core/api/dto"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/utils"
)

type HostService struct{}

type IHostService interface {
	ListGroup(req dto.PageInfo) (*dto.PageResult, error)
	List(req dto.ListHost) (*dto.PageResult, error)
	Create(req dto.CreateHost) (*dto.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
	UpdateSSH(req dto.UpdateHostSSH) error
	UpdateAgent(req dto.UpdateHostAgent) error
	TestSSH(req dto.TestSSH) error
	TestAgent(req dto.TestAgent) error
}

func NewIHostService() IHostService {
	return &HostService{}
}

// List host group
func (s *HostService) ListGroup(req dto.PageInfo) (*dto.PageResult, error) {
	total, groups, err := HostGroupRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &dto.PageResult{Total: total, Items: groups}, nil
}

// List host
func (s *HostService) List(req dto.ListHost) (*dto.PageResult, error) {
	var opts []repo.DBOption
	opts = append(opts, HostRepo.WithByGroupID(req.GroupID))
	if req.Keyword != "" {
		opts = append(opts, HostRepo.WithByName(req.Keyword))
	}

	total, hosts, err := HostRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &dto.PageResult{Total: total, Items: hosts}, nil
}

func (s *HostService) Create(req dto.CreateHost) (*dto.HostInfo, error) {
	var host model.Host
	if err := copier.Copy(&host, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	//找组
	group, err := HostGroupRepo.Get(HostGroupRepo.WithByID(req.GroupID))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	//Agent参数设置为默认的先
	host.AgentAddr = req.Addr
	host.AgentPort = 25901                  //TODO 从设置中获取
	host.AgentKey = utils.GenerateNonce(32) //TODO 添加以后，如何给到Agent端？
	host.AgentMode = "http"                 //TODO https连接，需要调整实现

	if err := HostRepo.Create(&host); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	return &dto.HostInfo{
		ID:         host.ID,
		CreatedAt:  host.CreatedAt,
		GroupInfo:  dto.GroupInfo{ID: host.GroupID, GroupName: group.GroupName, CreatedAt: group.CreatedAt},
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

func (s *HostService) Delete(ids []uint) error {
	return HostRepo.Delete(CommonRepo.WithIdsIn(ids))
}

func (s *HostService) UpdateSSH(req dto.UpdateHostSSH) error {
	//找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	//更新字段
	upMap := make(map[string]interface{})
	upMap["addr"] = req.Addr
	upMap["port"] = req.Port
	upMap["user"] = req.User
	upMap["auth_mode"] = req.AuthMode
	upMap["password"] = req.Password
	upMap["private_key"] = req.PrivateKey
	upMap["pass_phrase"] = req.PassPhrase

	return HostRepo.Update(host.ID, upMap)
}

func (s *HostService) UpdateAgent(req dto.UpdateHostAgent) error {
	// 找host
	host, err := HostRepo.Get(HostRepo.WithByID(req.HostID))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 更新字段
	upMap := make(map[string]interface{})
	upMap["agent_addr"] = req.AgentAddr
	upMap["agent_port"] = req.AgentPort
	upMap["agent_user"] = req.AgentKey
	upMap["agent_mode"] = req.AgentMode

	return HostRepo.Update(host.ID, upMap)
}

func (s *HostService) TestSSH(req dto.TestSSH) error {
	var host model.Host
	if err := copier.Copy(&host, &req); err != nil {
		return err
	}
	if err := conn.SSH.TestConnection(host); err != nil {
		return err
	}
	return nil
}

func (s *HostService) TestAgent(req dto.TestAgent) error {
	if err := conn.CENTER.TestAgent(req); err != nil {
		return err
	}
	return nil
}
