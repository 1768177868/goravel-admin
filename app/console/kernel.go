package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/facades"

	"goravel/app/console/commands"
)

type Kernel struct {
}

func (kernel *Kernel) Schedule() []schedule.Event {
	return []schedule.Event{
		// 每天凌晨2点执行，清理3个月前的日志
		facades.Schedule().Command("app:clear-logs").DailyAt("02:00").OnOneServer(),
	}
}
func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.ClearLogs{},
		&commands.CreateToken{},
	}
}
