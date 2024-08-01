package sysinfo

import (
	"fmt"

	"github.com/sensdata/idb/center/plugin/sysinfo/backend/model"
)

type OverviewResultHandler struct {
	Description string
	Handler     func(*model.Overview, string)
}

func (s *SysInfo) getOverview() (model.Overview, error) {
	overview := model.Overview{}

	command := model.CommandGroup{
		HostID: 1,
		Commands: []string{
			"date " + "%Y-%m-%d %H:%M:%S", //服务器时间
		},
	}

	var commandGroupResult model.CommandGroupResult

	resp, err := s.restyClient.R().
		SetBody(command).
		SetResult(&commandGroupResult).
		Post("http://127.0.0.1:8080/cmd/send/group")

	if err != nil {
		return overview, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.IsError() {
		return overview, fmt.Errorf("received error response: %s", resp.Status())
	}

	overviewHandlers := []OverviewResultHandler{
		{Description: "Server time", Handler: handlerServerTime},
	}

	for i, result := range commandGroupResult.Results {
		if i < len(overviewHandlers) {
			handler := overviewHandlers[i]
			handler.Handler(&overview, result)
		} else {
			fmt.Println("Unknown result at index", i, ":", result)
		}
	}

	return overview, nil
}

func handlerServerTime(overview *model.Overview, result string) {
	overview.ServerTime = result
}
