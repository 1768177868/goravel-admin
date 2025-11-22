package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000012CreateSystemLogsTable struct {
}

func (r *M20250101000012CreateSystemLogsTable) Signature() string {
	return "20250101000012_create_system_logs_table"
}

func (r *M20250101000012CreateSystemLogsTable) Up() error {
	if !facades.Schema().HasTable("system_logs") {
		return facades.Schema().Create("system_logs", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("level").Nullable()
			table.String("module").Nullable()
			table.Text("message").Nullable()
			table.Text("context").Nullable()
			table.String("ip").Nullable()
			table.String("user_agent").Nullable()
			table.Timestamps()
			table.Comment("系统日志表")
		})
	}

	return nil
}

func (r *M20250101000012CreateSystemLogsTable) Down() error {
	return facades.Schema().DropIfExists("system_logs")
}
