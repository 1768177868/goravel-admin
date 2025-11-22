package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000006CreateDictionariesTable struct {
}

func (r *M20250101000006CreateDictionariesTable) Signature() string {
	return "20250101000006_create_dictionaries_table"
}

func (r *M20250101000006CreateDictionariesTable) Up() error {
	if !facades.Schema().HasTable("dictionaries") {
		return facades.Schema().Create("dictionaries", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("type").Default("")
			table.String("label").Default("")
			table.String("value").Default("")
			table.String("description").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.Integer("sort").Default(0)
			table.String("remark").Nullable()
			table.Timestamps()
			table.Comment("字典表")
		})
	}

	return nil
}

func (r *M20250101000006CreateDictionariesTable) Down() error {
	return facades.Schema().DropIfExists("dictionaries")
}
