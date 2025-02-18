package sysinfo

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *SysInfo) setTime(hostID uint, req model.SetTimeReq) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.SysInfo_Set_Time,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to set time")
		return fmt.Errorf("failed to set time")
	}

	return nil
}

func (s *SysInfo) setTimeZone(hostID uint, req model.SetTimezoneReq) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.SysInfo_Set_Time_Zone,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to set timezone")
		return fmt.Errorf("failed to set timezone")
	}
	return nil
}

func (s *SysInfo) syncTime(hostID uint) error {
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.SysInfo_Sync_Time,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to sync time")
		return fmt.Errorf("failed to sync time")
	}
	return nil
}

func (s *SysInfo) clearMemCache(hostID uint) error {
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Sysinfo_Clear_Mem_Cache,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to clear mem cache")
		return fmt.Errorf("failed to clear mem cache")
	}
	return nil
}

func (s *SysInfo) setAutoClearInterval(hostID uint, req model.AutoClearMemCacheReq) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.SysInfo_Set_Auto_Clear,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to set auto clear interval")
		return fmt.Errorf("failed to set auto clear interval")
	}
	return nil
}

func (s *SysInfo) createSwap(hostID uint, req model.CreateSwapReq) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.SysInfo_Create_Swap,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to create swap")
		return fmt.Errorf("failed to create swap")
	}
	return nil
}

func (s *SysInfo) deleteSwap(hostID uint) error {
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Sysinfo_Delete_Swap,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to delete swap")
		return fmt.Errorf("failed to delete swap")
	}
	return nil
}

func (s *SysInfo) updateDNS(hostID uint, req model.UpdateDnsSettingsReq) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Sysinfo_Update_Dns,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to update dns")
		return fmt.Errorf("failed to update dns")
	}
	return nil
}

func (s *SysInfo) getSystemSettings(hostID uint) (*model.SystemSettings, error) {
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Sysinfo_Get_Sys_Setting,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return nil, err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to get system settings")
		return nil, fmt.Errorf("failed to get system settings")
	}
	var settings model.SystemSettings
	err = utils.FromJSONString(actionResponse.Data.Action.Data, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *SysInfo) updateSystemSettings(hostID uint, req model.UpdateSystemSettingsReq) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}
	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Sysinfo_Upd_Sys_Setting,
			Data:   data,
		},
	}
	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}
	if !actionResponse.Data.Action.Result {
		global.LOG.Error("failed to update system settings")
		return fmt.Errorf("failed to update system settings")
	}
	return nil
}
