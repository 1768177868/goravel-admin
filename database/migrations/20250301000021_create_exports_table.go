package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250301000021CreateExportsTable struct {
}

func (r *M20250301000021CreateExportsTable) Signature() string {
	return "20250301000021_create_exports_table"
}

func (r *M20250301000021CreateExportsTable) Up() error {
	if !facades.Schema().HasTable("exports") {
		return facades.Schema().Create("exports", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.UnsignedBigInteger("admin_id").Nullable().Comment("管理员ID")
			table.String("disk", 50).Nullable().Comment("存储驱动")
			table.String("path", 255).Nullable().Comment("文件路径")
			table.String("filename", 255).Nullable().Comment("文件名")
			table.String("extension", 20).Nullable().Comment("文件后缀")
			table.BigInteger("size").Nullable().Comment("文件大小(字节)")
			table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:成功 0:失败")
			table.Timestamps()

			table.Index("admin_id")
			table.Comment("导出记录表")
		})
	}
	return nil
}

func (r *M20250301000021CreateExportsTable) Down() error {
	return facades.Schema().DropIfExists("exports")
}


