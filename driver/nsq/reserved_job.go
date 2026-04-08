package nsq

import (
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/nsqio/go-nsq"
)

type reservedJob struct {
	message *nsq.Message
	task    contractsqueue.Task
}

func newReservedJob(message *nsq.Message, task contractsqueue.Task) *reservedJob {
	return &reservedJob{
		message: message,
		task:    task,
	}
}

func (j *reservedJob) Delete() error {
	j.message.Finish()
	return nil
}

func (j *reservedJob) Task() contractsqueue.Task {
	return j.task
}
