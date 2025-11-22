package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000010CreateOperationLogsTable struct {
}

func (r *M20250101000010CreateOperationLogsTable) Signature() string {
	return "20250101000010_create_operation_logs_table"
}

func (r *M20250101000010CreateOperationLogsTable) Up() error {
	if !facades.Schema().HasTable("operation_logs") {
		return facades.Schema().Create("operation_logs", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.UnsignedBigInteger("admin_id").Nullable()
			table.String("method").Nullable()
			table.String("path").Nullable()
			table.String("ip").Nullable()
			table.String("user_agent").Nullable()
			table.Text("request").Nullable()
			table.Text("response").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.Text("error_msg").Nullable()
			table.Integer("duration").Nullable()
			table.Timestamps()
			table.Comment("操作日志表")
		})
	}

	return nil
}

func (r *M20250101000010CreateOperationLogsTable) Down() error {
	return facades.Schema().DropIfExists("operation_logs")
}
