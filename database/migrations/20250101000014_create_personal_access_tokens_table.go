package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000014CreatePersonalAccessTokensTable struct {
}

func (r *M20250101000014CreatePersonalAccessTokensTable) Signature() string {
	return "20250101000014_create_personal_access_tokens_table"
}

func (r *M20250101000014CreatePersonalAccessTokensTable) Up() error {
	if !facades.Schema().HasTable("personal_access_tokens") {
		return facades.Schema().Create("personal_access_tokens", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("tokenable_type").Comment("模型类型，如：admin")
			table.BigInteger("tokenable_id").Comment("模型ID")
			table.String("name").Default("").Comment("token名称")
			table.String("token", 64).Comment("token值（hash）")
			table.Text("abilities").Nullable().Comment("权限列表（JSON）")
			table.Timestamp("last_used_at").Nullable().Comment("最后使用时间")
			table.Timestamp("expires_at").Nullable().Comment("过期时间，NULL表示永不过期")
			table.Timestamps()
			table.Unique("token")
			table.Index("tokenable_type")
			table.Index("tokenable_id")
			table.Comment("个人访问令牌表")
		})
	}

	return nil
}

func (r *M20250101000014CreatePersonalAccessTokensTable) Down() error {
	return facades.Schema().DropIfExists("personal_access_tokens")
}

