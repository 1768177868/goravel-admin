package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250131000001CreatePaymentMethodsTable struct {
}

func (r *M20250131000001CreatePaymentMethodsTable) Signature() string {
	return "20250131000001_create_payment_methods_table"
}

func (r *M20250131000001CreatePaymentMethodsTable) Up() error {
	if !facades.Schema().HasTable("payment_methods") {
		if err := facades.Schema().Create("payment_methods", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("name", 50).Comment("支付方式名称(如微信支付,支付宝)")
			table.String("code", 20).Comment("支付方式代码(如wechat,alipay)")
			table.String("type", 20).Comment("支付类型(如wechat,alipay,qq,allinpay,lakala,paypal,apple,saobei)")
			table.Text("config").Nullable().Comment("支付配置(JSON格式,存储密钥、证书等敏感信息)")
			table.Boolean("is_active").Default(true).Comment("是否启用")
			table.Integer("sort").Default(0).Comment("排序")
			table.Text("description").Nullable().Comment("描述")
			table.Timestamps()
			table.SoftDeletes()

			table.Unique("code")
			table.Index("type")
			table.Index("is_active")
			table.Index("sort")
			table.Comment("支付方式表")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20250131000001CreatePaymentMethodsTable) Down() error {
	return facades.Schema().DropIfExists("payment_methods")
}

