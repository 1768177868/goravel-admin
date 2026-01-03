package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250301000024AddTypeToExportsTable struct {
}

func (r *M20250301000024AddTypeToExportsTable) Signature() string {
	return "20250301000024_add_type_to_exports_table"
}

func (r *M20250301000024AddTypeToExportsTable) Up() error {
	if facades.Schema().HasTable("exports") {
		return facades.Schema().Table("exports", func(table schema.Blueprint) {
			if !facades.Schema().HasColumn("exports", "type") {
				table.String("type", 50).Nullable().Comment("导出类型 orders:订单导出 admins:管理员导出")
			}
		})
	}
	return nil
}

func (r *M20250301000024AddTypeToExportsTable) Down() error {
	if facades.Schema().HasTable("exports") {
		return facades.Schema().Table("exports", func(table schema.Blueprint) {
			if facades.Schema().HasColumn("exports", "type") {
				table.DropColumn("type")
			}
		})
	}
	return nil
}

