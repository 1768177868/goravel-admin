package utils

import (
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"
)

// DefaultMaxTimeRangeMonths 默认最大查询时间跨度（月），与 config sharding.max_time_range_months 一致。
const DefaultMaxTimeRangeMonths = 3

// GetTimeShardingSuffixLayout 时间分表后缀格式，默认 200601（按月）。
func GetTimeShardingSuffixLayout() string {
	layout := facades.Config().GetString("sharding.time_suffix_layout", "200601")
	if layout == "" {
		return "200601"
	}
	return layout
}

// GetMaxTimeRangeMonths 列表/导出允许的最大时间跨度（月）。
func GetMaxTimeRangeMonths() int {
	months := cast.ToInt(facades.Config().Get("sharding.max_time_range_months", DefaultMaxTimeRangeMonths))
	if months <= 0 {
		return DefaultMaxTimeRangeMonths
	}
	return months
}

// GetIDLookupScanMonths 仅按 ID 反查分表时向前扫描的月数。
func GetIDLookupScanMonths() int {
	months := cast.ToInt(facades.Config().Get("sharding.id_lookup_scan_months", 6))
	if months <= 0 {
		return 6
	}
	return months
}

// GetUserBalanceLogsShards 用户余额变动记录哈希分表数量。
func GetUserBalanceLogsShards() int {
	shards := cast.ToInt(facades.Config().Get("sharding.user_balance_logs_shards", 4))
	if shards <= 0 {
		return 4
	}
	return shards
}
