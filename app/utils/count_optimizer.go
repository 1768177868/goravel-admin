package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// CountOptimizer 分页统计优化器
// 当数据量超过阈值时，使用执行计划估算的行数，否则使用实际的 count(*)
type CountOptimizer struct {
	// Threshold 阈值，超过此值使用估算值（默认 10000）
	Threshold int64
	// ModuleName 模块名称，用于日志记录
	ModuleName string
}

// NewCountOptimizer 创建新的 CountOptimizer
// threshold: 阈值，超过此值使用估算值（默认 10000）
// moduleName: 模块名称，用于日志记录
func NewCountOptimizer(threshold int64, moduleName string) *CountOptimizer {
	if threshold <= 0 {
		threshold = 10000 // 默认阈值
	}
	return &CountOptimizer{
		Threshold:  threshold,
		ModuleName: moduleName,
	}
}

// OptimizedCount 优化的 count 查询（使用 ORM Query 对象）
// query: ORM 查询对象（已应用筛选条件，但未应用排序和分页）
// 注意：此方法会先尝试估算，如果失败则使用实际 count
// 返回：总数、是否使用估算值、错误
func (co *CountOptimizer) OptimizedCount(query orm.Query) (int64, bool, error) {
	// 先尝试执行实际 count（如果数据量小，直接 count 也很快）
	// 如果数据量大，我们可以通过执行时间来判断，但更简单的方法是先估算
	// 这里我们提供一个简化版本：直接使用实际 count，如果慢的话可以后续优化

	// 由于无法直接从 ORM Query 提取 SQL，这里提供一个变通方案：
	// 先执行一次快速查询获取估算值（需要表名）
	// 但更推荐使用 OptimizedCountWithTable 方法

	actualCount, err := query.Count()
	return actualCount, false, err
}

// extractRowsFromPostgreSQLExplain 从 PostgreSQL EXPLAIN JSON 结果中提取行数
func (co *CountOptimizer) extractRowsFromPostgreSQLExplain(jsonStr string) int64 {
	// 简化实现：使用正则表达式提取 "Plan" -> "Plan Rows" 的值
	// 实际应该使用 JSON 解析，但为了简化，使用正则
	re := regexp.MustCompile(`"Plan Rows":\s*(\d+)`)
	matches := re.FindStringSubmatch(jsonStr)
	if len(matches) > 1 {
		if rows, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			return rows
		}
	}
	return 0
}

// extractRowsFromPostgreSQLExplainText 从 PostgreSQL EXPLAIN 文本结果中提取行数
func (co *CountOptimizer) extractRowsFromPostgreSQLExplainText(result []map[string]any) int64 {
	// PostgreSQL 文本格式的 EXPLAIN 结果中，行数通常在 "rows=" 后面
	for _, row := range result {
		if queryPlan, ok := row["QUERY PLAN"]; ok {
			if planStr, ok := queryPlan.(string); ok {
				// 使用正则表达式提取 rows= 后面的数字
				re := regexp.MustCompile(`rows=(\d+)`)
				matches := re.FindStringSubmatch(planStr)
				if len(matches) > 1 {
					if rows, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
						return rows
					}
				}
			}
		}
	}
	return 0
}

