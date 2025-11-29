package helpers

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
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
