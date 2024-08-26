package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSONString 将结构体转换为 JSON 字符串
func ToJSONString(v interface{}) (string, error) {
	jsonString, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(jsonString), nil
}

// FromJSONString 从 JSON 字符串解析出结构体
func FromJSONString(data string, v interface{}) error {
	if err := json.Unmarshal([]byte(data), v); err != nil {
		return err
	}
	return nil
}
