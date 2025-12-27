package migrations

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

// M20251228004525AddPaymentMethodToOrdersShardingTables 为所有订单分表添加支付方式字段
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

	// 遍历所有分表，添加字段
	modifiedCount := 0
	for _, tableName := range ordersTables {
		// 检查表是否存在
		if !facades.Schema().HasTable(tableName) {
			facades.Log().Infof("跳过不存在的分表: %s", tableName)
			continue
		}

		// 检查字段是否已存在（避免重复添加）
		if facades.Schema().HasColumn(tableName, "payment_method") {
			facades.Log().Infof("分表 %s 的字段 payment_method 已存在，跳过", tableName)
			continue
		}

		// 添加字段
		if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
			// 添加支付方式字段
			table.String("payment_method", 50).Nullable().Comment("支付方式: alipay, wechat, bank").After("status")
		}); err != nil {
			return fmt.Errorf("修改分表 %s 失败: %v", tableName, err)
		}
		facades.Log().Infof("✓ 已为分表 %s 添加字段 payment_method", tableName)
		modifiedCount++
	}

	facades.Log().Info(fmt.Sprintf("完成！共为 %d 个订单分表添加了 payment_method 字段（共 %d 个分表）", modifiedCount, len(ordersTables)))
	return nil
}

func (r *M20251228004525AddPaymentMethodToOrdersShardingTables) Down() error {
	// 回滚操作：删除添加的字段
	ordersTables, err := utils.GetAllExistingShardingTables("orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	deletedCount := 0
	for _, tableName := range ordersTables {
		if !facades.Schema().HasTable(tableName) {
			continue
		}

		if facades.Schema().HasColumn(tableName, "payment_method") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.DropColumn("payment_method")
			}); err != nil {
				return fmt.Errorf("回滚分表 %s 失败: %v", tableName, err)
			}
			facades.Log().Infof("✓ 已从分表 %s 删除字段 payment_method", tableName)
			deletedCount++
		}
	}

	facades.Log().Info(fmt.Sprintf("完成！共从 %d 个订单分表删除了 payment_method 字段", deletedCount))
	return nil
}

