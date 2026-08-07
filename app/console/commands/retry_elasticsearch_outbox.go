package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	esorders "goravel/app/elasticsearch/orders"
	"goravel/app/queuejobs"
	"goravel/app/utils"
)

// RetryElasticsearchOutbox 重试 pending/failed 的 ES 同步 outbox 记录。
type RetryElasticsearchOutbox struct{}

func (r *RetryElasticsearchOutbox) Signature() string {
	return "es:retry-outbox"
}

func (r *RetryElasticsearchOutbox) Description() string {
	return "重试 Elasticsearch 同步 outbox（pending/failed 记录）"
}

func (r *RetryElasticsearchOutbox) Extend() command.Extend {
	return command.Extend{
		Category: "elasticsearch",
		Flags: []command.Flag{
			&command.IntFlag{
				Name:    "limit",
				Value:   100,
				Aliases: []string{"l"},
				Usage:   "单次处理条数上限",
			},
			&command.IntFlag{
				Name:    "max-attempts",
				Value:   5,
				Usage:   "最大重试次数，超过后跳过",
			},
			&command.BoolFlag{
				Name:    "dispatch",
				Aliases: []string{"q"},
				Usage:   "重新入队（默认直接同步执行）",
			},
		},
	}
}

func (r *RetryElasticsearchOutbox) Handle(ctx console.Context) error {
	// 未开启 ES 订单同步或 outbox 时静默跳过（定时任务每小时跑一次，避免刷日志）
	if !esorders.SyncEnabled() {
		return nil
	}
	if !facades.Config().GetBool("elasticsearch.outbox_enabled", true) {
		return nil
	}

	limit := cast.ToInt(ctx.Option("limit"))
	maxAttempts := cast.ToInt(ctx.Option("max-attempts"))
	dispatch := cast.ToBool(ctx.Option("dispatch"))

	records, err := utils.ListElasticsearchOutboxRetryable(limit, maxAttempts)
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	if len(records) == 0 {
		pending, failed := utils.CountElasticsearchOutboxBacklog()
		ctx.Info(fmt.Sprintf("无待重试记录（pending=%d failed=%d）", pending, failed))
		return nil
	}

	ctx.Info(fmt.Sprintf("开始重试 %d 条 outbox 记录...", len(records)))
	var okCount, failCount int
	runCtx := context.Background()

	for _, record := range records {
		payload := map[string]any{
			"order_id": record.EntityID,
			"op":       record.Op,
		}
		if record.EntityKey != "" {
			payload["order_no"] = record.EntityKey
		}
		if record.Payload != "" {
			var parsed map[string]any
			if err := json.Unmarshal([]byte(record.Payload), &parsed); err == nil && parsed != nil {
				payload = parsed
			}
		}

		if dispatch {
			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				utils.MarkElasticsearchOutboxFailedByID(record.ID, err.Error())
				failCount++
				continue
			}
			qargs := []queue.Arg{{Type: "string", Value: string(jsonBytes)}}
			queueName := facades.Config().GetString("elasticsearch.sync_queue", "elasticsearch")
			if err := facades.Queue().Job(&queuejobs.SyncOrderToElasticsearch{}, qargs).OnQueue(queueName).Dispatch(); err != nil {
				utils.MarkElasticsearchOutboxFailedByID(record.ID, err.Error())
				failCount++
				continue
			}
			okCount++
			continue
		}

		orderID := cast.ToUint(payload["order_id"])
		op, _ := utils.GetString(payload, "op")
		orderNo, _ := utils.GetString(payload, "order_no")
		if orderID == 0 || (op != "index" && op != "delete") {
			utils.MarkElasticsearchOutboxFailedByID(record.ID, "invalid outbox payload")
			failCount++
			continue
		}
		if err := esorders.PushOrderToElasticsearch(runCtx, orderID, orderNo, op); err != nil {
			utils.MarkElasticsearchOutboxFailedByID(record.ID, err.Error())
			failCount++
			continue
		}
		utils.MarkElasticsearchOutboxProcessedByID(record.ID)
		okCount++
	}

	if failCount > 0 {
		ctx.Warning(fmt.Sprintf("完成：成功 %d，失败 %d", okCount, failCount))
	} else {
		ctx.Success(fmt.Sprintf("完成：成功 %d", okCount))
	}
	return nil
}
