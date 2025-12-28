package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250130000001CreateUsersTable struct {
}

func (r *M20250130000001CreateUsersTable) Signature() string {
	return "20250130000001_create_users_table"
}

func (r *M20250130000001CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		if err := facades.Schema().Create("users", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("username", 50).Comment("用户名")
			table.String("password", 255).Comment("密码")
			table.String("nickname", 50).Nullable().Comment("昵称")
			table.String("avatar", 255).Nullable().Comment("头像")
			table.String("email", 100).Nullable().Comment("邮箱")
			table.String("phone", 20).Nullable().Comment("手机号")
			table.Decimal("balance").Total(18).Places(8).Default(0).Comment("当前余额(18,8)")
			table.UnsignedBigInteger("currency_id").Nullable().Comment("货币ID")
			table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:启用 0:禁用")
			table.Timestamp("last_login_at").Nullable().Comment("最后登录时间")
			table.Timestamps()
			table.SoftDeletes()

			// 创建唯一索引（考虑软删除）
			table.Unique("username", "deleted_at")
			table.Index("email")
			table.Index("phone")
			table.Index("currency_id")
			table.Index("status")
			table.Comment("用户表")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20250130000001CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
