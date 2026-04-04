package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260404000001CreatePositionsTable struct {
}

func (r *M20260404000001CreatePositionsTable) Signature() string {
	return "20260404000001_create_positions_table"
}

func (r *M20260404000001CreatePositionsTable) Up() error {
	if facades.Schema().HasTable("positions") {
		return nil
	}
	return facades.Schema().Create("positions", func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("name").Comment("岗位名称")
		table.String("code").Nullable().Comment("岗位编码")
		table.UnsignedTinyInteger("status").Default(1).Comment("状态 1:启用 0:禁用")
		table.Integer("sort").Default(0).Comment("排序")
		table.String("remark").Nullable().Comment("备注")
		table.Timestamps()
		table.Comment("岗位表")
	})
}

func (r *M20260404000001CreatePositionsTable) Down() error {
	return facades.Schema().DropIfExists("positions")
}
