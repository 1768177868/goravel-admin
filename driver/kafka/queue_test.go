package kafka

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
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockJob struct{}

func (m *mockJob) Signature() string {
	return "mock"
}

func (m *mockJob) Handle(args ...any) error {
	return nil
}

func TestQueuePushPop(t *testing.T) {
	queue, cleanup := newTestQueue(t, "kafka_test_push_pop")
	defer cleanup()

	topic := uniqueTopic("kafka-push-pop")
	task := contractsqueue.Task{
		UUID: "uuid-kafka-push-pop",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "hello-kafka"},
			},
		},
	}

	assert.NoError(t, queue.Push(task, topic))

	reserved, err := queue.Pop(topic)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.Equal(t, "mock", reserved.Task().Job.Signature())
	assert.Equal(t, "hello-kafka", reserved.Task().Args[0].Value)
}

func TestQueueDeleteAck(t *testing.T) {
	queue, cleanup := newTestQueue(t, "kafka_test_ack")
	defer cleanup()

	topic := uniqueTopic("kafka-ack")
	task := contractsqueue.Task{
		UUID: "uuid-kafka-ack",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
		},
	}

	assert.NoError(t, queue.Push(task, topic))

	reserved, err := queue.Pop(topic)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.NoError(t, reserved.Delete())
}

func TestQueuePopNoJob(t *testing.T) {
	queue, cleanup := newTestQueue(t, "kafka_test_no_job")
	defer cleanup()

	reserved, err := queue.Pop(uniqueTopic("kafka-no-job"))
	assert.Nil(t, reserved)
	assert.Error(t, err)
	assert.True(t, frameworkerrors.Is(err, frameworkerrors.QueueDriverNoJobFound))
}

func newTestQueue(t *testing.T, connection string) (*Queue, func()) {
	t.Helper()

	clearClientCache()
	broker := "127.0.0.1:9092"
	conn, err := kafkago.DialContext(context.Background(), "tcp", broker)
	if err != nil {
		t.Skipf("skip kafka tests, cannot connect broker %s", broker)
	}
	_ = conn.Close()

	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockJobStorer := mocksqueue.NewJobStorer(t)

	mockQueue.On("JobStorer").Return(mockJobStorer).Maybe()
	mockJobStorer.On("Get", "mock").Return(&mockJob{}, nil).Maybe()

	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.brokers", connection)).Return(broker).Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.group_id", connection), "").Return("goravel-test").Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.dial_timeout_ms", connection), 1000).Return(1000).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.write_timeout_ms", connection), 1000).Return(1000).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.read_timeout_ms", connection), 1000).Return(1000).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.max_wait_ms", connection), 1000).Return(200).Maybe()

	queue, err := NewQueue(context.Background(), mockConfig, mockQueue, frameworkjson.New(), connection)
	require.NoError(t, err)

	return queue, func() {
		clearClientCache()
	}
}

func clearClientCache() {
	clients.Range(func(key, value any) bool {
		if client, ok := value.(*client); ok && client.writer != nil {
			_ = client.writer.Close()
		}
		clients.Delete(key)
		return true
	})
}

func uniqueTopic(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
