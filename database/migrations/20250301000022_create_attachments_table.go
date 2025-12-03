package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250301000022CreateAttachmentsTable struct {
}

func (r *M20250301000022CreateAttachmentsTable) Signature() string {
	return "20250301000022_create_attachments_table"
}

func (r *M20250301000022CreateAttachmentsTable) Up() error {
	if !facades.Schema().HasTable("attachments") {
		return facades.Schema().Create("attachments", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.UnsignedBigInteger("admin_id").Nullable().Comment("管理员ID")
			table.String("disk", 50).Nullable().Comment("存储驱动")
			table.String("path", 500).Nullable().Comment("文件路径")
			table.String("filename", 255).Nullable().Comment("原始文件名")
			table.String("extension", 20).Nullable().Comment("文件后缀")
			table.String("mime_type", 100).Nullable().Comment("MIME类型")
			table.BigInteger("size").Nullable().Comment("文件大小(字节)")
			table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:成功 0:失败 2:上传中")
			table.String("file_type", 20).Nullable().Comment("文件类型 image/video/document/other")
			table.String("chunk_id", 100).Nullable().Comment("分片上传ID（用于断点续传）")
			table.Timestamps()

			table.Index("admin_id")
			table.Index("chunk_id")
			table.Comment("附件表")
		})
	}
	return nil
}

func (r *M20250301000022CreateAttachmentsTable) Down() error {
	return facades.Schema().DropIfExists("attachments")
}

