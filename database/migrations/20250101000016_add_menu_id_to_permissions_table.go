package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000016AddMenuIdToPermissionsTable struct {
}

func (r *M20250101000016AddMenuIdToPermissionsTable) Signature() string {
	return "20250101000016_add_menu_id_to_permissions_table"
}

func (r *M20250101000016AddMenuIdToPermissionsTable) Up() error {
	if facades.Schema().HasTable("permissions") {
		if !facades.Schema().HasColumn("permissions", "menu_id") {
			err := facades.Schema().Table("permissions", func(table schema.Blueprint) {
				table.UnsignedBigInteger("menu_id").Nullable().Default(0).Comment("关联菜单ID")
				table.Index("menu_id")
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *M20250101000016AddMenuIdToPermissionsTable) Down() error {
	if facades.Schema().HasTable("permissions") {
		if facades.Schema().HasColumn("permissions", "menu_id") {
			return facades.Schema().Table("permissions", func(table schema.Blueprint) {
				table.DropIndex("menu_id")
				table.DropColumn("menu_id")
			})
		}
	}

	return nil
}

