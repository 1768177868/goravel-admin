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
		table.UnsignedBigInteger("admin_id").Comment("管理员ID")
		table.String("title").Comment("标题")
		table.Text("content").Nullable().Comment("内容")
		table.UnsignedTinyInteger("status").Default(1).Comment("0:未发布 1:发布")
		table.Index("admin_id")
		table.Index("status")
		table.Timestamps()
		table.SoftDeletes()
	})
}
func (m *M20260115152848ArticleTable) Down() error {
	return facades.Schema().DropIfExists("articles")
}
