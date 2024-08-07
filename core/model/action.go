package model

const (
	Action_SysInfo_OverView string = "action_sysinfo_overview"
	Action_SysInfo_Network  string = "action_sysinfo_network"
	Action_SysInfo_System   string = "action_sysinfo_system"
)

// Action消息结构
type Action struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

type HostAction struct {
	HostID uint   `json:"hostId"`
	Action Action `json:"action"`
}

type ActionResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    HostAction `json:"data"`
}
