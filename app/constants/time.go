package constants

import "time"

const (
	// OnlineUserThreshold 在线用户判断阈值
	// 如果用户的 last_used_at 在这个时间范围内，则认为用户在线
	OnlineUserThreshold = 15 * time.Minute

	// DefaultCleanLogDays 默认清理日志的天数
	DefaultCleanLogDays = 30
)
