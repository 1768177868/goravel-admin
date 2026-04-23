package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"

	"goravel/app/console/commands"
	"goravel/app/facades"
)

type Kernel struct {
}

func (kernel *Kernel) Schedule() []schedule.Event {
	// testCron := frameworkfacades.Config().GetString("schedule.test_cron", "*/5 * * * * *")

	return []schedule.Event{
		// 每天凌晨2点执行（北京时间），清理6个月前的日志
		// 北京时间 02:00 = UTC 18:00（前一天）
		facades.Schedule().Command("app:clear-logs").DailyAt("18:00").OnOneServer(),
		// 每天凌晨3点执行（北京时间），清理3天前的分片文件
		// 北京时间 03:00 = UTC 19:00（前一天）
		facades.Schedule().Command("app:clear-chunks").DailyAt("19:00").OnOneServer(),
		// 每天凌晨3点30分执行（UTC 19:30 / 北京时间 03:30）, 分析表（更新统计信息）
		facades.Schedule().Command("db:analyze-stats").DailyAt("19:30").OnOneServer(),
		// 每月1号凌晨1点执行（UTC时间），创建下个月的订单分表
		facades.Schedule().Command("order:create-sharding-tables").Monthly().OnOneServer(),
		// 每月1号凌晨1点30分执行（UTC时间），创建下个月的支付记录分表
		facades.Schedule().Command("payment:create-sharding-tables").Monthly().OnOneServer(),
		// 测试任务：支持通过 SCHEDULE_TEST_CRON 自定义频率（默认每5秒）
		// facades.Schedule().Command("app:schedule-test-log").Cron(testCron).OnOneServer(),
	}
}

func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.ClearLogs{},
		&commands.ClearChunks{},
		&commands.CreateToken{},
		&commands.QueueStats{},
		&commands.QueueClear{},
		&commands.QueuePeek{},
		&commands.ScheduleTestLog{},
		commands.NewCreateOrderShardingTables(),
		commands.NewCreatePaymentShardingTables(),
		&commands.GenerateTestOrders{},
		&commands.GenerateTestPayments{},
		&commands.AnalyzeStats{},
		&commands.OptimizeTables{},
		&commands.ElasticsearchExample{},
		&commands.SyncOrdersElasticsearch{},
	}
}
