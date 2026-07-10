package esorders

import (
	"context"
	"fmt"

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

// DeleteByOrderNo 按订单号删除文档（_id 与 order_no 一致，避免分表 id 冲突）。
func (w *OrderIndexer) DeleteByOrderNo(ctx context.Context, orderNo string) error {
	if orderNo == "" {
		return fmt.Errorf("order_no required for elasticsearch delete")
	}
	return clients.ElasticsearchDeleteDocument(ctx, w.es, w.index, orderNo)
}

// IndexOrder 写入或覆盖订单文档。
func (w *OrderIndexer) IndexOrder(ctx context.Context, order *models.Order, details []models.OrderDetail) error {
	if order.OrderNo == "" {
		return fmt.Errorf("order_no required for elasticsearch indexing")
	}
	doc := OrderDocument(order, details)
	return clients.ElasticsearchIndexValue(ctx, w.es, w.index, order.OrderNo, doc)
}
