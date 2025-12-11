package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000026ModifyMenusPathLength struct{}

func (r *M20250101000026ModifyMenusPathLength) Signature() string {
	return "20250101000026_modify_menus_path_length"
}

func (r *M20250101000026ModifyMenusPathLength) Up() error {
	if !facades.Schema().HasTable("menus") {
		return nil
	}

	// 检查 path 字段是否存在
	if !facades.Schema().HasColumn("menus", "path") {
		return nil
	}

	// 使用 Table 方法和 Change() 修饰符修改字段长度
	// 框架会自动处理不同数据库的差异
	return facades.Schema().Table("menus", func(table schema.Blueprint) {
		table.String("path", 1000).Nullable().Comment("路由路径").Change()
	})
}

func (r *M20250101000026ModifyMenusPathLength) Down() error {
	if !facades.Schema().HasTable("menus") {
		return nil
	}

	// 检查 path 字段是否存在
	if !facades.Schema().HasColumn("menus", "path") {
		return nil
	}

	// 回滚时将 path 字段长度改回 255
	return facades.Schema().Table("menus", func(table schema.Blueprint) {
		table.String("path", 255).Nullable().Comment("路由路径").Change()
	})
}
