package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250131000020AddNoCacheToMenus struct{}

func (r *M20250131000020AddNoCacheToMenus) Signature() string {
	return "20250131000020_add_no_cache_to_menus"
}

func (r *M20250131000020AddNoCacheToMenus) Up() error {
	if facades.Schema().HasTable("menus") {
		return facades.Schema().Table("menus", func(table schema.Blueprint) {
			if !facades.Schema().HasColumn("menus", "no_cache") {
				table.UnsignedTinyInteger("no_cache").Default(0).Comment("是否缓存 1:否-每次进页面刷新 0:是")
			}
		})
	}

	return nil
}

func (r *M20250131000020AddNoCacheToMenus) Down() error {
	if facades.Schema().HasTable("menus") {
		return facades.Schema().Table("menus", func(table schema.Blueprint) {
			if facades.Schema().HasColumn("menus", "no_cache") {
				table.DropColumn("no_cache")
			}
		})
	}

	return nil
}
