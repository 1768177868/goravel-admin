package listeners

import (
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/facades"

	"goravel/app/errors"
	"goravel/app/events"
	"goravel/app/support"
)

// SyncOrderElasticsearch 订单 ES 同步监听器（异步入队）。
type SyncOrderElasticsearch struct{}

func (receiver *SyncOrderElasticsearch) Signature() string {
	return "sync_order_elasticsearch_listener"
}

func (receiver *SyncOrderElasticsearch) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     true,
		Connection: "",
		Queue:      facades.Config().GetString("elasticsearch.sync_queue", "elasticsearch"),
	}
}

func (receiver *SyncOrderElasticsearch) Handle(args ...any) error {
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing order elasticsearch sync args")
	}

	var syncArgs events.OrderElasticsearchSyncArgs
	switch v := args[0].(type) {
	case events.OrderElasticsearchSyncArgs:
		syncArgs = v
	case map[string]any:
		syncArgs = events.OrderElasticsearchSyncArgs{
			OrderID: uintFromAny(v["order_id"]),
			OrderNo: stringFromAny(v["order_no"]),
			Op:      stringFromAny(v["op"]),
		}
	default:
		return errors.ErrInvalidArgument.WithMessage("invalid order elasticsearch sync args")
	}

	if syncArgs.OrderID == 0 {
		return errors.ErrInvalidArgument.WithMessage("invalid order_id")
	}
	if syncArgs.Op != "index" && syncArgs.Op != "delete" {
		return errors.ErrInvalidArgument.WithMessage("invalid op")
	}

	support.DispatchOrderElasticsearchSync(syncArgs.OrderID, syncArgs.OrderNo, syncArgs.Op)
	return nil
}

func uintFromAny(v any) uint {
	switch n := v.(type) {
	case uint:
		return n
	case int:
		if n > 0 {
			return uint(n)
		}
	case int64:
		if n > 0 {
			return uint(n)
		}
	case float64:
		if n > 0 {
			return uint(n)
		}
	}
	return 0
}

func stringFromAny(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
