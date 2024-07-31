package model

type Command struct {
	HostID  uint   `json:"hostId"`
	Command string `json:"command"`
}

type CommandResult struct {
	HostID uint   `json:"hostId"`
	Result string `json:"result"`
}
