package events

import "github.com/goravel/framework/contracts/event"

// OrderElasticsearchSyncArgs 订单 ES 同步事件参数。
type OrderElasticsearchSyncArgs struct {
	OrderID uint
	OrderNo string
	Op      string // index | delete
}

// OrderElasticsearchSync 订单写入/删除 Elasticsearch 的异步同步事件。
type OrderElasticsearchSync struct{}

func (receiver *OrderElasticsearchSync) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}
