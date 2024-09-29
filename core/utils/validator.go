package utils

import (
	"fmt"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func InitValidator() *validator.Validate {
	validator := validator.New()
	if err := validator.RegisterValidation("name", checkNamePattern); err != nil {
		panic(err)
	}
	if err := validator.RegisterValidation("ip", checkIpPattern); err != nil {
		panic(err)
	}
	if err := validator.RegisterValidation("password", checkPasswordPattern); err != nil {
		panic(err)
	}
	// if err := validator.RegisterValidation("customRequired", customRequired); err != nil {
	// 	panic(err)
	// }

	return validator
}

// func customRequired(fl validator.FieldLevel) bool {
// 	field := fl.Field()

// 	// 对于int类型，允许0
// 	if field.Kind() == reflect.Int || field.Kind() == reflect.Int16 || field.Kind() == reflect.Int32 || field.Kind() == reflect.Int64 {
// 		return true // 允许任何整数，包括0
// 	}

// 	// 对于bool类型，允许false
// 	if field.Kind() == reflect.Bool {
// 		return true // 允许任何布尔值，包括false
// 	}

// 	// 其他类型可以根据需要处理或返回true
// 	return field.IsValid() // 返回字段是否有效，允许通过
// }

func checkNamePattern(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	result, err := regexp.MatchString("^[a-zA-Z\u4e00-\u9fa5]{1}[a-zA-Z0-9_\u4e00-\u9fa5]{0,30}$", value)
	if err != nil {
		fmt.Printf("regexp matchString failed, %v", err)
	}
	return result
}

func checkIpPattern(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	result, err := regexp.MatchString(`^((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}$`, value)
	if err != nil {
		fmt.Printf("regexp check ip matchString failed, %v", err)
	}
	return result
}

func checkPasswordPattern(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) < 8 || len(value) > 30 {
		return false
	}

	hasNum := false
	hasLetter := false
	for _, r := range value {
		if unicode.IsLetter(r) && !hasLetter {
			hasLetter = true
		}
		if unicode.IsNumber(r) && !hasNum {
			hasNum = true
		}
		if hasLetter && hasNum {
			return true
		}
	}

	return false
}
