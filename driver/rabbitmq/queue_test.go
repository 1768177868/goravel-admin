package rabbitmq

import (
	"context"
	"fmt"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	frameworkerrors "github.com/goravel/framework/errors"
	frameworkjson "github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/rabbitmq/amqp091-go"
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
	queue, cleanup := newTestQueue(t, "rabbitmq_test_push_pop")
	defer cleanup()

	task := contractsqueue.Task{
		UUID: "uuid-rabbitmq-push-pop",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "hello-rabbitmq"},
			},
		},
	}

	targetQueue := uniqueQueue("q-push-pop")
	assert.NoError(t, queue.Push(task, targetQueue))

	reserved, err := queue.Pop(targetQueue)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.Equal(t, "mock", reserved.Task().Job.Signature())
	assert.Equal(t, "hello-rabbitmq", reserved.Task().Args[0].Value)
}

func TestQueueDeleteAck(t *testing.T) {
	queue, cleanup := newTestQueue(t, "rabbitmq_test_ack")
	defer cleanup()

	task := contractsqueue.Task{
		UUID: "uuid-rabbitmq-ack",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
		},
	}

	targetQueue := uniqueQueue("q-ack")
	assert.NoError(t, queue.Push(task, targetQueue))

	reserved, err := queue.Pop(targetQueue)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.NoError(t, reserved.Delete())

	reservedAfter, err := queue.Pop(targetQueue)
	assert.Nil(t, reservedAfter)
	assert.Error(t, err)
	assert.True(t, frameworkerrors.Is(err, frameworkerrors.QueueDriverNoJobFound))
}

func TestQueuePopNoJob(t *testing.T) {
	queue, cleanup := newTestQueue(t, "rabbitmq_test_no_job")
	defer cleanup()

	targetQueue := uniqueQueue("q-no-job")
	reserved, err := queue.Pop(targetQueue)
	assert.Nil(t, reserved)
	assert.Error(t, err)
	assert.True(t, frameworkerrors.Is(err, frameworkerrors.QueueDriverNoJobFound))
}

func TestQueuePopInvalidPayload(t *testing.T) {
	queue, cleanup := newTestQueue(t, "rabbitmq_test_invalid_payload")
	defer cleanup()

	targetQueue := uniqueQueue("q-invalid-payload")
	queueName := queue.queueName(targetQueue)
	routingKey := queue.resolveRoutingKey(targetQueue)

	assert.NoError(t, queue.ensureTopology(queueName, routingKey))
	assert.NoError(t, queue.channel.PublishWithContext(context.Background(), queue.exchange, routingKey, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        []byte("not-json-payload"),
	}))

	reserved, err := queue.Pop(targetQueue)
	assert.Nil(t, reserved)
	assert.Error(t, err)
}

func newTestQueue(t *testing.T, connection string) (*Queue, func()) {
	t.Helper()

	clearClientCache()
	url := "amqp://guest:guest@127.0.0.1:5672/"
	if _, err := amqp091.Dial(url); err != nil {
		t.Skipf("skip rabbitmq tests, cannot connect to %s: %v", url, err)
	}

	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockJobStorer := mocksqueue.NewJobStorer(t)

	mockQueue.On("JobStorer").Return(mockJobStorer).Maybe()
	mockJobStorer.On("Get", "mock").Return(&mockJob{}, nil).Maybe()

	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.url", connection)).Return(url).Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.exchange", connection), "").Return("goravel.test.exchange").Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.exchange_type", connection), "").Return("direct").Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.queue_prefix", connection), "").Return("goravel-test").Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.routing_key", connection), "").Return("").Maybe()
	mockConfig.On("GetString", "app.name", "goravel").Return("goravel").Maybe()

	queue, err := NewQueue(context.Background(), mockConfig, mockQueue, frameworkjson.New(), connection)
	assert.NoError(t, err)

	cleanup := func() {
		clearClientCache()
	}

	return queue, cleanup
}

func clearClientCache() {
	clients.Range(func(key, value any) bool {
		if client, ok := value.(*brokerClient); ok {
			if client.channel != nil {
				_ = client.channel.Close()
			}
			if client.conn != nil {
				_ = client.conn.Close()
			}
		}
		clients.Delete(key)
		return true
	})
}

func uniqueQueue(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
