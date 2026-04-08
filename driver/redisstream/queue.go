package redisstream

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	frameworkerrors "github.com/goravel/framework/errors"

	"github.com/goravel/framework/contracts/config"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue/utils"
	"github.com/redis/go-redis/v9"
)

var _ contractsqueue.Driver = &Queue{}

type Queue struct {
	ctx         context.Context
	client      redis.UniversalClient
	jobStorer   contractsqueue.JobStorer
	json        contractsfoundation.Json
	queueKey    *queueKey
	group       string
	consumer    string
	block       time.Duration
	retryAfter  time.Duration
	claimCount  int64
	deleteOnAck bool
	streamMax   int64
}

func New(connection string) (contractsqueue.Driver, error) {
	app := facades.App()

	return NewQueue(context.Background(), app.MakeConfig(), app.MakeQueue(), app.GetJson(), connection)
}

func NewQueue(ctx context.Context, cfg config.Config, queue contractsqueue.Queue, json contractsfoundation.Json, connection string) (*Queue, error) {
	clientConnection := cfg.GetString(fmt.Sprintf("queue.connections.%s.connection", connection), "default")
	client, err := GetClient(cfg, clientConnection)
	if err != nil {
		return nil, fmt.Errorf("init redis client failed: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("redis client is nil for queue connection [%s]", connection)
	}

	group := cfg.GetString(fmt.Sprintf("queue.connections.%s.group", connection), "goravel")
	blockMS := cfg.GetInt(fmt.Sprintf("queue.connections.%s.block_ms", connection), 1000)
	retryAfterSeconds := cfg.GetInt(fmt.Sprintf("queue.connections.%s.retry_after", connection), 90)
	claimCount := int64(cfg.GetInt(fmt.Sprintf("queue.connections.%s.claim_count", connection), 10))
	if claimCount <= 0 {
		claimCount = 10
	}

	consumer := cfg.GetString(fmt.Sprintf("queue.connections.%s.consumer", connection), "")
	if consumer == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown-host"
		}
		consumer = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}

	return &Queue{
		ctx:         ctx,
		client:      client,
		jobStorer:   queue.JobStorer(),
		json:        json,
		queueKey:    newQueueKey(cfg.GetString("app.name", "goravel"), connection),
		group:       group,
		consumer:    consumer,
		block:       time.Duration(blockMS) * time.Millisecond,
		retryAfter:  time.Duration(retryAfterSeconds) * time.Second,
		claimCount:  claimCount,
		deleteOnAck: cfg.GetBool(fmt.Sprintf("queue.connections.%s.delete_on_ack", connection), false),
		streamMax:   int64(cfg.GetInt(fmt.Sprintf("queue.connections.%s.stream_max_len", connection), 0)),
	}, nil
}

func (q *Queue) Driver() string {
	return contractsqueue.DriverCustom
}

func (q *Queue) Push(task contractsqueue.Task, queue string) error {
	if !task.Delay.IsZero() {
		return q.later(task.Delay, task, queue)
	}

	payload, err := utils.TaskToJson(task, q.json)
	if err != nil {
		return err
	}

	args := &redis.XAddArgs{
		Stream: q.queueKey.stream(queue),
		Values: map[string]any{
			"payload": payload,
		},
	}
	if q.streamMax > 0 {
		args.MaxLen = q.streamMax
		args.Approx = true
	}

	return q.client.XAdd(q.ctx, args).Err()
}

func (q *Queue) Pop(queue string) (contractsqueue.ReservedJob, error) {
	stream := q.queueKey.stream(queue)

	if err := q.ensureGroup(stream); err != nil {
		return nil, err
	}
	if err := q.migrateDelayedJobs(queue); err != nil {
		return nil, err
	}

	claimed, _, err := q.client.XAutoClaim(q.ctx, &redis.XAutoClaimArgs{
		Stream:   stream,
		Group:    q.group,
		Consumer: q.consumer,
		MinIdle:  q.retryAfter,
		Start:    "0-0",
		Count:    q.claimCount,
	}).Result()
	if err != nil && !strings.Contains(err.Error(), "NOGROUP") {
		return nil, err
	}
	if len(claimed) > 0 {
		return q.messageToReservedJob(stream, claimed[0])
	}

	streams, err := q.client.XReadGroup(q.ctx, &redis.XReadGroupArgs{
		Group:    q.group,
		Consumer: q.consumer,
		Streams:  []string{stream, ">"},
		Count:    1,
		Block:    q.block,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, frameworkerrors.QueueDriverNoJobFound.Args(stream)
		}
		return nil, err
	}
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, frameworkerrors.QueueDriverNoJobFound.Args(stream)
	}

	return q.messageToReservedJob(stream, streams[0].Messages[0])
}

func (q *Queue) ensureGroup(stream string) error {
	err := q.client.XGroupCreateMkStream(q.ctx, stream, q.group, "0").Err()
	if err == nil || strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}

	return err
}

func (q *Queue) later(delay time.Time, task contractsqueue.Task, queue string) error {
	task.Delay = time.Time{}
	payload, err := utils.TaskToJson(task, q.json)
	if err != nil {
		return err
	}

	return q.client.ZAdd(q.ctx, q.queueKey.delayed(queue), redis.Z{
		Score:  float64(delay.Unix()),
		Member: payload,
	}).Err()
}

func (q *Queue) migrateDelayedJobs(queue string) error {
	delayedKey := q.queueKey.delayed(queue)
	streamKey := q.queueKey.stream(queue)

	jobs, err := q.client.ZRangeByScore(q.ctx, delayedKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(time.Now().Unix(), 10),
	}).Result()
	if err != nil || len(jobs) == 0 {
		return err
	}

	pipe := q.client.TxPipeline()
	for _, payload := range jobs {
		addArgs := &redis.XAddArgs{
			Stream: streamKey,
			Values: map[string]any{
				"payload": payload,
			},
		}
		if q.streamMax > 0 {
			addArgs.MaxLen = q.streamMax
			addArgs.Approx = true
		}
		pipe.XAdd(q.ctx, addArgs)
		pipe.ZRem(q.ctx, delayedKey, payload)
	}

	_, err = pipe.Exec(q.ctx)
	return err
}

func (q *Queue) messageToReservedJob(stream string, message redis.XMessage) (contractsqueue.ReservedJob, error) {
	rawPayload, exists := message.Values["payload"]
	if !exists {
		return nil, fmt.Errorf("stream message %s has no payload field", message.ID)
	}

	var payload string
	switch val := rawPayload.(type) {
	case string:
		payload = val
	case []byte:
		payload = string(val)
	default:
		payload = fmt.Sprint(val)
	}

	task, err := utils.JsonToTask(payload, q.jobStorer, q.json)
	if err != nil {
		return nil, err
	}

	return newReservedJob(q.ctx, q.client, stream, q.group, message.ID, task, q.deleteOnAck), nil
}
