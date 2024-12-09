package model

type ServiceFormField struct {
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Key        string      `json:"key"`
	Type       string      `json:"type"`
	Default    string      `json:"default"`
	Value      string      `json:"value"`
	Required   bool        `json:"required"`
	Hint       string      `json:"hint"`
	Options    []string    `json:"options,omitempty"`
	Validation *Validation `json:"validation,omitempty"`
}

type ServiceForm struct {
	Fields []ServiceFormField `json:"fields"`
}

type CreateServiceForm struct {
	Type     string     `json:"type" validate:"required"`
	Category string     `json:"category"`
	Name     string     `json:"name" validate:"required"`
	Form     []KeyValue `json:"form" validate:"dive,required"`
}

type UpdateServiceForm struct {
	Type     string     `json:"type" validate:"required"`
	Category string     `json:"category"`
	Name     string     `json:"name" validate:"required"`
	Form     []KeyValue `json:"form" validate:"dive,required"`
}

type ServiceAction struct {
	Type     string `json:"type" validate:"required"`
	Category string `json:"category"`
	Name     string `json:"name" validate:"required"`
	Action   string `json:"action" validate:"required,oneof=activate deactivate"`
}

type ServiceActionInfo struct {
	RepoPath     string `json:"repo_path" validate:"required"`
	RelativePath string `json:"relative_path" validate:"required"`
	Action       string `json:"action" validate:"required,oneof=activate deactivate"`
}
