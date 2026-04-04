package support

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"

	esorders "goravel/app/elasticsearch/orders"
	"goravel/app/queuejobs"
)

// DispatchOrderElasticsearchSync 异步入队同步任务（op 为 index 或 delete）。
func DispatchOrderElasticsearchSync(orderID uint, orderNo string, op string) {
	if !esorders.SyncEnabled() {
		return
	}
	payload := map[string]any{
		"order_id": orderID,
		"op":       op,
	}
	if orderNo != "" {
		payload["order_no"] = orderNo
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		facades.Log().Errorf("order es sync marshal failed: %v", err)
		return
	}
	qargs := []queue.Arg{{Type: "string", Value: string(jsonBytes)}}
	queueName := facades.Config().GetString("elasticsearch.sync_queue", "elasticsearch")
	if err := facades.Queue().Job(&queuejobs.SyncOrderToElasticsearch{}, qargs).OnQueue(queueName).Dispatch(); err != nil {
		facades.Log().Errorf("order es sync dispatch failed: order_id=%d op=%s err=%v", orderID, op, err)
	}
}
