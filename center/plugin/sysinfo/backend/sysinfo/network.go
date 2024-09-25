package sysinfo

import (
	"encoding/json"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

func (s *SysInfo) getNetwork() (*model.NetworkInfo, error) {
	network := model.NetworkInfo{}

	actionRequest := model.HostAction{
		HostID: 1,
		Action: model.Action{
			Action: model.SysInfo_Network,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &network, err
	}

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &network)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to Network: %v", err)
	}

	return &network, nil
}
