package migrations

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

// M20251228001858ExampleModifyOrdersShardingTables 示例：修改订单分表字段
//
// ⚠️ 注意：这是一个示例 migration 文件，仅用于参考学习
// 实际使用时，请：
// 1. 复制此文件并重命名为新的 migration（使用新的时间戳）
// 2. 根据实际需求修改字段定义和表名
// 3. 在 database/kernel.go 中注册新的 migration
// 4. 删除此示例文件或将其移到其他位置
//
// 详细使用说明请参考：docs/SHARDING_MIGRATION.md
type M20251228001858ExampleModifyOrdersShardingTables struct {
}

func (r *M20251228001858ExampleModifyOrdersShardingTables) Signature() string {
	return "20251228001858_example_modify_orders_sharding_tables"
}

func (r *M20251228001858ExampleModifyOrdersShardingTables) Up() error {
	// 示例1: 为所有订单主表分表添加新字段
	// 获取所有已存在的订单主表分表
	ordersTables, err := utils.GetAllExistingShardingTables("orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	// 遍历所有分表，添加字段
	for _, tableName := range ordersTables {
		// 检查表是否存在
		if !facades.Schema().HasTable(tableName) {
			continue
		}

		// 检查字段是否已存在（避免重复添加）
		if !facades.Schema().HasColumn(tableName, "payment_method") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				// 添加支付方式字段
				table.String("payment_method", 50).Nullable().Comment("支付方式: alipay, wechat, bank").After("status")
			}); err != nil {
				return fmt.Errorf("修改分表 %s 失败: %v", tableName, err)
			}
			facades.Log().Infof("✓ 已为分表 %s 添加字段 payment_method", tableName)
		}

		// 示例：修改字段（如果需要）
		// 注意：某些数据库可能不支持直接修改字段类型，需要先删除再添加
		// 这里只是示例，实际使用时需要根据数据库类型调整
		if facades.Schema().HasColumn(tableName, "remark") {
			// 如果需要修改字段，可能需要先删除再添加，或者使用 ALTER TABLE 语句
			// 这里仅作示例，实际使用时请谨慎操作
		}
	}

	// 示例2: 为所有订单详情表分表添加新字段
	orderDetailsTables, err := utils.GetAllExistingShardingTables("order_details")
	if err != nil {
		return fmt.Errorf("获取订单详情分表列表失败: %v", err)
	}

	for _, tableName := range orderDetailsTables {
		if !facades.Schema().HasTable(tableName) {
			continue
		}

		// 添加商品图片字段
		if !facades.Schema().HasColumn(tableName, "product_image") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.String("product_image", 500).Nullable().Comment("商品图片URL").After("product_name")
			}); err != nil {
				return fmt.Errorf("修改分表 %s 失败: %v", tableName, err)
			}
			facades.Log().Infof("✓ 已为分表 %s 添加字段 product_image", tableName)
		}
	}

	facades.Log().Info(fmt.Sprintf("完成！共修改了 %d 个订单主表分表和 %d 个订单详情表分表", len(ordersTables), len(orderDetailsTables)))
	return nil
}

func (r *M20251228001858ExampleModifyOrdersShardingTables) Down() error {
	// 回滚操作：删除添加的字段

	// 获取所有订单主表分表
	ordersTables, err := utils.GetAllExistingShardingTables("orders")
	if err != nil {
		return fmt.Errorf("获取订单分表列表失败: %v", err)
	}

	// 删除 payment_method 字段
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
		}
	}

	// 获取所有订单详情表分表
	orderDetailsTables, err := utils.GetAllExistingShardingTables("order_details")
	if err != nil {
		return fmt.Errorf("获取订单详情分表列表失败: %v", err)
	}

	// 删除 product_image 字段
	for _, tableName := range orderDetailsTables {
		if !facades.Schema().HasTable(tableName) {
			continue
		}

		if facades.Schema().HasColumn(tableName, "product_image") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.DropColumn("product_image")
			}); err != nil {
				return fmt.Errorf("回滚分表 %s 失败: %v", tableName, err)
			}
			facades.Log().Infof("✓ 已从分表 %s 删除字段 product_image", tableName)
		}
	}

	return nil
}

