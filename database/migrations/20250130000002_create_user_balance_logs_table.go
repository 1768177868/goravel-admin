package migrations

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250130000002CreateUserBalanceLogsTable struct {
}

func (r *M20250130000002CreateUserBalanceLogsTable) Signature() string {
	return "20250130000002_create_user_balance_logs_table"
}

func (r *M20250130000002CreateUserBalanceLogsTable) Up() error {
	// 用户余额变动记录表使用自定义分表逻辑（按 user_id 哈希分表）
	// 不创建基础表，分表通过 CreateUserBalanceLogsShardingTable 函数创建
	// 分表格式为 user_balance_logs_0, user_balance_logs_1, user_balance_logs_2, user_balance_logs_3
	// 分表逻辑：user_id % UserBalanceLogsShards
	// 表结构定义在 CreateUserBalanceLogsShardingTable 函数中
	return nil
}

func (r *M20250130000002CreateUserBalanceLogsTable) Down() error {
	// 用户余额变动记录表使用分表，不删除基础表（因为基础表不存在）
	// 如需删除分表，请手动执行 DROP TABLE 语句
	return nil
}

// CreateUserBalanceLogsShardingTable 创建用户余额变动记录分表（供服务层调用）
func CreateUserBalanceLogsShardingTable(tableName string) error {
	return facades.Schema().Create(tableName, func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.UnsignedBigInteger("user_id").Comment("用户ID")
		table.String("type", 20).Comment("变动类型:income收入,expense支出,refund退款")
		table.Decimal("amount").Total(18).Places(8).Comment("变动金额")
		table.Decimal("balance").Total(18).Places(8).Comment("变动后余额")
		table.String("source", 50).Nullable().Comment("来源:order订单,recharge充值,withdraw提现,manual手动")
		table.UnsignedBigInteger("source_id").Nullable().Comment("来源ID(如订单ID)")
		table.Text("description").Nullable().Comment("描述")
		table.UnsignedBigInteger("operator_id").Nullable().Comment("操作员ID")
		table.String("status", 20).Default("success").Comment("状态:success成功,failed失败")
		table.Text("remark").Nullable().Comment("备注")
		table.Timestamps()
		table.SoftDeletes()
		table.Index("user_id")
		table.Index("source_id")
		table.Index("operator_id")
		table.Index("created_at")
		table.Comment(fmt.Sprintf("用户余额变动记录表 - %s", tableName))
	})
}
