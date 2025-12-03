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
	if !facades.Schema().HasTable("attachments") {
		return nil
	}

	// 检查列是否已存在
	columns, err := facades.Schema().GetColumns("attachments")
	if err != nil {
		return err
	}

	hasDisplayName := false
	hasIndex := false
	for _, column := range columns {
		if column.Name == "display_name" {
			hasDisplayName = true
			break
		}
	}

	// 检查索引是否存在
	indexes, err := facades.Schema().GetIndexes("attachments")
	if err == nil {
		for _, index := range indexes {
			if index.Name == "display_name" {
				hasIndex = true
				break
			}
		}
	}

	// 如果列不存在，则添加
	if !hasDisplayName {
		return facades.Schema().Table("attachments", func(table schema.Blueprint) {
			table.String("display_name", 255).Nullable().Comment("显示名称（可编辑）").After("filename")
			if !hasIndex {
				table.Index("display_name")
			}
		})
	}

	return nil
}

func (r *M20250301000023AddDisplayNameToAttachments) Down() error {
	return facades.Schema().Table("attachments", func(table schema.Blueprint) {
		table.DropIndex("display_name")
		table.DropColumn("display_name")
	})
}

