# Goravel NSQ Driver

A custom queue driver for Goravel based on NSQ (`publish`, `deferred publish`, `consume`, `finish`).

## Install

```bash
go get github.com/wangxuancheng-dev/goravel-nsq
```

## Configuration

In `config/queue.go`:

```go
import (
	"github.com/goravel/framework/contracts/queue"
	nsq "github.com/wangxuancheng-dev/goravel-nsq"
)

"connections": map[string]any{
	"nsq": map[string]any{
		"driver":           "custom",
		"queue":            config.Env("QUEUE_NSQ_QUEUE", "default"),
		"nsqd_tcp_address": config.Env("QUEUE_NSQ_NSQD_TCP_ADDRESS", "127.0.0.1:4150"),
		"channel":          config.Env("QUEUE_NSQ_CHANNEL", "goravel"),
		"timeout_ms":       config.Env("QUEUE_NSQ_TIMEOUT_MS", 1000),
		"max_in_flight":    config.Env("QUEUE_NSQ_MAX_IN_FLIGHT", 1),
		"via": func() (queue.Driver, error) {
			return nsq.New("nsq")
		},
	},
},
```

In `.env`:

```env
QUEUE_CONNECTION=nsq
QUEUE_NSQ_QUEUE=default
QUEUE_NSQ_NSQD_TCP_ADDRESS=127.0.0.1:4150
QUEUE_NSQ_CHANNEL=goravel
QUEUE_NSQ_TIMEOUT_MS=1000
QUEUE_NSQ_MAX_IN_FLIGHT=1
```

## Notes

- Delayed jobs use NSQ `DeferredPublish`.
- `Delete()` maps to NSQ `Finish()`.
