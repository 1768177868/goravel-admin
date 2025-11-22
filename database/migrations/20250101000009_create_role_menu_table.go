package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000009CreateRoleMenuTable struct {
}

func (r *M20250101000009CreateRoleMenuTable) Signature() string {
	return "20250101000009_create_role_menu_table"
}

func (r *M20250101000009CreateRoleMenuTable) Up() error {
	if !facades.Schema().HasTable("role_menu") {
		return facades.Schema().Create("role_menu", func(table schema.Blueprint) {
			table.UnsignedBigInteger("role_id")
			table.UnsignedBigInteger("menu_id")
			table.Timestamps()
			table.Comment("角色菜单关联表")
		})
	}

	return nil
}

func (r *M20250101000009CreateRoleMenuTable) Down() error {
	return facades.Schema().DropIfExists("role_menu")
}
