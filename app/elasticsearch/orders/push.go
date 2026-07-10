package esorders

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/facades"

	"goravel/app/binding"
	orderrepo "goravel/app/repositories"
)

// PushOrderToElasticsearch 将单条订单写入或从 ES 删除（需集群已启用且容器已绑定客户端）。
func PushOrderToElasticsearch(ctx context.Context, orderID uint, orderNoHint string, op string) error {
	raw, err := facades.App().Make(binding.ElasticsearchClient)
	if err != nil {
		return err
	}
	es, ok := raw.(*elasticsearch.Client)
	if !ok || es == nil {
		return fmt.Errorf("elasticsearch client not available")
	}

	index := OrdersIndexName()
	idx := NewOrderIndexer(es, index)

	if op == "delete" {
		orderNo := orderNoHint
		if orderNo == "" {
			order, err := orderrepo.FindOrderByID(ctx, orderID)
			if err != nil {
				return err
			}
			orderNo = order.OrderNo
		}
		return idx.DeleteByOrderNo(ctx, orderNo)
	}

	if err := EnsureOrdersIndex(ctx); err != nil {
		return err
	}

	order, details, err := orderrepo.FindOrderWithDetails(ctx, orderID, orderNoHint)
	if err != nil {
		return err
	}
	return idx.IndexOrder(ctx, order, details)
}
