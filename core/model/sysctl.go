package model

type OperateServiceReq struct {
	Service   string `json:"service"`
	Operation string `json:"operation" validate:"required,oneof=start,stop,restart,enable,disable,reload,status"`
}
