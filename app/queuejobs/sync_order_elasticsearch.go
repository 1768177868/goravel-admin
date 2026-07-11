package queuejobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	esorders "goravel/app/elasticsearch/orders"
	"goravel/app/errors"
	"goravel/app/utils"
)

// SyncOrderToElasticsearch 订单 ES 同步队列任务（参数为 map：order_id, op, 可选 order_no）。
// 独立包避免 jobs 与 services 循环引用。
type SyncOrderToElasticsearch struct{}

func (r *SyncOrderToElasticsearch) Signature() string {
	return "sync_order_elasticsearch"
}

func (r *SyncOrderToElasticsearch) Handle(args ...any) error {
	if !esorders.SyncEnabled() {
		return nil
	}
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing sync args")
	}
	var m map[string]any
	switch v := args[0].(type) {
	case map[string]any:
		m = v
	case string:
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return errors.ErrInvalidArgument.WithMessage("sync args json invalid")
		}
	default:
		return errors.ErrInvalidArgument.WithMessage("sync args must be map or json string")
	}
	if m == nil {
		return errors.ErrInvalidArgument.WithMessage("sync args empty")
	}
	orderID := cast.ToUint(m["order_id"])
	if orderID == 0 {
		return errors.ErrInvalidArgument.WithMessage("invalid order_id")
	}
	op, _ := utils.GetString(m, "op")
	if op != "index" && op != "delete" {
		return fmt.Errorf("invalid op: %s", op)
	}
	orderNo, _ := utils.GetString(m, "order_no")
	ctx := context.Background()
	if err := esorders.PushOrderToElasticsearch(ctx, orderID, orderNo, op); err != nil {
		facades.Log().Errorf("SyncOrderToElasticsearch failed: order_id=%d op=%s err=%v", orderID, op, err)
		utils.MarkElasticsearchSyncOutboxFailed(orderID, op, err.Error())
		return err
	}
	utils.MarkElasticsearchSyncOutboxProcessed(orderID, op)
	return nil
}
