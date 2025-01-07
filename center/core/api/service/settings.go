package service

import (
	"encoding/json"
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
	About() (*model.About, error)
	Settings() (*model.SettingInfo, error)
	Update(req model.UpdateSettingRequest) error
}

func NewISettingsService() ISettingsService {
	return &SettingsService{}
}

func (s *SettingsService) About() (*model.About, error) {
	var about model.About

	about.Version = global.Version

	return &about, nil
}

func (s *SettingsService) Settings() (*model.SettingInfo, error) {
	setting, err := SettingsRepo.GetList()
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	settingMap := make(map[string]string)
	for _, set := range setting {
		settingMap[set.Key] = set.Value
	}
	var info model.SettingInfo
	arr, err := json.Marshal(settingMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}
	return &info, nil
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

	if err = s.updateMonitorIP(req.MonitorIP); err != nil {
		global.LOG.Error("Failed to save MonitorIP to %s: %v", req.MonitorIP, err)
		return err
	}

	if err = s.updateServerPort(req.ServerPort); err != nil {
		global.LOG.Error("Failed to save ServerPort to %d: %v", req.ServerPort, err)
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

func (s *SettingsService) updateMonitorIP(newIP string) error {
	if len(newIP) == 0 {
		return nil
	}
	return SettingsRepo.Update("MonitorIP", newIP)
}

func (s *SettingsService) updateServerPort(newPort int) error {
	if common.ScanPort(newPort) {
		return errors.New(constant.ErrPortInUsed)
	}

	_, err := SettingsRepo.Get(SettingsRepo.WithByKey("ServerPort"))
	if err != nil {
		return err
	}

	// TODO: 处理port的更换（调用nftables）

	return SettingsRepo.Update("ServerPort", strconv.Itoa(int(newPort)))
}

func (s *SettingsService) updateBindDomain(newDomain string) error {
	if len(newDomain) == 0 {
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
