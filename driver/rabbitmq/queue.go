package rabbitmq

import (
	"context"
	"fmt"
	"time"

	frameworkerrors "github.com/goravel/framework/errors"

	"github.com/goravel/framework/contracts/config"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue/utils"
	"github.com/rabbitmq/amqp091-go"
)

var _ contractsqueue.Driver = &Queue{}

type Queue struct {
	ctx          context.Context
	channel      *amqp091.Channel
	jobStorer    contractsqueue.JobStorer
	json         contractsfoundation.Json
	connection   string
	exchange     string
	exchangeType string
	queuePrefix  string
	routingKey   string
}

func New(connection string) (contractsqueue.Driver, error) {
	app := facades.App()

	return NewQueue(context.Background(), app.MakeConfig(), app.MakeQueue(), app.GetJson(), connection)
}

func NewQueue(ctx context.Context, cfg config.Config, queue contractsqueue.Queue, json contractsfoundation.Json, connection string) (*Queue, error) {
	url := cfg.GetString(fmt.Sprintf("queue.connections.%s.url", connection))
	if url == "" {
		url = cfg.GetString("queue.rabbitmq.url", "amqp://guest:guest@127.0.0.1:5672/")
	}
	if url == "" {
		return nil, fmt.Errorf("rabbitmq url is not configured for queue connection [%s]", connection)
	}

	client, err := GetClient(url)
	if err != nil {
		return nil, err
	}

	exchange := cfg.GetString(fmt.Sprintf("queue.connections.%s.exchange", connection), "")
	if exchange == "" {
		exchange = cfg.GetString("queue.rabbitmq.exchange", "goravel.exchange")
	}

	exchangeType := cfg.GetString(fmt.Sprintf("queue.connections.%s.exchange_type", connection), "")
	if exchangeType == "" {
		exchangeType = cfg.GetString("queue.rabbitmq.exchange_type", "direct")
	}

	queuePrefix := cfg.GetString(fmt.Sprintf("queue.connections.%s.queue_prefix", connection), "")
	if queuePrefix == "" {
		queuePrefix = cfg.GetString("queue.rabbitmq.queue_prefix", cfg.GetString("app.name", "goravel"))
	}

	routingKey := cfg.GetString(fmt.Sprintf("queue.connections.%s.routing_key", connection), "")
	if routingKey == "" {
		routingKey = cfg.GetString("queue.rabbitmq.routing_key", "")
	}

	return &Queue{
		ctx:          ctx,
		channel:      client.channel,
		jobStorer:    queue.JobStorer(),
		json:         json,
		connection:   connection,
		exchange:     exchange,
		exchangeType: exchangeType,
		queuePrefix:  queuePrefix,
		routingKey:   routingKey,
	}, nil
}

func (q *Queue) Driver() string {
	return contractsqueue.DriverCustom
}

func (q *Queue) Push(task contractsqueue.Task, queue string) error {
	queueName := q.queueName(queue)
	routingKey := q.resolveRoutingKey(queue)

	if err := q.ensureTopology(queueName, routingKey); err != nil {
		return err
	}

	payload, err := utils.TaskToJson(task, q.json)
	if err != nil {
		return err
	}

	headers := amqp091.Table{}
	if !task.Delay.IsZero() {
		delayMS := time.Until(task.Delay).Milliseconds()
		if delayMS > 0 {
			// Requires RabbitMQ delayed-message-exchange plugin when exchange_type is x-delayed-message.
			headers["x-delay"] = int32(delayMS)
		}
	}

	return q.channel.PublishWithContext(q.ctx, q.exchange, routingKey, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		Body:         []byte(payload),
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now(),
		Headers:      headers,
	})
}

func (q *Queue) Pop(queue string) (contractsqueue.ReservedJob, error) {
	queueName := q.queueName(queue)
	routingKey := q.resolveRoutingKey(queue)

	if err := q.ensureTopology(queueName, routingKey); err != nil {
		return nil, err
	}

	delivery, ok, err := q.channel.Get(queueName, false)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, frameworkerrors.QueueDriverNoJobFound.Args(queueName)
	}

	task, err := utils.JsonToTask(string(delivery.Body), q.jobStorer, q.json)
	if err != nil {
		_ = delivery.Nack(false, false)
		return nil, err
	}

	return newReservedJob(delivery, task), nil
}

func (q *Queue) ensureTopology(queueName, routingKey string) error {
	if err := q.channel.ExchangeDeclare(
		q.exchange,
		q.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := q.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	return q.channel.QueueBind(
		queueName,
		routingKey,
		q.exchange,
		false,
		nil,
	)
}

func (q *Queue) queueName(queue string) string {
	return q.queuePrefix + "." + q.connection + "." + queue
}

func (q *Queue) resolveRoutingKey(queue string) string {
	if q.routingKey != "" {
		return q.routingKey
	}

	return q.connection + "." + queue
}
