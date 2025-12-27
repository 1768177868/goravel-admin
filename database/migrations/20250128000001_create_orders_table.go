package migrations

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250128000001CreateOrdersTable struct {
}

func (r *M20250128000001CreateOrdersTable) Signature() string {
	return "20250128000001_create_orders_table"
}

func (r *M20250128000001CreateOrdersTable) Up() error {
	// 订单表使用按月分表策略，不创建基础表
	// 分表通过命令 order:create-sharding-tables 创建，格式为 orders_YYYYMM 和 order_details_YYYYMM
	// 也可以在创建订单时自动创建（如果分表不存在）
	// 表结构定义在 CreateOrdersShardingTable 和 CreateOrderDetailsShardingTable 函数中
	return nil
}

func (r *M20250128000001CreateOrdersTable) Down() error {
	// 订单表使用分表，不删除基础表（因为基础表不存在）
	// 如需删除分表，请手动执行 DROP TABLE 语句
	return nil
}

// CreateOrdersShardingTable 创建订单主表分表（供服务层和命令层调用）
func CreateOrdersShardingTable(tableName string) error {
	return facades.Schema().Create(tableName, func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("order_no", 50).Comment("订单号")
		table.UnsignedBigInteger("user_id").Comment("用户ID")
		table.Decimal("amount").Comment("订单金额(10,2)")
		table.String("status", 20).Default("pending").Comment("订单状态 pending:待支付 paid:已支付 cancelled:已取消")
		table.Text("remark").Nullable().Comment("备注")
		table.Timestamps()
		table.SoftDeletes()
		table.Unique("order_no") // 唯一索引，防止并发下订单号重复
		table.Index("user_id")
		table.Index("created_at")
		table.Comment(fmt.Sprintf("订单主表 - %s", tableName))
	})
}

// CreateOrderDetailsShardingTable 创建订单详情表分表（供服务层和命令层调用）
func CreateOrderDetailsShardingTable(tableName string) error {
	return facades.Schema().Create(tableName, func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.UnsignedBigInteger("order_id").Comment("订单ID")
		table.UnsignedBigInteger("product_id").Comment("商品ID")
		table.String("product_name", 200).Comment("商品名称")
		table.Decimal("price").Comment("单价(10,2)")
		table.Integer("quantity").Comment("数量")
		table.Decimal("subtotal").Comment("小计(10,2)")
		table.Timestamps()
		table.SoftDeletes()
		table.Index("order_id")
		table.Index("product_id")
		table.Index("created_at")
		table.Comment(fmt.Sprintf("订单详情表 - %s", tableName))
	})
}
