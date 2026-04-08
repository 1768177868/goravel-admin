package redisstream

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/redis/go-redis/v9"
)

var (
	clients sync.Map
	mu      sync.Mutex
)

func GetClient(config config.Config, connection string) (redis.UniversalClient, error) {
	if client, ok := clients.Load(connection); ok {
		return client.(redis.UniversalClient), nil
	}

	mu.Lock()
	defer mu.Unlock()

	if client, ok := clients.Load(connection); ok {
		return client.(redis.UniversalClient), nil
	}

	newClient, err := createClient(config, connection)
	if err != nil {
		return nil, err
	}
	clients.Store(connection, newClient)

	return newClient, nil
}

func createClient(config config.Config, connection string) (redis.UniversalClient, error) {
	configPrefix := fmt.Sprintf("database.redis.%s", connection)
	host := config.GetString(fmt.Sprintf("%s.host", configPrefix))
	if host == "" {
		return nil, fmt.Errorf("redis host is not configured for connection [%s] at path '%s.host'", connection, configPrefix)
	}

	port := config.GetString(fmt.Sprintf("%s.port", configPrefix), "6379")
	username := config.GetString(fmt.Sprintf("%s.username", configPrefix))
	password := config.GetString(fmt.Sprintf("%s.password", configPrefix))
	db := config.GetInt(fmt.Sprintf("%s.database", configPrefix), 0)
	cluster := config.GetBool(fmt.Sprintf("%s.cluster", configPrefix), false)

	options := &redis.UniversalOptions{
		Addrs:           []string{fmt.Sprintf("%s:%s", host, port)},
		Username:        username,
		Password:        password,
		DB:              db,
		IsClusterMode:   cluster,
		DisableIdentity: true,
	}

	tlsConfigRaw := config.Get(fmt.Sprintf("%s.tls", configPrefix))
	if tlsConfig, ok := tlsConfigRaw.(*tls.Config); ok && tlsConfig != nil {
		options.TLSConfig = tlsConfig
	}

	client := redis.NewUniversalClient(options)
	pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect redis [%s]: %w", connection, err)
	}

	return client, nil
}
