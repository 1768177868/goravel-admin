package helpers

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
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
	if idStr == "" {
		return []uint{}
	}

	var ids []uint
	idMap := make(map[uint]bool)

	// 分割字符串
	idStrs := strings.Split(idStr, ",")
	for _, idStr := range idStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
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

// ConvertNumericToString 将数字类型转换为字符串
// 用于 PrepareForValidation 中，将数字字段转换为字符串以便 in 规则能正确验证
// 支持所有常见的数字类型：int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
func ConvertNumericToString(val any) string {
	switch v := val.(type) {
	case float64:
		// JSON 数字会被解析为 float64
		return strconv.FormatInt(int64(v), 10)
	case float32:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case string:
		return v
	default:
		return ""
	}
}

// PrepareNumericFieldForValidation 在 PrepareForValidation 中准备数字字段
// 将指定的数字字段转换为字符串，以便 in 规则能正确验证
// 用法：在 PrepareForValidation 方法中调用此函数处理需要 in 验证的数字字段
// 示例：return PrepareNumericFieldForValidation(data, "status")
func PrepareNumericFieldForValidation(data validation.Data, fieldName string) error {
	if val, exist := data.Get(fieldName); exist {
		statusStr := ConvertNumericToString(val)
		return data.Set(fieldName, statusStr)
	}
	return nil
}
