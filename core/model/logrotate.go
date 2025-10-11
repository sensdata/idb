package model

type LogrotateOperate struct {
	Type      string `json:"type" validate:"required"`
	Category  string `json:"category"`
	Name      string `json:"name" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=test execute"`
}
