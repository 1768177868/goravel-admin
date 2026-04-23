package config

import "github.com/goravel/framework/facades"

func init() {
	config := facades.Config()

	config.Add("schedule", map[string]any{
		// 测试任务调度 Cron（默认每 5 秒）
		// 格式：
		// - 5 段：分 时 日 月 周
		// - 6 段（秒级）：秒 分 时 日 月 周
		// 秒级示例（6 段）：
		// - 每5秒：*/5 * * * * *
		// - 每10秒：*/10 * * * * *
		// - 每30秒：*/30 * * * * *
		// - 每分钟第0秒：0 * * * * *
		// 分钟级示例（5 段）：
		// - 每分钟：*/1 * * * *
		// - 每5分钟：*/5 * * * *
		// - 每10分钟：*/10 * * * *
		// - 每30分钟：*/30 * * * *
		// - 每小时整点：0 * * * *
		// - 每天凌晨2点：0 2 * * *
		// - 每天 02:30：30 2 * * *
		// - 每周一凌晨1点：0 1 * * 1
		// - 每月1号凌晨1点：0 1 1 * *
		// "test_cron": config.Env("SCHEDULE_TEST_CRON", "*/5 * * * * *"),
	})
}
