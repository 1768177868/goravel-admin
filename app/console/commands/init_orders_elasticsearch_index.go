package commands

import (
	"context"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"

	esorders "goravel/app/elasticsearch/orders"
)

// InitOrdersElasticsearchIndex 创建订单 ES 索引及 mapping（若不存在）。
type InitOrdersElasticsearchIndex struct{}

func (r *InitOrdersElasticsearchIndex) Signature() string {
	return "es:init-orders-index"
}

func (r *InitOrdersElasticsearchIndex) Description() string {
	return "创建订单 Elasticsearch 索引及 mapping（需 ELASTICSEARCH_ENABLED=true）"
}

func (r *InitOrdersElasticsearchIndex) Extend() command.Extend {
	return command.Extend{Category: "elasticsearch"}
}

func (r *InitOrdersElasticsearchIndex) Handle(ctx console.Context) error {
	if !facades.Config().GetBool("elasticsearch.enabled", false) {
		ctx.Warning("未启用 Elasticsearch。请在 .env 设置 ELASTICSEARCH_ENABLED=true。")
		return nil
	}

	index := esorders.OrdersIndexName()
	if err := esorders.InitOrdersIndex(context.Background()); err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("订单索引已就绪: " + index)
	return nil
}
