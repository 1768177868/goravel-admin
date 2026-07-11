package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260711000001CreateElasticsearchSyncOutboxTable struct{}

func (r *M20260711000001CreateElasticsearchSyncOutboxTable) Signature() string {
	return "20260711000001_create_elasticsearch_sync_outbox_table"
}

func (r *M20260711000001CreateElasticsearchSyncOutboxTable) Up() error {
	if facades.Schema().HasTable("elasticsearch_sync_outbox") {
		return nil
	}
	return facades.Schema().Create("elasticsearch_sync_outbox", func(table schema.Blueprint) {
		table.BigIncrements("id")
		table.String("entity_type", 50).Comment("实体类型，如 order")
		table.UnsignedBigInteger("entity_id").Comment("实体 ID")
		table.String("entity_key", 100).Nullable().Comment("业务键，如 order_no")
		table.String("op", 20).Comment("操作 index/delete")
		table.Text("payload").Nullable().Comment("队列载荷 JSON")
		table.String("status", 20).Default("pending").Comment("pending/processed/failed")
		table.UnsignedInteger("attempts").Default(0).Comment("处理尝试次数")
		table.Text("last_error").Nullable().Comment("最近一次失败原因")
		table.Timestamp("processed_at").Nullable().Comment("处理完成时间")
		table.Timestamps()
		table.Index("entity_type", "entity_id", "status")
		table.Index("status")
		table.Comment("Elasticsearch 同步 outbox")
	})
}

func (r *M20260711000001CreateElasticsearchSyncOutboxTable) Down() error {
	return facades.Schema().DropIfExists("elasticsearch_sync_outbox")
}
