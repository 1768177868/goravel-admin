package migrations

import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/contracts/database/schema"
)

type CreateArticleTable struct {
}

func (m *CreateArticleTable) Signature() string {
	return "create_articles_table"
}

func (m *CreateArticleTable) Up() {
	facades.Schema().Create("articles", func(table schema.Table) {
		table.ID()

		table.String("name").Comment("名称")

		table.String("status").Comment("状态")

		table.Timestamps()
		table.SoftDeletes()
	})
}
func (m *CreateArticleTable) Down() {
	facades.Schema().DropIfExists("articles")
}
