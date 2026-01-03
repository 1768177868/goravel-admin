package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/redis/go-redis/v9"
)

var (
	// redisClients 缓存不同连接名的 Redis 客户端
	redisClients sync.Map
	// redisMutex 用于保护客户端创建过程
	redisMutex sync.Mutex
)

// GetRedisClient 获取 Redis 客户端（使用连接池，支持多连接）
// connectionName: Redis 连接名称，默认为 "default"
// 返回缓存的 Redis 客户端，如果不存在则创建并缓存
func GetRedisClient(connectionName string) (*redis.Client, error) {
	if connectionName == "" {
		connectionName = "default"
	}

	// 先从缓存中获取
	if client, ok := redisClients.Load(connectionName); ok {
		redisClient := client.(*redis.Client)
		// 测试连接是否有效
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := redisClient.Ping(ctx).Err()
		cancel()
		if err == nil {
			return redisClient, nil
		}
		// 连接失效，从缓存中移除
		redisClients.Delete(connectionName)
	}

	// 使用互斥锁确保只创建一个客户端
	redisMutex.Lock()
	defer redisMutex.Unlock()

	// 双重检查，防止并发创建
	if client, ok := redisClients.Load(connectionName); ok {
		return client.(*redis.Client), nil
	}

	// 创建新的 Redis 客户端
	client, err := createRedisClient(connectionName)
	if err != nil {
		return nil, err
	}

	// 缓存客户端
	redisClients.Store(connectionName, client)

	return client, nil
}

// createRedisClient 创建 Redis 客户端
func createRedisClient(connectionName string) (*redis.Client, error) {
	// 获取 Redis 配置
	host := facades.Config().GetString(fmt.Sprintf("database.redis.%s.host", connectionName), "")
	if host == "" {
		// 尝试使用 default 连接
		host = facades.Config().GetString("database.redis.default.host", "127.0.0.1")
	}

	port := facades.Config().GetInt(fmt.Sprintf("database.redis.%s.port", connectionName), 0)
	if port == 0 {
		port = facades.Config().GetInt("database.redis.default.port", 6379)
	}

	password := facades.Config().GetString(fmt.Sprintf("database.redis.%s.password", connectionName), "")
	if password == "" {
		password = facades.Config().GetString("database.redis.default.password", "")
	}

	db := facades.Config().GetInt(fmt.Sprintf("database.redis.%s.database", connectionName), -1)
	if db == -1 {
		db = facades.Config().GetInt("database.redis.default.database", 0)
	}

	// 创建 Redis 客户端（使用连接池配置）
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		PoolSize:     10, // 连接池大小
		MinIdleConns: 5,  // 最小空闲连接数
		MaxRetries:   3,  // 最大重试次数
	})

	// 测试连接（设置超时）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		client.Close() // 连接失败，关闭客户端
		return nil, fmt.Errorf("Redis 连接失败 [%s]: %v", connectionName, err)
	}

	return client, nil
}

// CloseRedisClient 关闭指定的 Redis 客户端并从缓存中移除
func CloseRedisClient(connectionName string) error {
	if connectionName == "" {
		connectionName = "default"
	}

	if client, ok := redisClients.LoadAndDelete(connectionName); ok {
		redisClient := client.(*redis.Client)
		return redisClient.Close()
	}
	return nil
}

// CloseAllRedisClients 关闭所有缓存的 Redis 客户端
func CloseAllRedisClients() error {
	var errs []error
	redisClients.Range(func(key, value any) bool {
		client := value.(*redis.Client)
		if err := client.Close(); err != nil {
			errs = append(errs, err)
		}
		redisClients.Delete(key)
		return true
	})

	if len(errs) > 0 {
		return fmt.Errorf("关闭 Redis 客户端时发生错误: %v", errs)
	}
	return nil
}
