package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250130000003CreateCurrenciesTable struct {
}

func (r *M20250130000003CreateCurrenciesTable) Signature() string {
	return "20250130000003_create_currencies_table"
}

func (r *M20250130000003CreateCurrenciesTable) Up() error {
	if !facades.Schema().HasTable("currencies") {
		if err := facades.Schema().Create("currencies", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("code", 10).Comment("货币代码(如CNY,USD)")
			table.String("name", 50).Comment("货币名称(如人民币,美元)")
			table.String("symbol", 10).Nullable().Comment("货币符号(如¥,$)")
			table.Decimal("rate").Total(18).Places(8).Default(1).Comment("汇率(相对于基准货币)")
			table.Boolean("is_default").Default(false).Comment("是否默认货币")
			table.Boolean("is_active").Default(true).Comment("是否启用")
			table.Integer("sort").Default(0).Comment("排序")
			table.Text("description").Nullable().Comment("描述")
			table.Timestamps()
			table.SoftDeletes()
			
			table.Unique("code")
			table.Index("is_default")
			table.Index("is_active")
			table.Index("sort")
			table.Comment("货币表")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20250130000003CreateCurrenciesTable) Down() error {
	return facades.Schema().DropIfExists("currencies")
}

