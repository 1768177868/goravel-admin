package helpers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/cast"
)

// GetIntQuery 获取并验证整数查询参数
// 如果参数无效或不存在，返回默认值
func GetIntQuery(ctx http.Context, key string, defaultValue int) int {
	value := ctx.Request().Query(key, "")
	if value == "" {
		return defaultValue
	}
	result := cast.ToInt(value)
	if result < 1 {
		return defaultValue
	}
	return result
}

// GetUintQuery 获取并验证无符号整数查询参数
// 如果参数无效或不存在，返回默认值
func GetUintQuery(ctx http.Context, key string, defaultValue uint) uint {
	value := ctx.Request().Query(key, "")
	if value == "" {
		return defaultValue
	}
	result := cast.ToUint(value)
	if result == 0 {
		return defaultValue
	}
	return result
}

// GetUintRoute 获取并验证路由中的无符号整数参数
// 如果参数无效或不存在，返回 0
func GetUintRoute(ctx http.Context, key string) uint {
	value := ctx.Request().Route(key)
	return cast.ToUint(value)
}

// ParseIDsFromString 从逗号分隔的字符串中解析 ID 列表
// 返回去重后的 ID 列表
func ParseIDsFromString(idStr string) []uint {
	if str.Of(idStr).IsEmpty() {
		return []uint{}
	}

	var ids []uint
	idMap := make(map[uint]bool)

	// 分割字符串
	idStrs := str.Of(idStr).Split(",")
	for _, idStr := range idStrs {
		idStr = str.Of(idStr).Trim().String()
		if str.Of(idStr).IsEmpty() {
			continue
		}

		id := cast.ToUint(idStr)
		if id > 0 && !idMap[id] {
			idMap[id] = true
			ids = append(ids, id)
		}
	}

	return ids
}

// ConvertUintSliceToAny 将 uint 切片转换为 []any
// 用于 ORM 的 WhereIn 查询
func ConvertUintSliceToAny(ids []uint) []any {
	if len(ids) == 0 {
		return []any{}
	}

	result := make([]any, len(ids))
	for i, id := range ids {
		result[i] = id
	}
	return result
}

// PrepareNumericFieldForValidation 在 PrepareForValidation 中准备数字字段
// 将指定的数字字段转换为字符串，以便 in 规则能正确验证
// 使用 cast.ToString 自动处理所有数字类型转换（int, int8-int64, uint, uint8-uint64, float32, float64）
// 用法：在 PrepareForValidation 方法中调用此函数处理需要 in 验证的数字字段
// 示例：return PrepareNumericFieldForValidation(data, "status")
func PrepareNumericFieldForValidation(data validation.Data, fieldName string) error {
	if val, exist := data.Get(fieldName); exist {
		return data.Set(fieldName, cast.ToString(val))
	}
	return nil
}
