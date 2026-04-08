package rabbitmq

import (
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

type brokerClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

var (
	clients sync.Map
	mu      sync.Mutex
)

func GetClient(url string) (*brokerClient, error) {
	if client, ok := clients.Load(url); ok {
		return client.(*brokerClient), nil
	}

	mu.Lock()
	defer mu.Unlock()

	if client, ok := clients.Load(url); ok {
		return client.(*brokerClient), nil
	}

	newClient, err := createClient(url)
	if err != nil {
		return nil, err
	}
	clients.Store(url, newClient)

	return newClient, nil
}

func createClient(url string) (*brokerClient, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect rabbitmq: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	return &brokerClient{
		conn:    conn,
		channel: channel,
	}, nil
}
