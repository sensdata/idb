package dto

type PageInfo struct {
	Page     int `json:"page" validate:"required,number"`
	PageSize int `json:"pageSize" validate:"required,number"`
}

type PageResult struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
