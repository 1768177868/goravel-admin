package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260423000200CreateSlowQueryLogsTable struct{}

func (r *M20260423000200CreateSlowQueryLogsTable) Signature() string {
	return "20260423000200_create_slow_query_logs_table"
}

func (r *M20260423000200CreateSlowQueryLogsTable) Up() error {
	if facades.Schema().HasTable("slow_query_logs") {
		return nil
	}

	return facades.Schema().Create("slow_query_logs", func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("trace_id", 120).Nullable().Comment("链路ID")
		table.Text("sql_text").Nullable()
		table.Text("normalized_sql").Nullable()
		table.String("sql_hash", 64).Nullable()
		table.Decimal("duration_ms").Default(0).Comment("耗时毫秒")
		table.BigInteger("rows_affected").Default(0)
		table.String("source", 64).Default("gorm-log")
		table.String("occurred_at", 32).Nullable()
		table.Timestamps()
		table.Index("trace_id")
		table.Index("sql_hash")
		table.Index("duration_ms")
		table.Index("created_at")
		table.Comment("慢SQL日志表")
	})
}

func (r *M20260423000200CreateSlowQueryLogsTable) Down() error {
	return facades.Schema().DropIfExists("slow_query_logs")
}
