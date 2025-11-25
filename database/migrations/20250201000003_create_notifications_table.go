package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250201000003CreateNotificationsTable struct{}

func (m *M20250201000003CreateNotificationsTable) Signature() string {
	return "20250201000003_create_notifications_table"
}

func (m *M20250201000003CreateNotificationsTable) Up() error {
	if facades.Schema().HasTable("notifications") {
		return nil
	}

	return facades.Schema().Create("notifications", func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("title", 150).Comment("标题")
		table.Text("content").Comment("内容")
		table.String("type", 20).Default("announcement").Comment("类型: announcement|notice|message")
		table.UnsignedBigInteger("sender_id").Nullable().Comment("发送者ID")
		table.UnsignedBigInteger("receiver_id").Comment("接收者ID")
		table.Boolean("is_read").Default(false)
		table.Timestamp("read_at").Nullable()
		table.Timestamps()
		table.SoftDeletes()
		table.Index("receiver_id")
	})
}

func (m *M20250201000003CreateNotificationsTable) Down() error {
	return facades.Schema().DropIfExists("notifications")
}
