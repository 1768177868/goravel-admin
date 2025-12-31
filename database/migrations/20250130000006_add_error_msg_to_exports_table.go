package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250130000006AddErrorMsgToExportsTable struct {
}

func (r *M20250130000006AddErrorMsgToExportsTable) Signature() string {
	return "20250130000006_add_error_msg_to_exports_table"
}

func (r *M20250130000006AddErrorMsgToExportsTable) Up() error {
	if facades.Schema().HasTable("exports") {
		return facades.Schema().Table("exports", func(table schema.Blueprint) {
			if !facades.Schema().HasColumn("exports", "error_msg") {
				table.Text("error_msg").Nullable().Comment("错误信息")
			}
			// 更新 status 默认值为 0（处理中）
			// 注意：Goravel 的 Schema 可能不支持直接修改默认值，需要在应用层处理
		})
	}
	return nil
}

func (r *M20250130000006AddErrorMsgToExportsTable) Down() error {
	if facades.Schema().HasTable("exports") {
		return facades.Schema().Table("exports", func(table schema.Blueprint) {
			if facades.Schema().HasColumn("exports", "error_msg") {
				table.DropColumn("error_msg")
			}
		})
	}
	return nil
}

