package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000007CreateAdminRoleTable struct {
}

func (r *M20250101000007CreateAdminRoleTable) Signature() string {
	return "20250101000007_create_admin_role_table"
}

func (r *M20250101000007CreateAdminRoleTable) Up() error {
	if !facades.Schema().HasTable("admin_role") {
		return facades.Schema().Create("admin_role", func(table schema.Blueprint) {
			table.UnsignedBigInteger("admin_id")
			table.UnsignedBigInteger("role_id")
			table.Timestamps()
			table.Comment("管理员角色关联表")
		})
	}

	return nil
}

func (r *M20250101000007CreateAdminRoleTable) Down() error {
	return facades.Schema().DropIfExists("admin_role")
}
