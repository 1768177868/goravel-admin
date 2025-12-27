package utils

import (
	"fmt"
	"regexp"
	"time"

	"github.com/goravel/framework/facades"
)

// GetShardingTableName 根据时间获取分表名称
// baseTableName: 基础表名，如 "orders"
// orderTime: 订单时间
// 返回: 分表名称，如 "orders_202501"
func GetShardingTableName(baseTableName string, orderTime time.Time) string {
	return fmt.Sprintf("%s_%s", baseTableName, orderTime.Format("200601"))
}

// GetShardingTableNames 获取时间范围内的所有分表名称
// baseTableName: 基础表名
// startTime: 开始时间
// endTime: 结束时间
// 返回: 分表名称列表
func GetShardingTableNames(baseTableName string, startTime, endTime time.Time) []string {
	var tableNames []string

	// 确保开始时间不晚于结束时间
	if startTime.After(endTime) {
		return tableNames
	}

	// 从开始时间到结束时间，按月遍历
	current := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, startTime.Location())
	end := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, endTime.Location())

	for !current.After(end) {
		tableNames = append(tableNames, GetShardingTableName(baseTableName, current))
		current = current.AddDate(0, 1, 0) // 加一个月
	}

	return tableNames
}

// DefaultMaxTimeRangeMonths 默认最大时间范围（月数）
// 可以通过配置覆盖，用于限制查询时间范围，避免跨太多分表
const DefaultMaxTimeRangeMonths = 3

// TimeRangeError 时间范围验证错误
type TimeRangeError struct {
	Key    string
	Params map[string]any
}

func (e *TimeRangeError) Error() string {
	// 返回翻译键，由调用方进行翻译
	return e.Key
}

// ValidateTimeRange 验证时间范围是否超过指定月数
// startTime: 开始时间
// endTime: 结束时间
// maxMonths: 最大允许的月数，如果为0则使用默认值 DefaultMaxTimeRangeMonths
// 返回: 是否有效，错误信息（错误信息包含翻译键和参数）
func ValidateTimeRange(startTime, endTime time.Time, maxMonths ...int) (bool, error) {
	// 检查开始时间是否晚于结束时间
	if startTime.After(endTime) {
		return false, &TimeRangeError{
			Key:    "start_time_after_end_time",
			Params: nil,
		}
	}

	// 确定最大月数
	maxMonthsValue := DefaultMaxTimeRangeMonths
	if len(maxMonths) > 0 && maxMonths[0] > 0 {
		maxMonthsValue = maxMonths[0]
	}

	// 检查时间范围是否超过指定月数
	maxTimeLater := startTime.AddDate(0, maxMonthsValue, 0)
	if endTime.After(maxTimeLater) {
		return false, &TimeRangeError{
			Key: "time_range_exceeded",
			Params: map[string]any{
				"months": maxMonthsValue,
			},
		}
	}

	return true, nil
}

// GetAllExistingShardingTables 获取数据库中所有已存在的分表名称
// baseTableName: 基础表名，如 "orders" 或 "order_details"
// 返回: 已存在的分表名称列表
func GetAllExistingShardingTables(baseTableName string) ([]string, error) {
	var tableNames []string

	// 获取当前数据库名
	dbName := facades.Config().GetString("database.connections.mysql.database")
	if dbName == "" {
		dbName = facades.Config().GetString("database.connections.postgresql.database")
	}

	// 构建表名匹配模式：orders_YYYYMM 或 order_details_YYYYMM
	pattern := fmt.Sprintf("%s_%%", baseTableName)

	// 查询所有匹配的表名
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = ? 
		AND table_name LIKE ?
		ORDER BY table_name
	`

	// 执行查询，使用 Scan 获取结果
	var rows []map[string]any
	if err := facades.Orm().Query().Raw(query, dbName, pattern).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询分表失败: %v", err)
	}

	// 验证表名格式（确保是有效的分表名称，格式为 baseTableName_YYYYMM）
	patternRegex := regexp.MustCompile(fmt.Sprintf("^%s_\\d{6}$", regexp.QuoteMeta(baseTableName)))

	for _, row := range rows {
		// 尝试不同的字段名格式（MySQL 可能返回不同的大小写）
		var tableName string
		var ok bool

		// 尝试 table_name (小写)
		if tableName, ok = row["table_name"].(string); !ok {
			// 尝试 TABLE_NAME (大写)
			if tableName, ok = row["TABLE_NAME"].(string); !ok {
				// 尝试遍历所有键
				for key, value := range row {
					if (key == "table_name" || key == "TABLE_NAME" || key == "Table_Name") && value != nil {
						if str, ok := value.(string); ok {
							tableName = str
							ok = true
							break
						}
					}
				}
			}
		}

		if ok && tableName != "" {
			// 验证表名格式
			if patternRegex.MatchString(tableName) {
				tableNames = append(tableNames, tableName)
			}
		}
	}

	return tableNames, nil
}

// GetAllExistingShardingTablesByPattern 通过表名模式获取所有已存在的分表
// 这是一个更通用的方法，可以通过自定义模式匹配
// pattern: 表名匹配模式，如 "orders_%" 或 "order_details_%"
func GetAllExistingShardingTablesByPattern(pattern string) ([]string, error) {
	var tableNames []string

	// 获取当前数据库名
	dbName := facades.Config().GetString("database.connections.mysql.database")
	if dbName == "" {
		dbName = facades.Config().GetString("database.connections.postgresql.database")
	}

	// 查询所有匹配的表名
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = ? 
		AND table_name LIKE ?
		ORDER BY table_name
	`

	// 执行查询，使用 Scan 获取结果
	var rows []map[string]any
	if err := facades.Orm().Query().Raw(query, dbName, pattern).Scan(&rows); err != nil {
		return nil, fmt.Errorf("查询分表失败: %v", err)
	}

	for _, row := range rows {
		if tableName, ok := row["table_name"].(string); ok {
			tableNames = append(tableNames, tableName)
		}
	}

	return tableNames, nil
}
