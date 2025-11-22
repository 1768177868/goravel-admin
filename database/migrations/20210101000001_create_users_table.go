package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20210101000001CreateUsersTable struct {
}

// Signature The unique signature for the migration.
func (r *M20210101000001CreateUsersTable) Signature() string {
	return "20210101000001_create_users_table"
}

// Up Run the migrations.
func (r *M20210101000001CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("username", 50).Nullable().Comment("用户名")
			table.String("password", 255).Nullable().Comment("密码")
			table.String("name").Default("").Comment("姓名")
			table.String("avatar").Default("").Comment("头像")
			table.String("alias").Default("").Comment("别名")
			table.String("mail").Nullable().Comment("邮箱")
			table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:启用 0:禁用")
			table.Json("tags").Nullable().Comment("标签")
			table.Timestamps()
			table.SoftDeletes()
			table.Unique("username")
			table.Comment("user table")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20210101000001CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
