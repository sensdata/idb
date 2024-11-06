package model

type PageInfo struct {
	Page     int `form:"page" validate:"required"`
	PageSize int `form:"page_size" validate:"required"`
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

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyValueForUpdate struct {
	Key      string `json:"key" validate:"required"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}
