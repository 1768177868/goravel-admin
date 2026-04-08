# Goravel Kafka Driver

A custom queue driver for Goravel based on Kafka (`write`, `fetch`, `commit`).

## Install

```bash
go get github.com/wangxuancheng-dev/goravel-kafka
```

## Configuration

In `config/queue.go`:

```go
import (
	"github.com/goravel/framework/contracts/queue"
	kafka "github.com/wangxuancheng-dev/goravel-kafka"
)

"connections": map[string]any{
	"kafka": map[string]any{
		"driver":           "custom",
		"queue":            config.Env("QUEUE_KAFKA_QUEUE", "default"),
		"brokers":          config.Env("QUEUE_KAFKA_BROKERS", "127.0.0.1:9092"),
		"group_id":         config.Env("QUEUE_KAFKA_GROUP_ID", "goravel"),
		"dial_timeout_ms":  config.Env("QUEUE_KAFKA_DIAL_TIMEOUT_MS", 1000),
		"write_timeout_ms": config.Env("QUEUE_KAFKA_WRITE_TIMEOUT_MS", 1000),
		"read_timeout_ms":  config.Env("QUEUE_KAFKA_READ_TIMEOUT_MS", 1000),
		"max_wait_ms":      config.Env("QUEUE_KAFKA_MAX_WAIT_MS", 1000),
		"via": func() (queue.Driver, error) {
			return kafka.New("kafka")
		},
	},
},
```

In `.env`:

```env
QUEUE_CONNECTION=kafka
QUEUE_KAFKA_QUEUE=default
QUEUE_KAFKA_BROKERS=127.0.0.1:9092
QUEUE_KAFKA_GROUP_ID=goravel
QUEUE_KAFKA_DIAL_TIMEOUT_MS=1000
QUEUE_KAFKA_WRITE_TIMEOUT_MS=1000
QUEUE_KAFKA_READ_TIMEOUT_MS=1000
QUEUE_KAFKA_MAX_WAIT_MS=1000
```

## Notes

- `Delete()` maps to Kafka consumer group commit.
- Delay jobs are implemented via in-process wait before write.
