package redisstream

import (
	"context"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/redis/go-redis/v9"
)

type reservedJob struct {
	ctx         context.Context
	client      redis.UniversalClient
	stream      string
	group       string
	messageID   string
	task        contractsqueue.Task
	deleteOnAck bool
}

func newReservedJob(ctx context.Context, client redis.UniversalClient, stream, group, messageID string, task contractsqueue.Task, deleteOnAck bool) *reservedJob {
	return &reservedJob{
		ctx:         ctx,
		client:      client,
		stream:      stream,
		group:       group,
		messageID:   messageID,
		task:        task,
		deleteOnAck: deleteOnAck,
	}
}

func (j *reservedJob) Delete() error {
	pipe := j.client.TxPipeline()
	pipe.XAck(j.ctx, j.stream, j.group, j.messageID)
	if j.deleteOnAck {
		pipe.XDel(j.ctx, j.stream, j.messageID)
	}

	_, err := pipe.Exec(j.ctx)
	return err
}

func (j *reservedJob) Task() contractsqueue.Task {
	return j.task
}
