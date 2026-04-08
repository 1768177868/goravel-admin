package nsq

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	frameworkerrors "github.com/goravel/framework/errors"

	"github.com/goravel/framework/contracts/config"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue/utils"
	"github.com/nsqio/go-nsq"
)

var _ contractsqueue.Driver = &Queue{}

type Queue struct {
	ctx          context.Context
	producer     *nsq.Producer
	nsqdAddr     string
	jobStorer    contractsqueue.JobStorer
	json         contractsfoundation.Json
	connection   string
	channel      string
	popTimeout   time.Duration
	maxInFlight  int
	consumerLock sync.Mutex
	consumers    map[string]*consumerState
}

type consumerState struct {
	consumer   *nsq.Consumer
	deliveries chan *nsq.Message
}

type deliveryHandler struct {
	deliveries chan *nsq.Message
}

func (h *deliveryHandler) HandleMessage(message *nsq.Message) error {
	h.deliveries <- message
	return nil
}

func New(connection string) (contractsqueue.Driver, error) {
	app := facades.App()
	return NewQueue(context.Background(), app.MakeConfig(), app.MakeQueue(), app.GetJson(), connection)
}

func NewQueue(ctx context.Context, cfg config.Config, queue contractsqueue.Queue, json contractsfoundation.Json, connection string) (*Queue, error) {
	nsqdAddr := cfg.GetString(fmt.Sprintf("queue.connections.%s.nsqd_tcp_address", connection))
	if nsqdAddr == "" {
		nsqdAddr = cfg.GetString("queue.nsq.nsqd_tcp_address", "127.0.0.1:4150")
	}
	if strings.TrimSpace(nsqdAddr) == "" {
		return nil, fmt.Errorf("nsqd tcp address is not configured for queue connection [%s]", connection)
	}

	timeoutMS := cfg.GetInt(fmt.Sprintf("queue.connections.%s.timeout_ms", connection), 1000)
	if timeoutMS <= 0 {
		timeoutMS = 1000
	}
	timeout := time.Duration(timeoutMS) * time.Millisecond

	client, err := GetClient(nsqdAddr, timeout)
	if err != nil {
		return nil, err
	}

	channel := cfg.GetString(fmt.Sprintf("queue.connections.%s.channel", connection), "")
	if channel == "" {
		channel = cfg.GetString("queue.nsq.channel", "goravel")
	}

	maxInFlight := cfg.GetInt(fmt.Sprintf("queue.connections.%s.max_in_flight", connection), 1)
	if maxInFlight <= 0 {
		maxInFlight = 1
	}

	return &Queue{
		ctx:         ctx,
		producer:    client.producer,
		nsqdAddr:    client.nsqdAddr,
		jobStorer:   queue.JobStorer(),
		json:        json,
		connection:  connection,
		channel:     channel,
		popTimeout:  timeout,
		maxInFlight: maxInFlight,
		consumers:   make(map[string]*consumerState),
	}, nil
}

func (q *Queue) Driver() string {
	return contractsqueue.DriverCustom
}

func (q *Queue) Push(task contractsqueue.Task, queue string) error {
	payload, err := utils.TaskToJson(task, q.json)
	if err != nil {
		return err
	}

	if !task.Delay.IsZero() {
		delay := time.Until(task.Delay)
		if delay > 0 {
			return q.producer.DeferredPublish(queue, delay, []byte(payload))
		}
	}

	return q.producer.Publish(queue, []byte(payload))
}

func (q *Queue) Pop(queue string) (contractsqueue.ReservedJob, error) {
	state, err := q.getOrCreateConsumer(queue)
	if err != nil {
		return nil, err
	}

	select {
	case msg := <-state.deliveries:
		task, decodeErr := utils.JsonToTask(string(msg.Body), q.jobStorer, q.json)
		if decodeErr != nil {
			msg.RequeueWithoutBackoff(0)
			return nil, decodeErr
		}
		return newReservedJob(msg, task), nil
	case <-time.After(q.popTimeout):
		return nil, frameworkerrors.QueueDriverNoJobFound.Args(queue)
	case <-q.ctx.Done():
		return nil, q.ctx.Err()
	}
}

func (q *Queue) getOrCreateConsumer(topic string) (*consumerState, error) {
	q.consumerLock.Lock()
	defer q.consumerLock.Unlock()

	if state, ok := q.consumers[topic]; ok {
		return state, nil
	}

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = q.maxInFlight

	consumer, err := nsq.NewConsumer(topic, q.channel+"_"+sanitizeTopic(topic), cfg)
	if err != nil {
		return nil, err
	}

	state := &consumerState{
		consumer:   consumer,
		deliveries: make(chan *nsq.Message, q.maxInFlight*4),
	}
	consumer.AddHandler(&deliveryHandler{deliveries: state.deliveries})

	if err := consumer.ConnectToNSQD(q.nsqdAddr); err != nil {
		consumer.Stop()
		return nil, err
	}

	q.consumers[topic] = state
	return state, nil
}

func sanitizeTopic(topic string) string {
	topic = strings.ReplaceAll(topic, ".", "_")
	topic = strings.ReplaceAll(topic, "-", "_")
	if topic == "" {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return topic
}
