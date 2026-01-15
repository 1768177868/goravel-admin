package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260115152848ArticleTable struct {
}

func (m *M20260115152848ArticleTable) Signature() string {
	return "20260115152848_create_articles_table"
}

func (m *M20260115152848ArticleTable) Up() error {
	return facades.Schema().Create("articles", func(table schema.Blueprint) {
		table.ID()

		table.String("name").Comment("名称")
		table.String("status").Comment("状态")
		table.Timestamps()
		table.SoftDeletes()
	})
}
func (m *M20260115152848ArticleTable) Down() error {
	return facades.Schema().DropIfExists("articles")
}
