package migrations

import (
	"context"
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

// M20250131000003AddTransactionHashToUserBalanceLogsShardingTables 为用户余额变动记录分表添加交易哈希字段
// MySQL 8.0+ INSTANT ADD COLUMN 瞬间完成
type M20250131000003AddTransactionHashToUserBalanceLogsShardingTables struct {
}

func (r *M20250131000003AddTransactionHashToUserBalanceLogsShardingTables) Signature() string {
	return "20250131000003_add_transaction_hash_to_user_balance_logs_sharding_tables"
}

func (r *M20250131000003AddTransactionHashToUserBalanceLogsShardingTables) Up() error {
	// 获取所有已存在的用户余额变动记录分表
	balanceLogsTables, err := utils.GetAllExistingShardingTables(context.Background(), "user_balance_logs")
	if err != nil {
		return fmt.Errorf("获取用户余额变动记录分表列表失败: %v", err)
	}

	if len(balanceLogsTables) == 0 {
		facades.Log().Info("没有找到需要修改的分表")
		return nil
	}

	totalTables := len(balanceLogsTables)
	modifiedCount := 0
	skippedCount := 0
	failedTables := []string{}

	facades.Log().Info(fmt.Sprintf("开始处理 %d 张用户余额变动记录分表", totalTables))

	for i, tableName := range balanceLogsTables {
		progress := fmt.Sprintf("[%d/%d]", i+1, totalTables)

		// 检查表是否存在
		if !facades.Schema().HasTable(tableName) {
			skippedCount++
			continue
		}

		// 检查字段是否已存在
		if facades.Schema().HasColumn(tableName, "transaction_hash") {
			skippedCount++
			continue
		}

		if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
			table.String("transaction_hash", 64).Nullable().Comment("交易哈希(区块链交易标识)")
			table.String("blockchain_address", 100).Nullable().Comment("区块链地址")
			table.String("network", 20).Nullable().Comment("区块链网络:ethereum,btc,tron等")
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

func (r *M20250131000003AddTransactionHashToUserBalanceLogsShardingTables) Down() error {
	balanceLogsTables, err := utils.GetAllExistingShardingTables(context.Background(), "user_balance_logs")
	if err != nil {
		return fmt.Errorf("获取用户余额变动记录分表列表失败: %v", err)
	}

	deletedCount := 0
	failedTables := []string{}

	for i, tableName := range balanceLogsTables {
		progress := fmt.Sprintf("[%d/%d]", i+1, len(balanceLogsTables))

		if !facades.Schema().HasTable(tableName) {
			continue
		}

		// 删除新增的字段
		if facades.Schema().HasColumn(tableName, "transaction_hash") {
			if err := facades.Schema().Table(tableName, func(table schema.Blueprint) {
				table.DropColumn("transaction_hash")
				table.DropColumn("blockchain_address")
				table.DropColumn("network")
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
