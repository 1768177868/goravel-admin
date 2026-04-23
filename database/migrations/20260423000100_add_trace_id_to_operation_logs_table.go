package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260423000100AddTraceIdToOperationLogsTable struct{}

func (r *M20260423000100AddTraceIdToOperationLogsTable) Signature() string {
	return "20260423000100_add_trace_id_to_operation_logs_table"
}

func (r *M20260423000100AddTraceIdToOperationLogsTable) Up() error {
	if !facades.Schema().HasTable("operation_logs") {
		return nil
	}

	if facades.Schema().HasColumn("operation_logs", "trace_id") {
		return nil
	}

	return facades.Schema().Table("operation_logs", func(table schema.Blueprint) {
		table.String("trace_id", 120).Nullable().Comment("链路ID")
		table.Index("trace_id")
	})
}

func (r *M20260423000100AddTraceIdToOperationLogsTable) Down() error {
	if !facades.Schema().HasTable("operation_logs") {
		return nil
	}
	if !facades.Schema().HasColumn("operation_logs", "trace_id") {
		return nil
	}

	return facades.Schema().Table("operation_logs", func(table schema.Blueprint) {
		table.DropColumn("trace_id")
	})
}
