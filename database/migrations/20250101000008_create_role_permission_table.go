package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000008CreateRolePermissionTable struct {
}

func (r *M20250101000008CreateRolePermissionTable) Signature() string {
	return "20250101000008_create_role_permission_table"
}

func (r *M20250101000008CreateRolePermissionTable) Up() error {
	if !facades.Schema().HasTable("role_permission") {
		return facades.Schema().Create("role_permission", func(table schema.Blueprint) {
			table.UnsignedBigInteger("role_id")
			table.UnsignedBigInteger("permission_id")
			table.Timestamps()
			table.Comment("角色权限关联表")
		})
	}

	return nil
}

func (r *M20250101000008CreateRolePermissionTable) Down() error {
	return facades.Schema().DropIfExists("role_permission")
}
