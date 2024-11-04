package sysinfo

import (
	"encoding/json"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

func (s *SysInfo) getSystemInfo(hostID uint) (*model.SystemInfo, error) {
	systemInfo := model.SystemInfo{}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.SysInfo_System,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &systemInfo, err
	}

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &systemInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to SystemInfo: %v", err)
	}

	return &systemInfo, nil
}
