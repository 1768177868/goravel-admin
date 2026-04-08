package nsq

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
	nsqgo "github.com/nsqio/go-nsq"
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
	queue, cleanup := newTestQueue(t, "nsq_test_push_pop")
	defer cleanup()

	targetQueue := uniqueQueue("nsq-push-pop")
	task := contractsqueue.Task{
		UUID: "uuid-nsq-push-pop",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
			Args: []contractsqueue.Arg{
				{Type: "string", Value: "hello-nsq"},
			},
		},
	}

	assert.NoError(t, queue.Push(task, targetQueue))

	reserved, err := queue.Pop(targetQueue)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.Equal(t, "mock", reserved.Task().Job.Signature())
	assert.Equal(t, "hello-nsq", reserved.Task().Args[0].Value)
}

func TestQueueDeleteAck(t *testing.T) {
	queue, cleanup := newTestQueue(t, "nsq_test_ack")
	defer cleanup()

	targetQueue := uniqueQueue("nsq-ack")
	task := contractsqueue.Task{
		UUID: "uuid-nsq-ack",
		ChainJob: contractsqueue.ChainJob{
			Job: &mockJob{},
		},
	}

	assert.NoError(t, queue.Push(task, targetQueue))

	reserved, err := queue.Pop(targetQueue)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.NoError(t, reserved.Delete())
}

func TestQueuePopNoJob(t *testing.T) {
	queue, cleanup := newTestQueue(t, "nsq_test_no_job")
	defer cleanup()

	reserved, err := queue.Pop(uniqueQueue("nsq-no-job"))
	assert.Nil(t, reserved)
	assert.Error(t, err)
	assert.True(t, frameworkerrors.Is(err, frameworkerrors.QueueDriverNoJobFound))
}

func newTestQueue(t *testing.T, connection string) (*Queue, func()) {
	t.Helper()

	clearClientCache()
	nsqdAddr := "127.0.0.1:4150"
	cfg := nsqgo.NewConfig()
	producer, err := nsqgo.NewProducer(nsqdAddr, cfg)
	if err != nil || producer.Ping() != nil {
		if producer != nil {
			producer.Stop()
		}
		t.Skipf("skip nsq tests, cannot connect nsqd %s", nsqdAddr)
	}
	producer.Stop()

	mockConfig := mocksconfig.NewConfig(t)
	mockQueue := mocksqueue.NewQueue(t)
	mockJobStorer := mocksqueue.NewJobStorer(t)

	mockQueue.On("JobStorer").Return(mockJobStorer).Maybe()
	mockJobStorer.On("Get", "mock").Return(&mockJob{}, nil).Maybe()

	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.nsqd_tcp_address", connection)).Return(nsqdAddr).Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.timeout_ms", connection), 1000).Return(500).Maybe()
	mockConfig.On("GetString", fmt.Sprintf("queue.connections.%s.channel", connection), "").Return("goravel-test").Maybe()
	mockConfig.On("GetInt", fmt.Sprintf("queue.connections.%s.max_in_flight", connection), 1).Return(1).Maybe()
	mockConfig.On("GetString", "app.name", "goravel").Return("goravel").Maybe()

	queue, err := NewQueue(context.Background(), mockConfig, mockQueue, frameworkjson.New(), connection)
	require.NoError(t, err)

	cleanup := func() {
		clearClientCache()
	}
	return queue, cleanup
}

func clearClientCache() {
	clients.Range(func(key, value any) bool {
		if client, ok := value.(*brokerClient); ok {
			if client.producer != nil {
				client.producer.Stop()
			}
		}
		clients.Delete(key)
		return true
	})
}

func uniqueQueue(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
