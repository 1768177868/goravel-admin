package facades

import (
	"github.com/goravel/framework/contracts/queue"

	redisstream "github.com/wangxuancheng-dev/goravel-redis-stream"
)

func Queue(connection string) (queue.Driver, error) {
	if redisstream.App == nil {
		return nil, redisstream.ErrServiceProviderNotRegistered
	}
	if connection == "" {
		return nil, redisstream.ErrQueueConnectionIsRequired
	}

	instance, err := redisstream.App.MakeWith(redisstream.BindingQueue, map[string]any{
		"connection": connection,
	})
	if err != nil {
		return nil, err
	}

	return instance.(*redisstream.Queue), nil
}
