package esorders

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/facades"

	"goravel/app/binding"
	"goravel/app/clients"
	"goravel/app/repositories"
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

	short := facades.Config().GetString("elasticsearch.orders_index", "orders")
	index := clients.ElasticsearchIndexName(facades.Config(), short)
	idx := NewOrderIndexer(es, index)

	if op == "delete" {
		return idx.DeleteByOrderID(ctx, orderID)
	}

	order, details, err := orderrepo.FindOrderWithDetails(orderID, orderNoHint)
	if err != nil {
		return err
	}
	return idx.IndexOrder(ctx, order, details)
}
