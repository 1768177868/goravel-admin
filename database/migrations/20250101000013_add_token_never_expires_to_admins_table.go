package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000013AddTokenNeverExpiresToAdminsTable struct {
}

func (r *M20250101000013AddTokenNeverExpiresToAdminsTable) Signature() string {
	return "20250101000013_add_token_never_expires_to_admins_table"
}

func (r *M20250101000013AddTokenNeverExpiresToAdminsTable) Up() error {
	if !facades.Schema().HasTable("admins") {
		return nil
	}

	// 检查列是否已存在，如果不存在则添加
	if !facades.Schema().HasColumn("admins", "token_never_expires") {
		return facades.Schema().Table("admins", func(table schema.Blueprint) {
			table.UnsignedTinyInteger("token_never_expires").Default(0).Comment("token是否永久有效 1:永久有效 0:按配置过期")
		})
	}

	return nil
}

func (r *M20250101000013AddTokenNeverExpiresToAdminsTable) Down() error {
	if facades.Schema().HasTable("admins") {
		return facades.Schema().Table("admins", func(table schema.Blueprint) {
			table.DropColumn("token_never_expires")
		})
	}

	return nil
}
