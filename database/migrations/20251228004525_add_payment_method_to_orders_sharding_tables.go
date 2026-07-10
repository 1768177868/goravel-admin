package migrations

import (
	"context"
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

// M20251228004525AddPaymentMethodToOrdersShardingTables 为所有订单分表添加支付方式字段
// MySQL 8.0+ INSTANT ADD COLUMN 瞬间完成
type M20251228004525AddPaymentMethodToOrdersShardingTables struct {
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Signature() string {
	return "20251228004525_add_payment_method_to_orders_sharding_tables"
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Up() error {
	// 获取所有已存在的订单主表分表
	ordersTables, err := utils.GetAllExistingShardingTables(context.Background(), "orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	if len(ordersTables) == 0 {
		facades.Log().Info("没有找到需要修改的分表")
		return nil
	}

	totalTables := len(ordersTables)
	modifiedCount := 0
	skippedCount := 0
	failedTables := []string{}

	facades.Log().Info(fmt.Sprintf("开始处理 %d 张分表", totalTables))

	for i, tableName := range ordersTables {
		progress := fmt.Sprintf("[%d/%d]", i+1, totalTables)

		// 检查表是否存在
		if !facades.Schema().HasTable(tableName) {
			skippedCount++
			continue
		}

		// 检查字段是否已存在
		if facades.Schema().HasColumn(tableName, "payment_method") {
			skippedCount++
			continue
		}

		if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
			table.String("payment_method", 50).Nullable().Comment("支付方式: alipay, wechat, bank")
		}); err != nil {
			facades.Log().Errorf("%s ✗ 表 %s 修改失败: %v", progress, tableName, err)
			failedTables = append(failedTables, tableName)
			continue
		}

		facades.Log().Infof("%s ✓ %s", progress, tableName)
		modifiedCount++
	}

	facades.Log().Info(fmt.Sprintf("✅ 完成！成功: %d, 跳过: %d", modifiedCount, skippedCount))
	if len(failedTables) > 0 {
		facades.Log().Warning(fmt.Sprintf("❌ 失败: %v", failedTables))
		return fmt.Errorf("部分分表修改失败")
	}

	return nil
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Down() error {
	ordersTables, err := utils.GetAllExistingShardingTables(context.Background(), "orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	deletedCount := 0
	failedTables := []string{}

	for i, tableName := range ordersTables {
		progress := fmt.Sprintf("[%d/%d]", i+1, len(ordersTables))

		if !facades.Schema().HasTable(tableName) {
			continue
		}

		if facades.Schema().HasColumn(tableName, "payment_method") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.DropColumn("payment_method")
			}); err != nil {
				facades.Log().Errorf("%s ✗ 回滚 %s 失败: %v", progress, tableName, err)
				failedTables = append(failedTables, tableName)
				continue
			}
			facades.Log().Infof("%s ✓ %s", progress, tableName)
			deletedCount++
		}
	}

	facades.Log().Info(fmt.Sprintf("✅ 回滚完成！删除: %d", deletedCount))
	if len(failedTables) > 0 {
		facades.Log().Warning(fmt.Sprintf("❌ 失败: %v", failedTables))
	}

	return nil
}
