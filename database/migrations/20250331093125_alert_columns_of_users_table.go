package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250331093125AlertColumnsOfUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20250331093125AlertColumnsOfUsersTable) Signature() string {
	return "20250331093125_alert_columns_of_users_table"
}

// Up Run the migrations.
func (r *M20250331093125AlertColumnsOfUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return nil
	}

	hasEmail := facades.Schema().HasColumn("users", "email")
	hasMail := facades.Schema().HasColumn("users", "mail")
	hasAlias := facades.Schema().HasColumn("users", "alias")

	// 只有在需要修改时才执行
	if (hasEmail && !hasMail) || hasAlias {
		return facades.Schema().Table("users", func(table schema.Blueprint) {
			// 如果 alias 列存在，则修改默认值
			if hasAlias {
				table.String("alias").Default("test").Change()
			}
			// 如果 email 列存在且 mail 列不存在，则重命名
			if hasEmail && !hasMail {
				table.RenameColumn("email", "mail")
			}
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20250331093125AlertColumnsOfUsersTable) Down() error {
	return nil
}
