package model

type Command struct {
	HostID  uint   `json:"hostId"`
	Command string `json:"command"`
}

type CommandResult struct {
	HostID uint   `json:"hostId"`
	Result string `json:"result"`
}

type CommandGroup struct {
	HostID   uint     `json:"hostId"`
	Commands []string `json:"commands"`
}

type CommandGroupResult struct {
	HostID  uint     `json:"hostId"`
	Results []string `json:"results"`
}
