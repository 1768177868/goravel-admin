package redisstream

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	frameworkerrors "github.com/goravel/framework/errors"
	frameworkjson "github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type mockJob struct{}

func (m *mockJob) Signature() string {
	return "mock"
}

func (m *mockJob) Handle(args ...any) error {
	return nil
}

func TestQueuePushPop(t *testing.T) {
	queue, _ := newTestQueue(t, "c1", map[string]any{
		"retry_after": 1,
	})

	task := contractsqueue.Task{
		UUID: "uuid-push-pop",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "hello"},
			},
		},
	}

	assert.NoError(t, queue.Push(task, "default"))

	reserved, err := queue.Pop("default")
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.Equal(t, "mock", reserved.Task().Job.Signature())
	assert.Equal(t, "hello", reserved.Task().Args[0].Value)
}

func TestQueueDeleteAck(t *testing.T) {
	queue, _ := newTestQueue(t, "c1", map[string]any{
		"retry_after":   1,
		"delete_on_ack": true,
	})

	task := contractsqueue.Task{
		UUID: "uuid-ack",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
		},
	}

	assert.NoError(t, queue.Push(task, "ack"))
	reserved, err := queue.Pop("ack")
	assert.NoError(t, err)

	stream := queue.queueKey.stream("ack")
	pendingBefore, err := queue.client.XPending(queue.ctx, stream, queue.group).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), pendingBefore.Count)

	assert.NoError(t, reserved.Delete())

	pendingAfter, err := queue.client.XPending(queue.ctx, stream, queue.group).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pendingAfter.Count)

	length, err := queue.client.XLen(queue.ctx, stream).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), length)
}

func TestQueueDeleteAckWithoutXDel(t *testing.T) {
	queue, _ := newTestQueue(t, "c1", map[string]any{
		"retry_after":   1,
		"delete_on_ack": false,
	})

	task := contractsqueue.Task{
		UUID: "uuid-ack-keep-stream",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
		},
	}

	assert.NoError(t, queue.Push(task, "ack-keep"))
	reserved, err := queue.Pop("ack-keep")
	assert.NoError(t, err)
	assert.NoError(t, reserved.Delete())

	stream := queue.queueKey.stream("ack-keep")
	pendingAfter, err := queue.client.XPending(queue.ctx, stream, queue.group).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pendingAfter.Count)

	length, err := queue.client.XLen(queue.ctx, stream).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), length)
}

func TestQueueDelayMigration(t *testing.T) {
	queue, _ := newTestQueue(t, "c1", map[string]any{
		"retry_after": 1,
	})

	task := contractsqueue.Task{
		UUID: "uuid-delay",
		ChainJob: contractsqueue.ChainJob{
			Job:   &mockJob{},
			Delay: time.Now().Add(2 * time.Second),
		},
	}

	assert.NoError(t, queue.Push(task, "delay"))

	delayedCount, err := queue.client.ZCard(queue.ctx, queue.queueKey.delayed("delay")).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), delayedCount)

	time.Sleep(2200 * time.Millisecond)

	reserved, err := queue.Pop("delay")
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.Equal(t, "mock", reserved.Task().Job.Signature())
}

func TestQueueDelayFutureNotConsumed(t *testing.T) {
	queue, _ := newTestQueue(t, "c1", map[string]any{
		"retry_after": 1,
		"block_ms":    10,
	})

	task := contractsqueue.Task{
		UUID: "uuid-future-delay",
		ChainJob: contractsqueue.ChainJob{
			Job:   &mockJob{},
			Delay: time.Now().Add(30 * time.Second),
		},
	}

	assert.NoError(t, queue.Push(task, "future-delay"))

	reserved, err := queue.Pop("future-delay")
	assert.Nil(t, reserved)
	assert.Error(t, err)
	assert.True(t, frameworkerrors.Is(err, frameworkerrors.QueueDriverNoJobFound))

	delayedCount, zErr := queue.client.ZCard(queue.ctx, queue.queueKey.delayed("future-delay")).Result()
	assert.NoError(t, zErr)
	assert.Equal(t, int64(1), delayedCount)
}

