package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250201000016AddTraceIdToSystemLogsTable struct {
}

func (r *M20250201000016AddTraceIdToSystemLogsTable) Signature() string {
	return "20250201000016_add_trace_id_to_system_logs_table"
}

func (r *M20250201000016AddTraceIdToSystemLogsTable) Up() error {
	if !facades.Schema().HasTable("system_logs") {
		return nil
	}

	// 如果列不存在，则添加
	if !facades.Schema().HasColumn("system_logs", "trace_id") {
		return facades.Schema().Table("system_logs", func(table schema.Blueprint) {
			table.String("trace_id").Nullable().Comment("链路ID")
		})
	}

	return nil
}

func (r *M20250201000016AddTraceIdToSystemLogsTable) Down() error {
	return facades.Schema().Table("system_logs", func(table schema.Blueprint) {
		table.DropColumn("trace_id")
	})
}
