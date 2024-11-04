package sysinfo

import (
	"encoding/json"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

func (s *SysInfo) getOverview(hostID uint) (*model.Overview, error) {
	overview := model.Overview{}

	actionRequest := model.HostAction{
		HostID: hostID,
		Action: model.Action{
			Action: model.SysInfo_OverView,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &overview, err
	}

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &overview)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to Overview: %v", err)
	}

	return &overview, nil
}
