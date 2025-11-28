package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000016CreateBlacklistsTable struct {
}

func (r *M20250101000016CreateBlacklistsTable) Signature() string {
	return "20250101000016_create_blacklists_table"
}

func (r *M20250101000016CreateBlacklistsTable) Up() error {
	if !facades.Schema().HasTable("blacklists") {
		return facades.Schema().Create("blacklists", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("ip").Default("").Comment("IP地址或IP段，支持单个IP、CIDR格式、IP范围，多个用逗号分隔")
			table.String("remark").Nullable().Comment("备注")
			table.TinyInteger("status").Default(1).Comment("状态 1:启用 0:禁用")
			table.Timestamps()
			table.Index("ip")
			table.Index("status")
			table.Comment("IP黑名单表")
		})
	}
	return nil
}

func (r *M20250101000016CreateBlacklistsTable) Down() error {
	return facades.Schema().DropIfExists("blacklists")
}

