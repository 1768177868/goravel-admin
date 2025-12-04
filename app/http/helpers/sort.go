package helpers

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/str"
)

// ApplySort 应用排序到查询
// orderBy: 排序参数，格式为 "field:direction" 或 "field1:direction1,field2:direction2"
// defaultSort: 默认排序，格式为 "field:direction"，如果 orderBy 为空则使用此默认值
// 返回: 应用了排序的查询对象
func ApplySort(query orm.Query, orderBy string, defaultSort string) orm.Query {
	// 如果提供了排序参数，使用它；否则使用默认排序
	sortStr := orderBy
	if sortStr == "" {
		sortStr = defaultSort
	}

	// 如果排序字符串为空，返回原查询
	if sortStr == "" {
		return query
	}

	// 解析多个排序字段（逗号分隔）
	sortFields := str.Of(sortStr).Split(",")
	var orderClauses []string
	
	for _, field := range sortFields {
		field = str.Of(field).Trim().String()
		if str.Of(field).IsEmpty() {
			continue
		}

		// 解析字段和方向（格式: "field:direction" 或 "field"）
		parts := str.Of(field).Split(":")
		fieldName := str.Of(parts[0]).Trim().String()
		direction := "asc" // 默认升序

		if len(parts) > 1 {
			direction = str.Of(parts[1]).Trim().Lower().String()
		}

		// 验证方向
		if direction != "asc" && direction != "desc" {
			direction = "asc"
		}

		// 收集排序子句
		orderClauses = append(orderClauses, fieldName+" "+direction)
	}

	// 如果有排序子句，组合成一个字符串并应用
	if len(orderClauses) > 0 {
		var orderStr string
		if len(orderClauses) > 0 {
			orderStr = orderClauses[0]
			for i := 1; i < len(orderClauses); i++ {
				orderStr = str.Of(orderStr).Append(", ").Append(orderClauses[i]).String()
			}
		}
		query = query.Order(orderStr)
	}

	return query
}

// ParseSort 解析排序参数
// orderBy: 排序参数，格式为 "field:direction" 或 "field1:direction1,field2:direction2"
// 返回: 排序字段和方向的映射
func ParseSort(orderBy string) map[string]string {
	result := make(map[string]string)
	
	if orderBy == "" {
		return result
	}

	// 解析多个排序字段（逗号分隔）
	sortFields := str.Of(orderBy).Split(",")
	
	for _, field := range sortFields {
		field = str.Of(field).Trim().String()
		if str.Of(field).IsEmpty() {
			continue
		}

		// 解析字段和方向
		parts := str.Of(field).Split(":")
		fieldName := str.Of(parts[0]).Trim().String()
		direction := "asc" // 默认升序

		if len(parts) > 1 {
			direction = str.Of(parts[1]).Trim().Lower().String()
		}

		// 验证方向
		if direction != "asc" && direction != "desc" {
			direction = "asc"
		}

		result[fieldName] = direction
	}

	return result
}

