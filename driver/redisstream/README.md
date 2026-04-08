# Goravel Redis Stream Driver

A custom queue driver for Goravel based on Redis Streams (`XADD`, `XREADGROUP`, `XAUTOCLAIM`, `XACK`).

## Install

```bash
go get github.com/wangxuancheng-dev/goravel-redis-stream
```

## Update
```bash
go get github.com/wangxuancheng-dev/goravel-redis-stream@latest
go mod tidy
```

## Configuration

In `config/queue.go`:

```go
import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/contracts/queue"
	redisstream "github.com/wangxuancheng-dev/goravel-redis-stream"
)

// ...
"connections": map[string]any{
	"redis_stream": map[string]any{
		"driver":         "custom",
		"connection":     facades.Config().Env("QUEUE_REDIS_STREAM_CONNECTION", "default"), // database.redis.{connection}
		"queue":          facades.Config().Env("QUEUE_REDIS_STREAM_QUEUE", "default"),
		"group":          facades.Config().Env("QUEUE_REDIS_STREAM_GROUP", "goravel"),
		"consumer":       facades.Config().Env("QUEUE_REDIS_STREAM_CONSUMER", ""), // default: {hostname}-{pid}
		"block_ms":       facades.Config().Env("QUEUE_REDIS_STREAM_BLOCK_MS", 1000),
		"retry_after":    facades.Config().Env("QUEUE_REDIS_STREAM_RETRY_AFTER", 90),
		"claim_count":    facades.Config().Env("QUEUE_REDIS_STREAM_CLAIM_COUNT", 10),
		"delete_on_ack":  facades.Config().Env("QUEUE_REDIS_STREAM_DELETE_ON_ACK", false), // true => XACK + XDEL
		"stream_max_len": facades.Config().Env("QUEUE_REDIS_STREAM_MAX_LEN", 100000),       // 0 disables trimming, >0 uses approximate MAXLEN
		"via": func() (queue.Driver, error) {
			return redisstream.New("redis_stream")
		},
	},
},
```

In `.env`:

```env
QUEUE_CONNECTION=redis_stream
QUEUE_REDIS_STREAM_CONNECTION=default
QUEUE_REDIS_STREAM_QUEUE=default
QUEUE_REDIS_STREAM_GROUP=goravel
QUEUE_REDIS_STREAM_CONSUMER=
QUEUE_REDIS_STREAM_BLOCK_MS=1000
QUEUE_REDIS_STREAM_RETRY_AFTER=90
QUEUE_REDIS_STREAM_CLAIM_COUNT=10
QUEUE_REDIS_STREAM_DELETE_ON_ACK=false
QUEUE_REDIS_STREAM_MAX_LEN=100000
```

## Behavior

- Immediate jobs: `XADD` to stream.
- Delayed jobs: stored in `ZSET`, migrated to stream when due.
- Consume: `XREADGROUP`.
- Recovery: `XAUTOCLAIM` reclaims stale pending messages after `retry_after`.
- ACK: on `ReservedJob.Delete()` with optional `XDEL`.

## Notes

- When `delete_on_ack=false`, ACKed entries remain in stream for audit/replay.
- Set `stream_max_len` to avoid unbounded stream growth.
- Set `retry_after` greater than typical job runtime to avoid early reclaim.

## Testing

Run unit tests in this module:

```bash
go test ./...
```

Current test coverage includes:

- `Push + Pop` basic flow.
- `Delete()` ACK behavior with `delete_on_ack=true` and `delete_on_ack=false`.
- delayed job migration from `ZSET` to stream.
- `XAUTOCLAIM` reclaim of stale pending messages.
- edge cases:
  - future delayed jobs are not consumed early;
  - invalid stream message payload returns explicit error.

## Integration Verification

In your application:

1. Set `QUEUE_CONNECTION=redis_stream`.
2. Start the app (`go run .`).
3. Confirm worker logs show:
   - `Processing jobs from [redis_stream] connection and [default] queue`
   - `Processing jobs from [redis_stream] connection and [long-running] queue` (if enabled)
4. Dispatch and verify:
   - one immediate job;
   - one delayed job;
   - one job on a custom queue (e.g. `long-running`).


```bash
# 1) 看状态
git status

# 2) 提交代码
git add .
git commit -m "feat: add redis stream queue driver with tests"

# 3) 打语义化版本标签（示例 v0.1.0）
git tag -a v0.1.0 -m "v0.1.0"

# 4) 推送分支
git push origin main

# 5) 推送标签
git push origin v0.1.0
# 或一次推所有标签：git push origin --tags

git tag                 # 查看本地标签
git show v0.1.0         # 看标签详情

# 删除本地标签
git tag -d v0.1.0

# 删除远程标签
git push origin :refs/tags/v0.1.0

# 更新标签版本
go get github.com/wangxuancheng-dev/goravel-redis-stream@v0.1.0
go mod tidy
```