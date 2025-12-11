package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000025AddLinkTypeToMenus struct{}

func (r *M20250101000025AddLinkTypeToMenus) Signature() string {
	return "20250101000025_add_link_type_to_menus"
}

func (r *M20250101000025AddLinkTypeToMenus) Up() error {
	if facades.Schema().HasTable("menus") {
		return facades.Schema().Table("menus", func(table schema.Blueprint) {
			// 链接类型：1-内部页面，2-外部链接
			if !facades.Schema().HasColumn("menus", "link_type") {
				table.UnsignedTinyInteger("link_type").Default(1).Comment("链接类型 1:内部页面 2:外部链接")
			}
			// 打开方式：1-iframe嵌套，2-新窗口打开（仅外部链接有效）
			if !facades.Schema().HasColumn("menus", "open_type") {
				table.UnsignedTinyInteger("open_type").Default(1).Comment("打开方式 1:iframe嵌套 2:新窗口打开")
			}
		})
	}

	return nil
}

func (r *M20250101000025AddLinkTypeToMenus) Down() error {
	if facades.Schema().HasTable("menus") {
		return facades.Schema().Table("menus", func(table schema.Blueprint) {
			if facades.Schema().HasColumn("menus", "link_type") {
				table.DropColumn("link_type")
			}
			if facades.Schema().HasColumn("menus", "open_type") {
				table.DropColumn("open_type")
			}
		})
	}

	return nil
}

