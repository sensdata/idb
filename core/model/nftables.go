package model

type ToggleOptions struct {
	Option string `json:"option" validate:"required,oneof=on off"`
}

type SwitchOptions struct {
	Option string `json:"option" validate:"required,oneof=nftables iptables"`
}
