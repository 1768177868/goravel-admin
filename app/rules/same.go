package rules

import (
	"github.com/goravel/framework/contracts/validation"
)

/**
 * same 验证一个字段的值必须与另一个字段的值相同
 * same verify a field value must be the same as another field value
 * 用法：same:字段名称
 * Usage: same:field_name
 * 例子：same:new_password
 * Example: same:new_password
 */

type Same struct {
}

// Signature The name of the rule.
func (receiver *Same) Signature() string {
	return "same"
}

// Passes Determine if the validation rule passes.
func (receiver *Same) Passes(data validation.Data, val any, options ...any) bool {
	if len(options) == 0 {
		return false
	}

	fieldName := options[0].(string)
	compareValue, exist := data.Get(fieldName)
	if !exist {
		return false
	}

	// 转换为字符串进行比较
	valStr := ""
	compareStr := ""

	if strVal, ok := val.(string); ok {
		valStr = strVal
	} else {
		return false
	}

	if strCompareVal, ok := compareValue.(string); ok {
		compareStr = strCompareVal
	} else {
		return false
	}

	return valStr == compareStr
}

// Message Get the validation error message.
func (receiver *Same) Message() string {
	return "The :attribute must match :other."
}

