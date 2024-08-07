package sysinfo

import (
	"encoding/json"
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

func (s *SysInfo) getNetwork() (model.NetworkInfo, error) {
	network := model.NetworkInfo{}

	actionRequest := model.HostAction{
		HostID: 1,
		Action: model.Action{
			Action: model.Action_SysInfo_Network,
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
		return network, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return network, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("network result: %v", actionResponse)

	err = json.Unmarshal([]byte(actionResponse.Data.Action.Data), &network)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to Network: %v", err)
	}

	return network, nil
}
