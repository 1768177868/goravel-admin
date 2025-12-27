package migrations

import (
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

// M20251228004525AddPaymentMethodToOrdersShardingTables 为所有订单分表添加支付方式字段
// ⚠️ 生产环境大规模数据场景：
// 建议：在业务低峰期执行
type M20251228004525AddPaymentMethodToOrdersShardingTables struct {
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Signature() string {
	return "20251228004525_add_payment_method_to_orders_sharding_tables"
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Up() error {
	// 获取所有已存在的订单主表分表
	ordersTables, err := utils.GetAllExistingShardingTables("orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	if len(ordersTables) == 0 {
		facades.Log().Info("没有找到需要修改的分表")
		return nil
	}

	// 大规模数据场景：每次只处理1张表，避免并发压力
	batchSize := 1
	totalTables := len(ordersTables)
	modifiedCount := 0
	skippedCount := 0
	failedTables := []string{}

	separator := strings.Repeat("=", 60)
	facades.Log().Warning(separator)
	facades.Log().Warning(fmt.Sprintf("⚠️  大规模数据场景：共 %d 张表，单表千万级数据", totalTables))
	facades.Log().Warning("⚠️  建议在业务低峰期执行")
	facades.Log().Warning("⚠️  每张表处理完成后会根据执行时间动态等待，让数据库恢复")
	facades.Log().Warning(separator)
	facades.Log().Info(fmt.Sprintf("开始处理，每次处理 %d 张表", batchSize))

	startTime := time.Now()

	// 分批处理
	for i := 0; i < totalTables; i += batchSize {
		end := i + batchSize
		if end > totalTables {
			end = totalTables
		}

		batch := ordersTables[i:end]
		progress := fmt.Sprintf("[%d/%d]", i+1, totalTables)

		for _, tableName := range batch {
			// 检查表是否存在
			if !facades.Schema().HasTable(tableName) {
				facades.Log().Infof("%s 跳过不存在的分表: %s", progress, tableName)
				skippedCount++
				continue
			}

			// 检查字段是否已存在（避免重复添加，支持中断后继续）
			if facades.Schema().HasColumn(tableName, "payment_method") {
				facades.Log().Infof("%s 分表 %s 的字段已存在，跳过", progress, tableName)
				skippedCount++
				continue
			}

			// 记录开始时间
			tableStartTime := time.Now()
			facades.Log().Infof("%s 开始处理表: %s (预计5-15分钟)", progress, tableName)

			// ⚠️ 重要：不使用 AFTER，可以显著加快速度（快30-50%）
			// 如果必须指定位置，建议使用 pt-online-schema-change
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.String("payment_method", 50).Nullable().Comment("支付方式: alipay, wechat, bank")
			}); err != nil {
				duration := time.Since(tableStartTime)
				facades.Log().Errorf("%s ✗ 表 %s 修改失败 (耗时: %v): %v", progress, tableName, duration, err)
				failedTables = append(failedTables, tableName)
				continue
			}

			duration := time.Since(tableStartTime)
			elapsed := time.Since(startTime)
			avgTime := elapsed / time.Duration(modifiedCount+1)
			estimatedRemaining := avgTime * time.Duration(totalTables-modifiedCount-1)

			facades.Log().Infof("%s ✓ 表 %s 完成 (耗时: %v, 总耗时: %v, 预计剩余: %v)",
				progress, tableName, duration, elapsed, estimatedRemaining)
			modifiedCount++

			// 根据执行时间动态调整等待时间，让数据库恢复
			// 执行时间短则等待时间短，执行时间长则等待时间长
			if modifiedCount < totalTables {
				// 最小等待10秒，最大等待60秒
				// 等待时间 = 执行时间的20%，但不少于10秒，不超过60秒
				waitTime := duration / 5
				if waitTime < 10*time.Second {
					waitTime = 10 * time.Second
				}
				if waitTime > 60*time.Second {
					waitTime = 60 * time.Second
				}
				facades.Log().Infof("等待 %v 后继续下一张表...", waitTime)
				time.Sleep(waitTime)
			}
		}
	}

	// 输出最终结果
	totalDuration := time.Since(startTime)
	facades.Log().Info(separator)
	facades.Log().Info(fmt.Sprintf("✅ 完成！总耗时: %v", totalDuration))
	facades.Log().Info(fmt.Sprintf("成功: %d 张表", modifiedCount))
	if skippedCount > 0 {
		facades.Log().Info(fmt.Sprintf("跳过: %d 张表（字段已存在或表不存在）", skippedCount))
	}
	if len(failedTables) > 0 {
		facades.Log().Warning(fmt.Sprintf("❌ 失败: %d 张表，需要手动处理: %v", len(failedTables), failedTables))
		return fmt.Errorf("部分分表修改失败，请检查日志并手动处理")
	}
	facades.Log().Info(separator)

	return nil
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Down() error {
	// 回滚操作：删除添加的字段
	ordersTables, err := utils.GetAllExistingShardingTables("orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	facades.Log().Warning("⚠️  开始回滚：删除 payment_method 字段")
	deletedCount := 0
	failedTables := []string{}

	for i, tableName := range ordersTables {
		progress := fmt.Sprintf("[%d/%d]", i+1, len(ordersTables))

		if !facades.Schema().HasTable(tableName) {
			facades.Log().Infof("%s 跳过不存在的分表: %s", progress, tableName)
			continue
		}

		if facades.Schema().HasColumn(tableName, "payment_method") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.DropColumn("payment_method")
			}); err != nil {
				facades.Log().Errorf("%s ✗ 回滚分表 %s 失败: %v", progress, tableName, err)
				failedTables = append(failedTables, tableName)
				continue
			}
			facades.Log().Infof("%s ✓ 已从分表 %s 删除字段 payment_method", progress, tableName)
			deletedCount++

			// 每张表之间稍作延迟
			if i < len(ordersTables)-1 {
				time.Sleep(10 * time.Second)
			}
		}
	}

	facades.Log().Info(fmt.Sprintf("完成！共从 %d 个订单分表删除了 payment_method 字段", deletedCount))
	if len(failedTables) > 0 {
		facades.Log().Warning(fmt.Sprintf("回滚失败的表: %v", failedTables))
	}

	return nil
}
