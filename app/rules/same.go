package rules

import (
	"context"

	"github.com/goravel/framework/contracts/validation"
)

// Same 验证一个字段的值必须与另一个字段的值相同。
// 用法：same:字段名称
// 例子：same:new_password
type Same struct{}

func (receiver *Same) Signature() string {
	return "same"
}

func (receiver *Same) Passes(_ context.Context, data validation.Data, val any, options ...any) bool {
	if len(options) == 0 {
		return false
	}

	fieldName, ok := options[0].(string)
	if !ok || fieldName == "" {
		return false
	}

	compareValue, exist := data.Get(fieldName)
	if !exist {
		return false
	}

	valStr, ok := val.(string)
	if !ok {
		return false
	}
	compareStr, ok := compareValue.(string)
	if !ok {
		return false
	}

	return valStr == compareStr
}

func (receiver *Same) Message(ctx context.Context) string {
	return "The :attribute must match :other."
}
