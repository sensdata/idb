package plugin

type PluginRequest struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PluginResponse struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Error   string      `json:"error,omitempty"`
}
