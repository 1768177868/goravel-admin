package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000018AddTitleToOperationLogs struct {
}

func (r *M20250101000018AddTitleToOperationLogs) Signature() string {
	return "20250101000018_add_title_to_operation_logs"
}

func (r *M20250101000018AddTitleToOperationLogs) Up() error {
	if !facades.Schema().HasTable("operation_logs") {
		return nil
	}

	// 检查列是否已存在
	columns, err := facades.Schema().GetColumns("operation_logs")
	if err != nil {
		return err
	}

	hasTitle := false
	for _, column := range columns {
		if column.Name == "title" {
			hasTitle = true
			break
		}
	}

	// 如果列不存在，则添加
	if !hasTitle {
		return facades.Schema().Table("operation_logs", func(table schema.Blueprint) {
			table.String("title", 255).Nullable().Comment("操作标题")
		})
	}

	return nil
}

func (r *M20250101000018AddTitleToOperationLogs) Down() error {
	if facades.Schema().HasTable("operation_logs") {
		return facades.Schema().Table("operation_logs", func(table schema.Blueprint) {
			table.DropColumn("title")
		})
	}
	return nil
}

