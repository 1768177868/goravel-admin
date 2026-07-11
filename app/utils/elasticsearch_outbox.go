package utils

import (
	"context"
	"encoding/json"
	"time"

	appfacades "goravel/app/facades"

	"github.com/goravel/framework/facades"
)

const (
	esOutboxStatusPending   = "pending"
	esOutboxStatusProcessed = "processed"
	esOutboxStatusFailed    = "failed"
)

func elasticsearchOutboxEnabled() bool {
	return facades.Config().GetBool("elasticsearch.outbox_enabled", true)
}

// CreateElasticsearchSyncOutbox 写入 ES 同步 outbox 记录。
func CreateElasticsearchSyncOutbox(orderID uint, orderNo, op string, payload map[string]any) {
	if !elasticsearchOutboxEnabled() || !facades.Schema().HasTable("elasticsearch_sync_outbox") {
		return
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		facades.Log().Errorf("elasticsearch outbox marshal failed: %v", err)
		return
	}

	now := time.Now().UTC()
	record := map[string]any{
		"entity_type": "order",
		"entity_id":   orderID,
		"entity_key":  orderNo,
		"op":          op,
		"payload":     string(payloadJSON),
		"status":      esOutboxStatusPending,
		"attempts":    0,
		"created_at":  now,
		"updated_at":  now,
	}

	if err := appfacades.OrmQuery(context.Background()).Table("elasticsearch_sync_outbox").Create(record); err != nil {
		facades.Log().Errorf("elasticsearch outbox create failed: order_id=%d err=%v", orderID, err)
	}
}

func findLatestPendingOutboxID(orderID uint, op string) uint {
	if orderID == 0 || !facades.Schema().HasTable("elasticsearch_sync_outbox") {
		return 0
	}
	var row struct {
		ID uint `gorm:"column:id"`
	}
	if err := appfacades.OrmQuery(context.Background()).Table("elasticsearch_sync_outbox").
		Where("entity_type", "order").
		Where("entity_id", orderID).
		Where("op", op).
		Where("status", esOutboxStatusPending).
		Order("id DESC").
		Limit(1).
		First(&row); err != nil || row.ID == 0 {
		return 0
	}
	return row.ID
}

func updateOutboxByID(id uint, data map[string]any) {
	if id == 0 {
		return
	}
	data["updated_at"] = time.Now().UTC()
	_, _ = appfacades.OrmQuery(context.Background()).Table("elasticsearch_sync_outbox").Where("id", id).Update(data)
}

// MarkElasticsearchSyncOutboxProcessed 标记最近一条 pending outbox 为已处理。
func MarkElasticsearchSyncOutboxProcessed(orderID uint, op string) {
	id := findLatestPendingOutboxID(orderID, op)
	if id == 0 {
		return
	}
	now := time.Now().UTC()
	updateOutboxByID(id, map[string]any{
		"status":       esOutboxStatusProcessed,
		"processed_at": now,
	})
}

// MarkElasticsearchSyncOutboxFailed 标记最近一条 pending outbox 为失败。
func MarkElasticsearchSyncOutboxFailed(orderID uint, op, errMsg string) {
	id := findLatestPendingOutboxID(orderID, op)
	if id == 0 {
		return
	}
	var attempts int
	_ = appfacades.OrmQuery(context.Background()).Table("elasticsearch_sync_outbox").Where("id", id).Pluck("attempts", &attempts)
	updateOutboxByID(id, map[string]any{
		"status":     esOutboxStatusFailed,
		"attempts":   attempts + 1,
		"last_error": errMsg,
	})
}
