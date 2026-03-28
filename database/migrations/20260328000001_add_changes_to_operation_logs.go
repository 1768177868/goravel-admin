package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260328000001AddChangesToOperationLogs struct {
}

func (r *M20260328000001AddChangesToOperationLogs) Signature() string {
	return "20260328000001_add_changes_to_operation_logs"
}

func (r *M20260328000001AddChangesToOperationLogs) Up() error {
	if !facades.Schema().HasTable("operation_logs") {
		return nil
	}

	columns, err := facades.Schema().GetColumns("operation_logs")
	if err != nil {
		return err
	}

	hasChanges := false
	for _, column := range columns {
		if column.Name == "changes" {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return facades.Schema().Table("operation_logs", func(table schema.Blueprint) {
			table.Text("changes").Nullable().Comment("变更详情(JSON diff)")
		})
	}

	return nil
}

func (r *M20260328000001AddChangesToOperationLogs) Down() error {
	if facades.Schema().HasTable("operation_logs") {
		return facades.Schema().Table("operation_logs", func(table schema.Blueprint) {
			table.DropColumn("changes")
		})
	}
	return nil
}
