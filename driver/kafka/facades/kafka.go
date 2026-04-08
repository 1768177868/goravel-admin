package facades

import (
	"github.com/goravel/framework/contracts/queue"

	kafka "github.com/wangxuancheng-dev/goravel-kafka"
)

func Queue(connection string) (queue.Driver, error) {
	if kafka.App == nil {
		return nil, kafka.ErrServiceProviderNotRegistered
	}
	if connection == "" {
		return nil, kafka.ErrQueueConnectionIsRequired
	}

	instance, err := kafka.App.MakeWith(kafka.BindingQueue, map[string]any{
		"connection": connection,
	})
	if err != nil {
		return nil, err
	}

	return instance.(*kafka.Queue), nil
}