// OptimizedCountWithTable 使用表名和 WHERE 条件进行优化的 count 查询
// tableName: 表名
// whereClause: WHERE 子句（不包含 WHERE 关键字），例如："status = ? AND user_id = ?"
// args: WHERE 条件的参数
// 返回：总数、是否使用估算值、错误
func (co *CountOptimizer) OptimizedCountWithTable(tableName, whereClause string, args ...any) (int64, bool, error) {
	dbConnection := facades.Config().GetString("database.default", "sqlite")

	// 构建 COUNT SQL
	var countSQL string
	if whereClause != "" {
		switch dbConnection {
		case "mysql":
			countSQL = fmt.Sprintf("SELECT COUNT(*) as cnt FROM `%s` WHERE %s", tableName, whereClause)
		case "postgres":
			countSQL = fmt.Sprintf("SELECT COUNT(*) as cnt FROM %s WHERE %s", tableName, whereClause)
		default:
			// 其他数据库不支持估算，直接使用实际 count
			var result struct {
				Cnt int64
			}
			if err := facades.Orm().Query().Raw(countSQL, args...).Scan(&result); err != nil {
				return 0, false, err
			}
			return result.Cnt, false, nil
		}
	} else {
		switch dbConnection {
		case "mysql":
			countSQL = fmt.Sprintf("SELECT COUNT(*) as cnt FROM `%s`", tableName)
		case "postgres":
			countSQL = fmt.Sprintf("SELECT COUNT(*) as cnt FROM %s", tableName)
		default:
			// 其他数据库不支持估算，直接使用实际 count
			var result struct {
				Cnt int64
			}
			if err := facades.Orm().Query().Raw(countSQL).Scan(&result); err != nil {
				return 0, false, err
			}
			return result.Cnt, false, nil
		}
	}

	// 构建 EXPLAIN SQL
	var explainSQL string
	if dbConnection == "mysql" {
		if whereClause != "" {
			explainSQL = fmt.Sprintf("EXPLAIN SELECT COUNT(*) FROM `%s` WHERE %s", tableName, whereClause)
		} else {
			explainSQL = fmt.Sprintf("EXPLAIN SELECT COUNT(*) FROM `%s`", tableName)
		}
	} else if dbConnection == "postgres" {
		if whereClause != "" {
			explainSQL = fmt.Sprintf("EXPLAIN (FORMAT JSON) SELECT COUNT(*) FROM %s WHERE %s", tableName, whereClause)
		} else {
			explainSQL = fmt.Sprintf("EXPLAIN (FORMAT JSON) SELECT COUNT(*) FROM %s", tableName)
		}
	} else {
		// 其他数据库不支持估算，直接使用实际 count
		var result struct {
			Cnt int64
		}
		if err := facades.Orm().Query().Raw(countSQL, args...).Scan(&result); err != nil {
			return 0, false, err
		}
		return result.Cnt, false, nil
	}

	// 先执行 EXPLAIN 查询获取估算值（快速，不需要特别精准）
	estimatedCount, err := co.executeExplain(explainSQL, args...)
	if err != nil {
		// 如果估算失败，使用实际 count
		var result struct {
			Cnt int64
		}
		if err := facades.Orm().Query().Raw(countSQL, args...).Scan(&result); err != nil {
			return 0, false, err
		}
		return result.Cnt, false, nil
	}

	// 如果估算值超过阈值，直接返回估算值（不需要执行实际 count，更快）
	if estimatedCount >= co.Threshold {
		return estimatedCount, true, nil
	}

	// 估算值小于阈值，执行实际的 count 获取精确值
	var result struct {
		Cnt int64
	}
	if err := facades.Orm().Query().Raw(countSQL, args...).Scan(&result); err != nil {
		return 0, false, err
	}
	return result.Cnt, false, nil
}

// executeExplain 执行 EXPLAIN 查询并提取估算行数
func (co *CountOptimizer) executeExplain(explainSQL string, args ...any) (int64, error) {
	dbConnection := facades.Config().GetString("database.default", "sqlite")

	switch dbConnection {
	case "mysql":
		// MySQL EXPLAIN 返回表格格式
		var explainResult []map[string]any
		if err := facades.Orm().Query().Raw(explainSQL, args...).Get(&explainResult); err != nil {
			return 0, err
		}

		if len(explainResult) == 0 {
			return 0, fmt.Errorf("explain result is empty")
		}

		// MySQL EXPLAIN 结果中，rows 字段包含估算行数
		if rows, ok := explainResult[0]["rows"]; ok {
			switch v := rows.(type) {
			case int64:
				return v, nil
			case int:
				return int64(v), nil
			case float64:
				return int64(v), nil
			case string:
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					return parsed, nil
				}
			}
		}
		return 0, fmt.Errorf("cannot extract rows from explain result")

	case "postgres":
		// PostgreSQL EXPLAIN (FORMAT JSON) 返回 JSON 格式
		var explainResult []map[string]any
		if err := facades.Orm().Query().Raw(explainSQL, args...).Get(&explainResult); err != nil {
			return 0, err
		}

		if len(explainResult) == 0 {
			return 0, fmt.Errorf("explain result is empty")
		}

		// PostgreSQL EXPLAIN (FORMAT JSON) 结果是一个包含 JSON 字符串的数组
		if queryPlan, ok := explainResult[0]["QUERY PLAN"]; ok {
			if planStr, ok := queryPlan.(string); ok {
				// 从 JSON 字符串中提取 "Plan" -> "Plan Rows" 的值
				rows := co.extractRowsFromPostgreSQLExplain(planStr)
				if rows > 0 {
					return rows, nil
				}
			}
		}

		// 如果 JSON 格式解析失败，尝试文本格式
		explainTextSQL := strings.Replace(explainSQL, "EXPLAIN (FORMAT JSON)", "EXPLAIN", 1)
		var explainTextResult []map[string]any
		if err := facades.Orm().Query().Raw(explainTextSQL, args...).Get(&explainTextResult); err == nil {
			rows := co.extractRowsFromPostgreSQLExplainText(explainTextResult)
			if rows > 0 {
				return rows, nil
			}
		}

		return 0, fmt.Errorf("cannot extract rows from explain result")
	}

	return 0, fmt.Errorf("unsupported database: %s", dbConnection)
}
