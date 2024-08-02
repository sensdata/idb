package sysinfo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
			"date +\"%Y-%m-%d %H:%M:%S\"",                         //服务器时间
			"timedatectl | grep \"Time zone\" | awk '{print $3}'", //当前时区
			"uptime -s",                  //启动时间
			"uptime -p | sed 's/^up //'", //运行时长
			"grep '^cpu ' /proc/stat | awk '{print $5}'",         //空闲时长
			"top -bn1 | grep \"Cpu(s)\" | awk '{print $2 + $4}'", //CPU使用率
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
		{Description: "Time zone", Handler: s.handlerTimeZone},
		{Description: "Boot time", Handler: s.handlerBootTime},
		{Description: "Run time", Handler: s.handlerRunTime},
		{Description: "Idle time", Handler: s.handlerIdleTime},
		{Description: "Cpu Usage", Handler: s.handlerCpuUsage},
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

func (s *SysInfo) handlerTimeZone(overview *model.Overview, result string) {
	overview.ServerTimeZone = strings.TrimSpace(result)
}

func (s *SysInfo) handlerBootTime(overview *model.Overview, result string) {
	overview.BootTime = strings.TrimSpace(result)
}

func (s *SysInfo) handlerRunTime(overview *model.Overview, result string) {
	overview.RunTime = strings.TrimSpace(result)
}

func (s *SysInfo) handlerIdleTime(overview *model.Overview, result string) {
	idleJiffies, err := strconv.Atoi(strings.TrimSpace(result))
	if err != nil {
		fmt.Printf("Error parsing idle jiffies: %v\n", err)
		overview.IdleTime = "Unknown"
		return
	}

	// 获取系统时钟频率（通常是每秒100 jiffies）
	const jiffiesPerSecond = 100

	// 将 jiffies 转换为秒
	idleSeconds := idleJiffies / jiffiesPerSecond

	// 将秒转换为天、小时、分钟、秒
	idleDuration := time.Duration(idleSeconds) * time.Second
	days := idleDuration / (24 * time.Hour)
	hours := (idleDuration % (24 * time.Hour)) / time.Hour
	minutes := (idleDuration % time.Hour) / time.Minute
	seconds := (idleDuration % time.Minute) / time.Second

	// 格式化输出
	idelTime := fmt.Sprintf("%d days %d hours %d minutes %d seconds", days, hours, minutes, seconds)
	overview.IdleTime = idelTime
}

func (s *SysInfo) handlerCpuUsage(overview *model.Overview, result string) {
	cpuUsage := fmt.Sprintf("%s%%", strings.TrimSpace(result))
	overview.CpuUsage = cpuUsage
}
