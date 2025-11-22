package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000015AddUsernamePasswordToUsersTable struct {
}

func (r *M20250101000015AddUsernamePasswordToUsersTable) Signature() string {
	return "20250101000015_add_username_password_to_users_table"
}

func (r *M20250101000015AddUsernamePasswordToUsersTable) Up() error {
	if facades.Schema().HasTable("users") {
		return facades.Schema().Table("users", func(table schema.Blueprint) {
			table.String("username", 50).Nullable().Comment("用户名")
			table.String("password", 255).Nullable().Comment("密码")
			table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:启用 0:禁用")
			table.Unique("username")
		})
	}

	return nil
}

func (r *M20250101000015AddUsernamePasswordToUsersTable) Down() error {
	if facades.Schema().HasTable("users") {
		return facades.Schema().Table("users", func(table schema.Blueprint) {
			table.DropColumn("username")
			table.DropColumn("password")
			table.DropColumn("status")
		})
	}

	return nil
}

