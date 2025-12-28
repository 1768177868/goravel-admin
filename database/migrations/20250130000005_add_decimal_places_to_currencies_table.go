package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250130000005AddDecimalPlacesToCurrenciesTable struct {
}

func (r *M20250130000005AddDecimalPlacesToCurrenciesTable) Signature() string {
	return "20250130000005_add_decimal_places_to_currencies_table"
}

func (r *M20250130000005AddDecimalPlacesToCurrenciesTable) Up() error {
	if facades.Schema().HasColumn("currencies", "decimal_places") {
		return nil
	}

	return facades.Schema().Table("currencies", func(table schema.Blueprint) {
		table.Integer("decimal_places").Default(2).Comment("小数位数(如日元为0,人民币为2,虚拟货币为8)")
	})
}

func (r *M20250130000005AddDecimalPlacesToCurrenciesTable) Down() error {
	if !facades.Schema().HasColumn("currencies", "decimal_places") {
		return nil
	}

	return facades.Schema().Table("currencies", func(table schema.Blueprint) {
		table.DropColumn("decimal_places")
	})
}
