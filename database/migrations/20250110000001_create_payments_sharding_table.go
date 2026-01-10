package migrations

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250110000001CreatePaymentsShardingTable struct {
}

func (r *M20250110000001CreatePaymentsShardingTable) Signature() string {
	return "20250110000001_create_payments_sharding_table"
}

func (r *M20250110000001CreatePaymentsShardingTable) Up() error {
	// 支付记录表使用按月分表策略，不创建基础表
	// 分表通过命令 payment:create-sharding-tables 创建，格式为 payments_YYYYMM
	// 也可以在创建支付记录时自动创建（如果分表不存在）
	// 表结构定义在 CreatePaymentsShardingTable 函数中
	return nil
}

func (r *M20250110000001CreatePaymentsShardingTable) Down() error {
	// 支付记录表使用分表，不删除基础表（因为基础表不存在）
	// 如需删除分表，请手动执行 DROP TABLE 语句
	return nil
}

// CreatePaymentsShardingTable 创建支付记录分表（供服务层和命令层调用）
func CreatePaymentsShardingTable(tableName string) error {
	return facades.Schema().Create(tableName, func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("payment_no", 50).Comment("支付单号")
		table.String("order_no", 50).Comment("订单号")
		table.UnsignedBigInteger("payment_method_id").Comment("支付方式ID")
		table.UnsignedBigInteger("user_id").Comment("用户ID")
		table.Decimal("amount").Total(10).Places(2).Comment("支付金额")
		table.String("status", 20).Default("pending").Comment("支付状态 pending:待支付 paid:已支付 failed:支付失败 cancelled:已取消")
		table.String("third_party_no", 100).Nullable().Comment("第三方支付单号")
		table.DateTime("pay_time").Nullable().Comment("支付时间")
		table.Text("fail_reason").Nullable().Comment("失败原因")
		table.Text("notify_data").Nullable().Comment("回调通知数据(JSON格式)")
		table.Text("remark").Nullable().Comment("备注")
		table.Timestamps()
		table.SoftDeletes()

		// 唯一索引，防止支付单号重复
		table.Unique("payment_no")
		// 复合索引，优化常用查询场景
		// 1. 时间范围 + 状态查询（最常用，用于按状态筛选支付记录）
		table.Index("created_at", "status")
		// 2. 时间范围 + 用户ID查询（用于查询特定用户的支付记录）
		table.Index("created_at", "user_id")
		// 3. 订单号查询
		table.Index("order_no")
		// 4. 第三方支付单号查询
		table.Index("third_party_no")
		// 5. 支付方式ID查询
		table.Index("payment_method_id")

		table.Comment(fmt.Sprintf("支付记录表 - %s", tableName))
	})
}
