package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("login_security", map[string]any{
		// 登录失败锁定：同一 IP + 同一账号维度
		// 同一 IP 下不同账号互不影响

		// 允许的最大连续失败次数，超过后锁定
		"max_attempts": 5,

		// 锁定时长（分钟）
		"lock_duration_minutes": 15,

		// 失败计数的衰减窗口（分钟），在此时间内未再次失败则自动清零
		"decay_minutes": 5,
	})
}
