package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000015CreateConfigsTable struct {
}

func (r *M20250101000015CreateConfigsTable) Signature() string {
	return "20250101000015_create_configs_table"
}

func (r *M20250101000015CreateConfigsTable) Up() error {
	if !facades.Schema().HasTable("configs") {
		return facades.Schema().Create("configs", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("group").Default("").Comment("配置分组 website:网站配置 email:邮箱配置")
			table.String("key").Default("").Comment("配置键")
			table.Text("value").Nullable().Comment("配置值")
			table.String("label").Nullable().Comment("配置标签")
			table.String("type").Default("input").Comment("配置类型 input:text:textarea:select:switch")
			table.Integer("sort").Default(0).Comment("排序")
			table.String("remark").Nullable().Comment("备注")
			table.Timestamps()
			table.Index("group")
			table.Index("key")
			table.Comment("配置表")
		})
	}

	return nil
}

func (r *M20250101000015CreateConfigsTable) Down() error {
	return facades.Schema().DropIfExists("configs")
}

