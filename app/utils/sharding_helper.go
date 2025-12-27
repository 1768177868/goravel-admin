package utils

import (
	"fmt"
	"time"
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

// ValidateTimeRange 验证时间范围是否超过指定月数
// startTime: 开始时间
// endTime: 结束时间
// maxMonths: 最大允许的月数，如果为0则使用默认值 DefaultMaxTimeRangeMonths
// 返回: 是否有效，错误信息
func ValidateTimeRange(startTime, endTime time.Time, maxMonths ...int) (bool, error) {
	// 检查开始时间是否晚于结束时间
	if startTime.After(endTime) {
		return false, fmt.Errorf("开始时间不能晚于结束时间")
	}

	// 确定最大月数
	maxMonthsValue := DefaultMaxTimeRangeMonths
	if len(maxMonths) > 0 && maxMonths[0] > 0 {
		maxMonthsValue = maxMonths[0]
	}

	// 检查时间范围是否超过指定月数
	maxTimeLater := startTime.AddDate(0, maxMonthsValue, 0)
	if endTime.After(maxTimeLater) {
		return false, fmt.Errorf("查询时间范围不能超过%d个月", maxMonthsValue)
	}

	return true, nil
}
