package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000019AddRequestToLoginLogsTable struct {
}

func (r *M20250101000019AddRequestToLoginLogsTable) Signature() string {
	return "20250101000019_add_request_to_login_logs_table"
}

func (r *M20250101000019AddRequestToLoginLogsTable) Up() error {
	if facades.Schema().HasTable("login_logs") {
		if !facades.Schema().HasColumn("login_logs", "request") {
			return facades.Schema().Table("login_logs", func(table schema.Blueprint) {
				table.Text("request").Nullable().Comment("请求数据")
			})
		}
	}

	return nil
}

func (r *M20250101000019AddRequestToLoginLogsTable) Down() error {
	if facades.Schema().HasTable("login_logs") {
		if facades.Schema().HasColumn("login_logs", "request") {
			return facades.Schema().Table("login_logs", func(table schema.Blueprint) {
				table.DropColumn("request")
			})
		}
	}

	return nil
}

