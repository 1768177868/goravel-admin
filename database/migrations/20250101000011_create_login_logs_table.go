package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000011CreateLoginLogsTable struct {
}

func (r *M20250101000011CreateLoginLogsTable) Signature() string {
	return "20250101000011_create_login_logs_table"
}

func (r *M20250101000011CreateLoginLogsTable) Up() error {
	if !facades.Schema().HasTable("login_logs") {
		return facades.Schema().Create("login_logs", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.UnsignedBigInteger("admin_id").Nullable()
			table.String("username").Nullable()
			table.String("ip").Nullable()
			table.String("user_agent").Nullable()
			table.String("location").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.String("message").Nullable()
			table.Timestamps()
			table.Comment("登录日志表")
		})
	}

	return nil
}

func (r *M20250101000011CreateLoginLogsTable) Down() error {
	return facades.Schema().DropIfExists("login_logs")
}