func TestQueueAutoClaim(t *testing.T) {
	queueConsumer1, mr := newTestQueue(t, "c1", map[string]any{
		"retry_after": 1,
	})
	queueConsumer2, _ := newTestQueueWithRedis(t, mr, "c2", map[string]any{
		"retry_after": 1,
	})

	task := contractsqueue.Task{
		UUID: "uuid-claim",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
		},
	}

	assert.NoError(t, queueConsumer1.Push(task, "claim"))
	firstReserved, err := queueConsumer1.Pop("claim")
	assert.NoError(t, err)
	assert.NotNil(t, firstReserved)

	// Keep the first message pending by not calling Delete.
	time.Sleep(1100 * time.Millisecond)

	claimedReserved, err := queueConsumer2.Pop("claim")
	assert.NoError(t, err)
	assert.NotNil(t, claimedReserved)
	assert.Equal(t, "mock", claimedReserved.Task().Job.Signature())
}

func TestQueuePopInvalidPayload(t *testing.T) {
	queue, _ := newTestQueue(t, "c1", map[string]any{
		"retry_after": 1,
		"block_ms":    10,
	})

	stream := queue.queueKey.stream("invalid-payload")
	assert.NoError(t, queue.ensureGroup(stream))
	_, addErr := queue.client.XAdd(queue.ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"foo": "bar",
		},
	}).Result()
	assert.NoError(t, addErr)

	reserved, err := queue.Pop("invalid-payload")
	assert.Nil(t, reserved)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has no payload field")
}

func newTestQueue(t *testing.T, consumer string, opts map[string]any) (*Queue, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	assert.NoError(t, err)
	t.Cleanup(mr.Close)

	queue, _ := newTestQueueWithRedis(t, mr, consumer, opts)
	return queue, mr
}

func newTestQueueWithRedis(t *testing.T, mr *miniredis.Miniredis, consumer string, opts map[string]any) (*Queue, *miniredis.Miniredis) {
	t.Helper()
	clearClientCache()

	hostPort := strings.Split(mr.Addr(), ":")
	host, port := hostPort[0], hostPort[1]

	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockJobStorer := mocksqueue.NewJobStorer(t)

	mockQueue.On("JobStorer").Return(mockJobStorer).Maybe()
	mockJobStorer.On("Get", "mock").Return(&mockJob{}, nil).Maybe()

	connection := "redis_stream_test"
	queueRedisConnection := "default"

	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.connection", connection), "default").Return(queueRedisConnection).Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.group", connection), "goravel").Return("test-group").Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.block_ms", connection), 1000).Return(getOptInt(opts, "block_ms", 50)).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.retry_after", connection), 90).Return(getOptInt(opts, "retry_after", 1)).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.claim_count", connection), 10).Return(getOptInt(opts, "claim_count", 10)).Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.consumer", connection), "").Return(consumer).Maybe()
	mockConfig.On("GetString", "app.name", "goravel").Return("test-app").Maybe()
	mockConfig.On("GetBool", fmt.Sprintf("queue.connections.%s.delete_on_ack", connection), false).Return(getOptBool(opts, "delete_on_ack", false)).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.stream_max_len", connection), 0).Return(getOptInt(opts, "stream_max_len", 0)).Maybe()

	mockConfig.On("GetString", "database.redis.default.host").Return(host).Maybe()
	mockConfig.On("GetString", "database.redis.default.port", "6379").Return(port).Maybe()
	mockConfig.On("GetString", "database.redis.default.username").Return("").Maybe()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Maybe()
	mockConfig.On("GetInt", "database.redis.default.database", 0).Return(0).Maybe()
	mockConfig.On("GetBool", "database.redis.default.cluster", false).Return(false).Maybe()
	mockConfig.On("Get", "database.redis.default.tls").Return(nil).Maybe()

	queue, err := NewQueue(context.Background(), mockConfig, mockQueue, frameworkjson.New(), connection)
	assert.NoError(t, err)

	return queue, mr
}

func clearClientCache() {
	clients.Range(func(key, _ any) bool {
		clients.Delete(key)
		return true
	})
}

func getOptInt(opts map[string]any, key string, fallback int) int {
	if val, ok := opts[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return fallback
}

func getOptBool(opts map[string]any, key string, fallback bool) bool {
	if val, ok := opts[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return fallback
}
