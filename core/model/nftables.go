package model

type NftablesStatus struct {
	Status string `json:"status"`
	Active string `json:"active"`
}

type ToggleOptions struct {
	Option string `json:"option" validate:"required,oneof=on off"`
}

type SwitchOptions struct {
	Option string `json:"option" validate:"required,oneof=nftables iptables"`
}

type PortAccessStatus struct {
	Address string `json:"address"`
	Status  string `json:"status" validate:"required,oneof=accepted rejected local-only restricted fully-accepted unknown"`
}

type ProcessStatus struct {
	Process string             `json:"process"`
	Pid     int                `json:"pid"`
	Port    int                `json:"port"`
	Access  []PortAccessStatus `json:"access"`
}

const (
	RuleRateLimit       string = "rate_limit"
	RuleConcurrentLimit string = "concurrent_limit"
	RuleDefault         string = "default"
)

type BaseRules struct {
	InputPolicy string `json:"input_policy" validate:"required,oneof=drop accept reject"`
}

type RuleItem struct {
	Type   string `json:"type" validate:"required,oneof=rate_limit concurrent_limit default"`
	Rate   string `json:"rate,omitempty"`
	Count  int    `json:"count,omitempty"`
	Action string `json:"action" validate:"required,oneof=accept reject drop"`
}

type PortRule struct {
	Protocol string     `json:"protocol"`
	Port     int        `json:"port"`
	Rules    []RuleItem `json:"rules"`
}

type SetPortRule struct {
	Port  int        `json:"port"`
	Rules []RuleItem `json:"rules"`
}

type IPRequest struct {
	IP string `json:"ip" validate:"required"`
}

type PingStatus struct {
	Allowed bool `json:"allowed"`
}

type ConfRaw struct {
	Content string `json:"content" validate:"required"`
}
