package rules

import (
	"context"

	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/validation"
)

// NotExists 验证一个值在某个表的一个或多个字段中不存在。
// 用法：not_exists:表名称,字段名称[,字段名称...]
// 例子：not_exists:users,phone,email
type NotExists struct{}

func (receiver *NotExists) Signature() string {
	return "not_exists"
}

func (receiver *NotExists) Passes(ctx context.Context, _ validation.Data, val any, options ...any) bool {
	if len(options) < 2 {
		return false
	}

	tableName, ok := options[0].(string)
	if !ok || tableName == "" {
		return false
	}
	fieldName, ok := options[1].(string)
	if !ok || fieldName == "" {
		return false
	}

	requestValue, ok := val.(string)
	if !ok || requestValue == "" {
		return false
	}

	query := appfacades.OrmQuery(ctx).Table(tableName).Where(fieldName, requestValue)
	if len(options) > 2 {
		for i := 2; i < len(options); i++ {
			if col, ok := options[i].(string); ok && col != "" {
				query = query.OrWhere(col, requestValue)
			}
		}
	}

	count, err := query.Count()
	if err != nil {
		return false
	}

	return count == 0
}

func (receiver *NotExists) Message(ctx context.Context) string {
	return "record already exists"
}
