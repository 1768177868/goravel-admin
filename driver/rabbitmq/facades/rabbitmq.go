package facades

import (
	"github.com/goravel/framework/contracts/queue"

	rabbitmq "github.com/wangxuancheng-dev/goravel-rabbitmq"
)

func Queue(connection string) (queue.Driver, error) {
	if rabbitmq.App == nil {
		return nil, rabbitmq.ErrServiceProviderNotRegistered
	}
	if connection == "" {
		return nil, rabbitmq.ErrQueueConnectionIsRequired
	}

	instance, err := rabbitmq.App.MakeWith(rabbitmq.BindingQueue, map[string]any{
		"connection": connection,
	})
	if err != nil {
		return nil, err
	}

	return instance.(*rabbitmq.Queue), nil
}
