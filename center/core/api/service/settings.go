package service

import (
	"errors"
	"strconv"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils/common"
	"github.com/sensdata/idb/core/utils/systemctl"
)

type SettingsService struct{}

type ISettingsService interface {
	Profile(userId uint) (*model.Profile, error)
	About() (*model.About, error)
	Settings() (*model.SettingInfo, error)
	Update(req model.UpdateSettingRequest) error
}

func NewISettingsService() ISettingsService {
	return &SettingsService{}
}

func (s *SettingsService) Profile(userId uint) (*model.Profile, error) {
	user, err := UserRepo.Get(UserRepo.WithByID(userId))
	if err != nil {
		return nil, err
	}
	return &model.Profile{
		ID:   user.ID,
		Name: user.Name,
	}, nil
}

func (s *SettingsService) About() (*model.About, error) {
	var about model.About

	about.Version = global.Version

	return &about, nil
}

func (s *SettingsService) Settings() (*model.SettingInfo, error) {
	bindIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return nil, err
	}
	bindPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
	if err != nil {
		return nil, err
	}
	bindPortValue, err := strconv.Atoi(bindPort.Value)
	if err != nil {
		return nil, err
	}
	https, err := SettingsRepo.Get(SettingsRepo.WithByKey("Https"))
	if err != nil {
		return nil, err
	}
	httpsCertType, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return nil, err
	}
	httpsCertPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return nil, err
	}
	httpsKeyPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return nil, err
	}

	return &model.SettingInfo{
		BindIP:        bindIP.Value,
		BindPort:      bindPortValue,
		Https:         https.Value,
		HttpsCertType: httpsCertType.Value,
		HttpsCertPath: httpsCertPath.Value,
		HttpsKeyPath:  httpsKeyPath.Value,
	}, nil
}

func (s *SettingsService) Update(req model.UpdateSettingRequest) error {

	switch req.Https {
	case "no":
	case "yes":
		// 检查type和path
		switch req.HttpsCertType {
		case "default":
		case "path":
			if len(req.HttpsCertPath) == 0 || len(req.HttpsKeyPath) == 0 {
				return errors.New("invalid cert path or key path")
			}
		default:
			return errors.New("invalid cert type")
		}
	default:
		return errors.New("invalid https value")
	}

	// 开始事务
	tx := global.DB.Begin()
	var err error // 用于跟踪错误
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			global.LOG.Error("Transaction failed: %v  - rollback", r)
		} else if err != nil {
			tx.Rollback() // 如果发生错误，回滚事务
			global.LOG.Error("Error Happend - rollback")
		}
	}()

	if err = s.updateBindIP(req.BindIP); err != nil {
		global.LOG.Error("Failed to save BindIP to %s: %v", req.BindIP, err)
		return err
	}

	if err = s.updateBindPort(req.BindPort); err != nil {
		global.LOG.Error("Failed to save BindPort to %d: %v", req.BindPort, err)
		return err
	}

	if err = s.updateBindDomain(req.BindDomain); err != nil {
		global.LOG.Error("Failed to save BindDomain to %s: %v", req.BindDomain, err)
		return err
	}

	if err = s.updateHttps(req); err != nil {
		global.LOG.Error("Failed to save Https settings: %v", err)
		return err
	}

	// 提交事务
	tx.Commit()

	go func() {
		// TODO: 确定重启的方式和细节
		systemctl.Restart("idb.service")
	}()

	return nil
}

func (s *SettingsService) updateBindIP(newIP string) error {
	if len(newIP) == 0 {
		return errors.New("invalid bind ip")
	}

	oldIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return err
	}
	if newIP == oldIP.Value {
		return nil
	}

	return SettingsRepo.Update("BindIP", newIP)
}

func (s *SettingsService) updateBindPort(newPort int) error {
	if newPort <= 0 || newPort > 65535 {
		return errors.New("server port must between 1 - 65535")
	}
	oldPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
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

	return SettingsRepo.Update("BindPort", newPortStr)
}

func (s *SettingsService) updateBindDomain(newDomain string) error {
	if len(newDomain) == 0 {
		return nil
	}
	oldDomain, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindDomain"))
	if err != nil {
		return err
	}
	if newDomain == oldDomain.Value {
		return nil
	}
	return SettingsRepo.Update("BindDomain", newDomain)
}

func (s *SettingsService) updateHttps(req model.UpdateSettingRequest) error {
	if err := SettingsRepo.Update("Https", req.Https); err != nil {
		return err
	}

	if req.Https == "yes" {
		if err := SettingsRepo.Update("HttpsCertType", req.HttpsCertType); err != nil {
			return err
		}

		if req.HttpsCertType == "path" {
			if err := SettingsRepo.Update("HttpsCertPath", req.HttpsCertPath); err != nil {
				return err
			}

			if err := SettingsRepo.Update("HttpsKeyPath", req.HttpsKeyPath); err != nil {
				return err
			}
		}
	}

	return nil
}
