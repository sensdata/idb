package sysinfo

import (
	"encoding/json"
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

func (s *SysInfo) getSystemInfo() (model.SystemInfo, error) {
	systemInfo := model.SystemInfo{}

	actionRequest := model.HostAction{
		HostID: 1,
		Action: model.Action{
			Action: model.Action_SysInfo_System,
			Data:   "",
		},
	}

	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/idb/api/act/send")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return systemInfo, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return systemInfo, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("system result: %v", actionResponse)

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &systemInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to SystemInfo: %v", err)
	}

	return systemInfo, nil
}
