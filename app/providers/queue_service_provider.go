package providers

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"

	"goravel/app/jobs"
)

type QueueServiceProvider struct {
}

func (receiver *QueueServiceProvider) Register(app foundation.Application) {
	facades.Queue().Register(receiver.Jobs())
}

func (receiver *QueueServiceProvider) Boot(app foundation.Application) {

}

func (receiver *QueueServiceProvider) Jobs() []queue.Job {
	return []queue.Job{
		&jobs.Test{},
		&jobs.TestErr{},
		// 实际场景的Job
		&jobs.SendEmail{},
		&jobs.ProcessImage{},
		&jobs.GenerateReport{},
	}
}
