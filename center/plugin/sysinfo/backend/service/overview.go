package service

import (
	"fmt"

	"github.com/sensdata/idb/center/plugin/sysinfo/backend/model"
)

func getOverview() (model.Overview, error) {
	overview := model.Overview{}

	//服务器时间
	serverTime, err := getServiceTime(1)
	if err != nil {
		fmt.Println(fmt.Errorf("get service time error %w", err))
	} else {
		overview.ServerTime = serverTime
	}

	return overview, nil
}

func getServiceTime(hostId uint) (string, error) {
	command := model.Command{
		HostID:  hostId,
		Command: "date " + "%Y-%m-%d %H:%M:%S",
	}

	var commandResult model.CommandResult

	resp, err := restyClient.R().
		SetBody(command).
		SetResult(&commandResult).
		Post("http://127.0.0.1:8080/cmd/send")

	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("received error response: %s", resp.Status())
	}

	return commandResult.Result, nil
}
