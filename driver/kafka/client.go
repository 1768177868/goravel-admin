package kafka

import (
	"fmt"
	"strings"
	"sync"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type client struct {
	writer *kafkago.Writer
}

var (
	clients sync.Map
	mu      sync.Mutex
)

func GetClient(brokers []string, dialTimeout, writeTimeout time.Duration) (*client, error) {
	key := strings.Join(brokers, ",")
	if cached, ok := clients.Load(key); ok {
		return cached.(*client), nil
	}

	mu.Lock()
	defer mu.Unlock()

	if cached, ok := clients.Load(key); ok {
		return cached.(*client), nil
	}

	newClient, err := createClient(brokers, dialTimeout, writeTimeout)
	if err != nil {
		return nil, err
	}
	clients.Store(key, newClient)
	return newClient, nil
}

func createClient(brokers []string, dialTimeout, writeTimeout time.Duration) (*client, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are empty")
	}

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
		Transport: &kafkago.Transport{
			DialTimeout: dialTimeout,
		},
		WriteTimeout: writeTimeout,
	}

	return &client{writer: writer}, nil
}
