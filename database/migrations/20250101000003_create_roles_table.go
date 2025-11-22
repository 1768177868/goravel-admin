package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000003CreateRolesTable struct {
}

func (r *M20250101000003CreateRolesTable) Signature() string {
	return "20250101000003_create_roles_table"
}

func (r *M20250101000003CreateRolesTable) Up() error {
	if !facades.Schema().HasTable("roles") {
		return facades.Schema().Create("roles", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("name")
			table.String("slug")
			table.String("description").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.Integer("sort").Default(0)
			table.Timestamps()
			table.Unique("name")
			table.Unique("slug")
			table.Comment("角色表")
		})
	}

	return nil
}

func (r *M20250101000003CreateRolesTable) Down() error {
	return facades.Schema().DropIfExists("roles")
}
