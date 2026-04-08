# Goravel RabbitMQ Driver

A custom queue driver for Goravel based on RabbitMQ (`publish`, `basic.get`, `ack`).

## Install

```bash
go get github.com/wangxuancheng-dev/goravel-rabbitmq
```

## Configuration

In `config/queue.go`:

```go
import (
	"github.com/goravel/framework/contracts/queue"
	rabbitmq "github.com/wangxuancheng-dev/goravel-rabbitmq"
)

"connections": map[string]any{
	"rabbitmq": map[string]any{
		"driver":        "custom",
		"queue":         "default",
		"url":           config.Env("QUEUE_RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
		"exchange":      config.Env("QUEUE_RABBITMQ_EXCHANGE", "goravel.exchange"),
		"exchange_type": config.Env("QUEUE_RABBITMQ_EXCHANGE_TYPE", "direct"),
		"queue_prefix":  config.Env("QUEUE_RABBITMQ_QUEUE_PREFIX", config.GetString("app.name", "goravel")),
		"routing_key":   config.Env("QUEUE_RABBITMQ_ROUTING_KEY", ""),
		"via": func() (queue.Driver, error) {
			return rabbitmq.New("rabbitmq")
		},
	},
},
```

In `.env`:

```env
QUEUE_CONNECTION=rabbitmq
QUEUE_RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/
QUEUE_RABBITMQ_EXCHANGE=goravel.exchange
QUEUE_RABBITMQ_EXCHANGE_TYPE=direct
QUEUE_RABBITMQ_QUEUE_PREFIX=goravel-admin
QUEUE_RABBITMQ_ROUTING_KEY=
```

## Delay Jobs

This driver supports delayed publishing via RabbitMQ `x-delay` header when using
`x-delayed-message` exchange plugin.
