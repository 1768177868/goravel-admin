package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250131000002CreatePaymentsTable struct {
}

func (r *M20250131000002CreatePaymentsTable) Signature() string {
	return "20250131000002_create_payments_table"
}

func (r *M20250131000002CreatePaymentsTable) Up() error {
	if !facades.Schema().HasTable("payments") {
		if err := facades.Schema().Create("payments", func(table schema.Blueprint) {
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

			table.Index("payment_no")
			table.Index("order_no")
			table.Index("payment_method_id")
			table.Index("user_id")
			table.Index("status")
			table.Index("third_party_no")
			table.Comment("支付记录表")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20250131000002CreatePaymentsTable) Down() error {
	return facades.Schema().DropIfExists("payments")
}

