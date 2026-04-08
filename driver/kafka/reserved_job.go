package kafka

import (
	"context"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	kafkago "github.com/segmentio/kafka-go"
)

type reservedJob struct {
	ctx     context.Context
	reader  *kafkago.Reader
	message kafkago.Message
	task    contractsqueue.Task
}

func newReservedJob(ctx context.Context, reader *kafkago.Reader, message kafkago.Message, task contractsqueue.Task) *reservedJob {
	return &reservedJob{
		ctx:     ctx,
		reader:  reader,
		message: message,
		task:    task,
	}
}

func (j *reservedJob) Delete() error {
	return j.reader.CommitMessages(j.ctx, j.message)
}

func (j *reservedJob) Task() contractsqueue.Task {
	return j.task
}
