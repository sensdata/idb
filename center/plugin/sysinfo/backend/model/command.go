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

type CommandResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    CommandResult `json:"data"`
}

type CommandGroupResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    CommandGroupResult `json:"data"`
}
