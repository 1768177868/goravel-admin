package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	frameworkerrors "github.com/goravel/framework/errors"

	"github.com/goravel/framework/contracts/config"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue/utils"
	kafkago "github.com/segmentio/kafka-go"
)

var _ contractsqueue.Driver = &Queue{}

type Queue struct {
	ctx          context.Context
	writer       *kafkago.Writer
	jobStorer    contractsqueue.JobStorer
	json         contractsfoundation.Json
	brokers      []string
	connection   string
	groupID      string
	readTimeout  time.Duration
	maxWait      time.Duration
	readerLock   sync.Mutex
	readerByTopic map[string]*kafkago.Reader
}

func New(connection string) (contractsqueue.Driver, error) {
	app := facades.App()
	return NewQueue(context.Background(), app.MakeConfig(), app.MakeQueue(), app.GetJson(), connection)
}

func NewQueue(ctx context.Context, cfg config.Config, queue contractsqueue.Queue, json contractsfoundation.Json, connection string) (*Queue, error) {
	rawBrokers := cfg.GetString(fmt.Sprintf("queue.connections.%s.brokers", connection))
	if rawBrokers == "" {
		rawBrokers = cfg.GetString("queue.kafka.brokers", "127.0.0.1:9092")
	}
	brokers := parseBrokers(rawBrokers)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are not configured for queue connection [%s]", connection)
	}

	groupID := cfg.GetString(fmt.Sprintf("queue.connections.%s.group_id", connection), "")
	if groupID == "" {
		groupID = cfg.GetString("queue.kafka.group_id", "goravel")
	}

	dialTimeout := time.Duration(maxInt(cfg.GetInt(fmt.Sprintf("queue.connections.%s.dial_timeout_ms", connection), 1000), 1)) * time.Millisecond
	writeTimeout := time.Duration(maxInt(cfg.GetInt(fmt.Sprintf("queue.connections.%s.write_timeout_ms", connection), 1000), 1)) * time.Millisecond
	readTimeout := time.Duration(maxInt(cfg.GetInt(fmt.Sprintf("queue.connections.%s.read_timeout_ms", connection), 1000), 1)) * time.Millisecond
	maxWait := time.Duration(maxInt(cfg.GetInt(fmt.Sprintf("queue.connections.%s.max_wait_ms", connection), 1000), 1)) * time.Millisecond

	client, err := GetClient(brokers, dialTimeout, writeTimeout)
	if err != nil {
		return nil, err
	}

	return &Queue{
		ctx:           ctx,
		writer:        client.writer,
		jobStorer:     queue.JobStorer(),
		json:          json,
		brokers:       brokers,
		connection:    connection,
		groupID:       groupID,
		readTimeout:   readTimeout,
		maxWait:       maxWait,
		readerByTopic: make(map[string]*kafkago.Reader),
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

	ctx := q.ctx
	if !task.Delay.IsZero() {
		delay := time.Until(task.Delay)
		if delay > 0 {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return q.writer.WriteMessages(ctx, kafkago.Message{
		Topic: queue,
		Value: []byte(payload),
		Time:  time.Now(),
	})
}

func (q *Queue) Pop(queue string) (contractsqueue.ReservedJob, error) {
	reader := q.getOrCreateReader(queue)

	readCtx, cancel := context.WithTimeout(q.ctx, q.readTimeout)
	defer cancel()

	message, err := reader.FetchMessage(readCtx)
	if err != nil {
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil, frameworkerrors.QueueDriverNoJobFound.Args(queue)
		}
		return nil, err
	}

	task, decodeErr := utils.JsonToTask(string(message.Value), q.jobStorer, q.json)
	if decodeErr != nil {
		_ = reader.CommitMessages(q.ctx, message)
		return nil, decodeErr
	}

	return newReservedJob(q.ctx, reader, message, task), nil
}

func (q *Queue) getOrCreateReader(topic string) *kafkago.Reader {
	q.readerLock.Lock()
	defer q.readerLock.Unlock()

	if reader, ok := q.readerByTopic[topic]; ok {
		return reader
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:   q.brokers,
		GroupID:   q.groupID + "_" + q.connection + "_" + sanitizeTopic(topic),
		Topic:     topic,
		MinBytes:  1,
		MaxBytes:  10e6,
		MaxWait:   q.maxWait,
	})

	q.readerByTopic[topic] = reader
	return reader
}

func parseBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func sanitizeTopic(topic string) string {
	topic = strings.ReplaceAll(topic, ".", "_")
	topic = strings.ReplaceAll(topic, "-", "_")
	if topic == "" {
		return "default"
	}
	return topic
}

func maxInt(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}
