package sysinfo

import (
	"encoding/json"

	"github.com/sensdata/idb/core/model"
)

func (s *SysInfo) getHardware(hostID uint) (*model.HardwareInfo, error) {
	hardwareInfo := model.HardwareInfo{}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.Sysinfo_Hardware,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &hardwareInfo, err
	}

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &hardwareInfo)
	if err != nil {
		LOG.Error("Error unmarshaling data to HardwareInfo: %v", err)
	}

	return &hardwareInfo, nil
}
