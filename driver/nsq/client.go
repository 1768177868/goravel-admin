package nsq

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

type brokerClient struct {
	producer *nsq.Producer
	nsqdAddr string
}

var (
	clients sync.Map
	mu      sync.Mutex
)

func GetClient(nsqdAddr string, _ time.Duration) (*brokerClient, error) {
	if client, ok := clients.Load(nsqdAddr); ok {
		return client.(*brokerClient), nil
	}

	mu.Lock()
	defer mu.Unlock()

	if client, ok := clients.Load(nsqdAddr); ok {
		return client.(*brokerClient), nil
	}

	newClient, err := createClient(nsqdAddr)
	if err != nil {
		return nil, err
	}
	clients.Store(nsqdAddr, newClient)

	return newClient, nil
}

func createClient(nsqdAddr string) (*brokerClient, error) {
	addr := strings.TrimSpace(nsqdAddr)
	if addr == "" {
		return nil, fmt.Errorf("nsqd tcp address is empty")
	}

	cfg := nsq.NewConfig()

	producer, err := nsq.NewProducer(addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create nsq producer: %w", err)
	}
	if err := producer.Ping(); err != nil {
		producer.Stop()
		return nil, fmt.Errorf("failed to ping nsqd [%s]: %w", addr, err)
	}

	return &brokerClient{
		producer: producer,
		nsqdAddr: addr,
	}, nil
}
