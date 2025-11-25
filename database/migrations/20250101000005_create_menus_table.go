package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000005CreateMenusTable struct {
}

func (r *M20250101000005CreateMenusTable) Signature() string {
	return "20250101000005_create_menus_table"
}

func (r *M20250101000005CreateMenusTable) Up() error {
	if !facades.Schema().HasTable("menus") {
		return facades.Schema().Create("menus", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.UnsignedBigInteger("parent_id").Default(0)
			table.String("title").Default("")
			table.String("slug").Nullable()
			table.String("icon").Nullable()
			table.String("path").Nullable()
			table.String("component").Nullable()
			table.String("permission").Nullable()
			table.UnsignedTinyInteger("type").Default(1)
			table.UnsignedTinyInteger("status").Default(1)
			table.Integer("sort").Default(0)
			table.UnsignedTinyInteger("is_hidden").Default(0)
			table.Timestamps()
			table.Unique("slug")
			table.Comment("菜单表")
		})
	}

	return nil
}

func (r *M20250101000005CreateMenusTable) Down() error {
	return facades.Schema().DropIfExists("menus")
}
