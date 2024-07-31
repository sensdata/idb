package dto

type Command struct {
	HostID  uint   `json:"hostId" validate:"required,number"`
	Command string `json:"command" validate:"required,string"`
}

type CommandResult struct {
	HostID uint   `json:"hostId"`
	Result string `json:"result"`
}
