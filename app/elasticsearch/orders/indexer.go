package esorders

import (
	"context"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"

	"goravel/app/clients"
	"goravel/app/models"
)

// OrderIndexer 订单索引的 ES 写入/删除（与领域文档 OrderDocument 分离）。
type OrderIndexer struct {
	es    *elasticsearch.Client
	index string
}

// NewOrderIndexer 使用已解析的完整索引名（含 prefix）。
func NewOrderIndexer(es *elasticsearch.Client, indexFullName string) *OrderIndexer {
	return &OrderIndexer{es: es, index: indexFullName}
}

// DeleteByOrderID 按订单主键删除文档。
func (w *OrderIndexer) DeleteByOrderID(ctx context.Context, orderID uint) error {
	docID := strconv.FormatUint(uint64(orderID), 10)
	return clients.ElasticsearchDeleteDocument(ctx, w.es, w.index, docID)
}

// IndexOrder 写入或覆盖订单文档。
func (w *OrderIndexer) IndexOrder(ctx context.Context, order *models.Order, details []models.OrderDetail) error {
	docID := strconv.FormatUint(uint64(order.ID), 10)
	doc := OrderDocument(order, details)
	return clients.ElasticsearchIndexValue(ctx, w.es, w.index, docID, doc)
}
