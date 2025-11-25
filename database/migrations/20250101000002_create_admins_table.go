package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000002CreateAdminsTable struct {
}

func (r *M20250101000002CreateAdminsTable) Signature() string {
	return "20250101000002_create_admins_table"
}

func (r *M20250101000002CreateAdminsTable) Up() error {
	if !facades.Schema().HasTable("admins") {
		if err := facades.Schema().Create("admins", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("username")
			table.String("password").Default("")
			table.String("nickname").Nullable()
			table.String("avatar").Nullable()
			table.String("email").Nullable()
			table.String("phone").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.UnsignedBigInteger("department_id").Nullable()
			table.UnsignedTinyInteger("token_never_expires").Default(0).Comment("token是否永久有效 1:永久有效 0:按配置过期")
			table.Timestamps()
			table.SoftDeletes()
			// 创建 username 和 deleted_at 的联合唯一索引
			// 这样软删除的记录（deleted_at 不为 NULL）不会与未删除的记录冲突
			table.Unique("username", "deleted_at")
			table.Comment("管理员表")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20250101000002CreateAdminsTable) Down() error {
	return facades.Schema().DropIfExists("admins")
}
