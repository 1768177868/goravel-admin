package constants

import "time"

const (
	// OnlineAdminThreshold 在线管理员判断阈值
	// 如果管理员的 last_used_at 在这个时间范围内，则认为管理员在线
	OnlineAdminThreshold = 15 * time.Minute

	// DefaultCleanLogDays 默认清理日志的天数
	DefaultCleanLogDays = 30
)
