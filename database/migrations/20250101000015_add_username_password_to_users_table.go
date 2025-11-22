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
	if !facades.Schema().HasTable("users") {
		return nil
	}

	hasUsername := facades.Schema().HasColumn("users", "username")
	hasPassword := facades.Schema().HasColumn("users", "password")
	hasStatus := facades.Schema().HasColumn("users", "status")

	if !hasUsername || !hasPassword || !hasStatus {
		return facades.Schema().Table("users", func(table schema.Blueprint) {
			// 检查列是否存在，如果不存在则添加
			if !hasUsername {
				table.String("username", 50).Nullable().Comment("用户名")
			}
			if !hasPassword {
				table.String("password", 255).Nullable().Comment("密码")
			}
			if !hasStatus {
				table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:启用 0:禁用")
			}
			// 只有在添加username列时才添加唯一索引
			if !hasUsername {
				table.Unique("username")
			}
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

