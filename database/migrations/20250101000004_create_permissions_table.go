package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000004CreatePermissionsTable struct {
}

func (r *M20250101000004CreatePermissionsTable) Signature() string {
	return "20250101000004_create_permissions_table"
}

func (r *M20250101000004CreatePermissionsTable) Up() error {
	if !facades.Schema().HasTable("permissions") {
		return facades.Schema().Create("permissions", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("name")
			table.String("slug")
			table.String("method").Nullable()
			table.String("path").Nullable()
			table.String("description").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.Integer("sort").Default(0)
			table.UnsignedBigInteger("menu_id").Nullable().Default(0).Comment("关联菜单ID")
			table.Index("menu_id")
			table.Timestamps()
			table.Unique("name")
			table.Unique("slug")
			table.Comment("权限表")
		})
	}

	return nil
}

func (r *M20250101000004CreatePermissionsTable) Down() error {
	return facades.Schema().DropIfExists("permissions")
}
