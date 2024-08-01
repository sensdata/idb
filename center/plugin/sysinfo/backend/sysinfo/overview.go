package sysinfo

import (
	"fmt"
	"strings"

	"github.com/sensdata/idb/center/global"
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
			"date +\"%Y-%m-%d %H:%M:%S\"", //服务器时间
		},
	}

	var commandGroupResponse model.CommandGroupResponse

	resp, err := s.restyClient.R().
		SetBody(command).
		SetResult(&commandGroupResponse).
		Post("http://127.0.0.1:8080/idb/api/cmd/send/group")

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return overview, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("failed to send request: %v", err)
		return overview, fmt.Errorf("received error response: %s", resp.Status())
	}

	global.LOG.Info("overview result: %v", commandGroupResponse)

	overviewHandlers := []OverviewResultHandler{
		{Description: "Server time", Handler: s.handlerServerTime},
	}

	for i, result := range commandGroupResponse.Data.Results {
		if i < len(overviewHandlers) {
			handler := overviewHandlers[i]
			handler.Handler(&overview, result)
		} else {
			fmt.Println("Unknown result at index", i, ":", result)
		}
	}

	return overview, nil
}

func (s *SysInfo) handlerServerTime(overview *model.Overview, result string) {
	overview.ServerTime = strings.TrimSpace(result)
}
