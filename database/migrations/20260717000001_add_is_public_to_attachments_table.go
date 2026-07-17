package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260717000001AddIsPublicToAttachmentsTable struct {
}

func (r *M20260717000001AddIsPublicToAttachmentsTable) Signature() string {
	return "20260717000001_add_is_public_to_attachments_table"
}

func (r *M20260717000001AddIsPublicToAttachmentsTable) Up() error {
	if !facades.Schema().HasTable("attachments") {
		return nil
	}

	if facades.Schema().HasColumn("attachments", "is_public") {
		return nil
	}

	if err := facades.Schema().Table("attachments", func(table schema.Blueprint) {
		table.UnsignedTinyInteger("is_public").Default(1).Comment("是否公开 1:公开 0:私有")
		table.Index("is_public")
	}); err != nil {
		return err
	}

	return nil
}

func (r *M20260717000001AddIsPublicToAttachmentsTable) Down() error {
	if !facades.Schema().HasTable("attachments") {
		return nil
	}
	if !facades.Schema().HasColumn("attachments", "is_public") {
		return nil
	}

	hasIndex := false
	indexes, err := facades.Schema().GetIndexes("attachments")
	if err == nil {
		for _, index := range indexes {
			if index.Name == "is_public" {
				hasIndex = true
				break
			}
		}
	}

	return facades.Schema().Table("attachments", func(table schema.Blueprint) {
		if hasIndex {
			table.DropIndex("is_public")
		}
		table.DropColumn("is_public")
	})
}
