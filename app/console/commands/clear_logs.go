package commands

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type ClearLogs struct {
}

// Signature The name and signature of the console command.
func (r *ClearLogs) Signature() string {
	return "app:clear-logs"
}

// Description The console command description.
func (r *ClearLogs) Description() string {
	return "清理3个月前的日志记录（操作日志、登录日志、系统日志）"
}

// Extend The console command extend.
func (r *ClearLogs) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *ClearLogs) Handle(ctx console.Context) error {
	// 计算3个月前的日期
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	ctx.Info("开始清理3个月前的日志...")

	// 清理操作日志
	operationLogResult, err := facades.Orm().Query().Model(&models.OperationLog{}).
		Where("created_at < ?", threeMonthsAgo).
		Delete(&models.OperationLog{})
	if err != nil {
		ctx.Error("清理操作日志失败: " + err.Error())
		return err
	}
	ctx.Info(fmt.Sprintf("已清理操作日志: %d 条", operationLogResult.RowsAffected))

	// 清理登录日志
	loginLogResult, err := facades.Orm().Query().Model(&models.LoginLog{}).
		Where("created_at < ?", threeMonthsAgo).
		Delete(&models.LoginLog{})
	if err != nil {
		ctx.Error("清理登录日志失败: " + err.Error())
		return err
	}
	ctx.Info(fmt.Sprintf("已清理登录日志: %d 条", loginLogResult.RowsAffected))

	// 清理系统日志
	systemLogResult, err := facades.Orm().Query().Model(&models.SystemLog{}).
		Where("created_at < ?", threeMonthsAgo).
		Delete(&models.SystemLog{})
	if err != nil {
		ctx.Error("清理系统日志失败: " + err.Error())
		return err
	}
	ctx.Info(fmt.Sprintf("已清理系统日志: %d 条", systemLogResult.RowsAffected))

	ctx.Info("日志清理完成！")
	return nil
}
