package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type ScheduleTestLog struct{}

func (r *ScheduleTestLog) Signature() string {
	return "app:schedule-test-log"
}

func (r *ScheduleTestLog) Description() string {
	return "定时任务测试：写入一条心跳日志"
}

func (r *ScheduleTestLog) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

func (r *ScheduleTestLog) Handle(ctx console.Context) error {
	now := time.Now()
	logDir := filepath.Join("storage", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("schedule-test-%s.log", now.Format("2006-01-02")))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("[%s] [schedule-test] heartbeat\n", now.Format("2006-01-02 15:04:05"))
	if _, err := f.WriteString(line); err != nil {
		return err
	}

	facades.Log().Infof("[schedule-test] heartbeat written to %s", logPath)
	ctx.Info("schedule test log written: " + logPath)
	return nil
}
