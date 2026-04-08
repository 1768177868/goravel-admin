package facades

import (
	"github.com/goravel/framework/contracts/queue"

	nsq "github.com/wangxuancheng-dev/goravel-nsq"
)

func Queue(connection string) (queue.Driver, error) {
	if nsq.App == nil {
		return nil, nsq.ErrServiceProviderNotRegistered
	}
	if connection == "" {
		return nil, nsq.ErrQueueConnectionIsRequired
	}

	instance, err := nsq.App.MakeWith(nsq.BindingQueue, map[string]any{
		"connection": connection,
	})
	if err != nil {
		return nil, err
	}

	return instance.(*nsq.Queue), nil
}
