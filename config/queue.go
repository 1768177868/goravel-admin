package config

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	redisfacades "github.com/goravel/redis/facades"
)

func init() {
	config := facades.Config()
	config.Add("queue", map[string]any{
		// Default Queue Connection Name
		"default": config.Env("QUEUE_CONNECTION", "sync"),
		// Queue Connections
		//
		// Here you may configure the connection information for each server that is used by your application.
		// Drivers: "sync", "database", "custom"
		"connections": map[string]any{
			"sync": map[string]any{
				"driver": "sync",
			},
			"database": map[string]any{
				"driver":     "database",
				"connection": "sqlite",
				"queue":      "default",
				"concurrent": 1,
				// "tries": 3,        // 最大重试次数（可选，默认由队列工作进程设置）
				// "retry_after": 90, // 重试延迟时间（秒，可选）
			},
			"machinery": map[string]any{
				"driver":     "machinery",
				"connection": "default",
				"queue":      "default",
				"concurrent": 1,
				// "tries": 3,        // 最大重试次数（可选，默认由队列工作进程设置）
				// "retry_after": 90, // 重试延迟时间（秒，可选）
			},
			"redis1": map[string]any{
				"driver":     "custom",
				"connection": "default",
				"queue":      "default",
				"via": func() (queue.Driver, error) {
					return redisfacades.Queue("redis1") // The `redis` value is the key of `connections`
				},
			},
			"redis": map[string]any{
				"driver":     "custom",
				"connection": "default",
				"queue":      "default",
				"via": func() (queue.Driver, error) {
					return redisfacades.Queue("redis") // The `redis` value is the key of `connections`
				},
			},
		},
		// Failed Queue Jobs
		//
		// These options configure the behavior of failed queue job logging so you
		// can control how and where failed jobs are stored.
		"failed": map[string]any{
			"database": config.Env("DB_CONNECTION", "postgres"),
			"table":    "failed_jobs",
		},
		// Retry Configuration
		//
		// 队列工作进程的最大重试次数
		// 可以通过环境变量 QUEUE_TRIES 设置，默认值为 10
		// 注意：这个值是上限，实际重试次数由每个 Job 的 ShouldRetry 方法决定
		"tries": config.Env("QUEUE_TRIES", 10), // 最大重试次数上限（建议设置较大值，如 10）

		// Concurrent Configuration
		//
		// 队列工作进程的并发数（同时处理的任务数量）
		// 可以通过环境变量 QUEUE_CONCURRENT 设置，默认值为 1
		// 建议值：根据服务器性能和任务特性调整（1-10 或更多）
		"concurrent": config.Env("QUEUE_CONCURRENT", 3), // 并发数（同时处理的任务数量）
	})
}
