package model

type PageInfo struct {
	Page     int `form:"page" validate:"required"`
	PageSize int `form:"page_size" validate:"required"`
}

type SearchPageInfo struct {
	PageInfo
	Info    string `json:"info"`
	OrderBy string `json:"order_by"`
	Order   string `json:"order"`
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
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type KeyValueForUpdate struct {
	Key      string `json:"key" validate:"required"`
	OldValue string `json:"old_value" validate:"required"`
	NewValue string `json:"new_value" validate:"required"`
}

type Options struct {
	Option string `json:"option"`
}

type Rename struct {
	Name    string `json:"name" validate:"required"`
	NewName string `json:"new_name" validate:"required"`
}

type BatchDelete struct {
	Force bool     `json:"force"`
	Names []string `json:"names" validate:"required"`
}
