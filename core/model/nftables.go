package model

type ToggleOptions struct {
	Option string `json:"option" validate:"required,oneof=on off"`
}

type SwitchOptions struct {
	Option string `json:"option" validate:"required,oneof=nftables iptables"`
}

type ProcessStatus struct {
	Process   string   `json:"process"`
	Pid       int      `json:"pid"`
	Port      int      `json:"port"`
	Addresses []string `json:"addresses"`
	Status    string   `json:"status"`
}
