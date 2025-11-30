package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250301000023AddDisplayNameToAttachments struct {
}

func (r *M20250301000023AddDisplayNameToAttachments) Signature() string {
	return "20250301000023_add_display_name_to_attachments"
}

func (r *M20250301000023AddDisplayNameToAttachments) Up() error {
	return facades.Schema().Table("attachments", func(table schema.Blueprint) {
		table.String("display_name", 255).Nullable().Comment("显示名称（可编辑）").After("filename")
		table.Index("display_name")
	})
}

func (r *M20250301000023AddDisplayNameToAttachments) Down() error {
	return facades.Schema().Table("attachments", func(table schema.Blueprint) {
		table.DropIndex("display_name")
		table.DropColumn("display_name")
	})
}

