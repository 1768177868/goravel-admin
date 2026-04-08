package rabbitmq

import (
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/rabbitmq/amqp091-go"
)

type reservedJob struct {
	delivery amqp091.Delivery
	task     contractsqueue.Task
}

func newReservedJob(delivery amqp091.Delivery, task contractsqueue.Task) *reservedJob {
	return &reservedJob{
		delivery: delivery,
		task:     task,
	}
}

func (j *reservedJob) Delete() error {
	return j.delivery.Ack(false)
}

func (j *reservedJob) Task() contractsqueue.Task {
	return j.task
}
