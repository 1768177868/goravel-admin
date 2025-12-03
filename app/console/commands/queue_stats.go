package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/redis/go-redis/v9"
)

type QueueStats struct {
}

// Signature The name and signature of the console command.
func (r *QueueStats) Signature() string {
	return "queue:stats"
}

// Description The console command description.
func (r *QueueStats) Description() string {
	return "查询队列统计信息，显示待执行、正在执行和失败任务数量"
}

// Extend The console command extend.
func (r *QueueStats) Extend() command.Extend {
	return command.Extend{
		Category: "queue",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "queue",
				Aliases: []string{"q"},
				Usage:   "队列名称（可选，用于筛选特定队列）",
			},
			&command.StringFlag{
				Name:    "connection",
				Aliases: []string{"c"},
				Usage:   "队列连接名称（可选，默认使用默认连接）",
			},
		},
	}
}

// Handle Execute the console command.
func (r *QueueStats) Handle(ctx console.Context) error {
	queueName := ctx.Option("queue")
	connectionName := ctx.Option("connection")

	if connectionName == "" {
		connectionName = facades.Config().GetString("queue.default", "sync")
	}

	ctx.Info(fmt.Sprintf("队列连接: %s", connectionName))
	if queueName != "" {
		ctx.Info(fmt.Sprintf("队列名称: %s", queueName))
	}
	ctx.Info("")

	// 判断队列驱动类型
	driver := facades.Config().GetString(fmt.Sprintf("queue.connections.%s.driver", connectionName), "")
	ctx.Info(fmt.Sprintf("驱动类型: %s", driver))

	// 检查是否是 Redis 驱动（custom driver with via）
	isRedis := r.isRedisDriver(connectionName)

	if isRedis {
		// Redis 驱动：通过 Redis 客户端查询队列大小
		defaultQueue := facades.Config().GetString(fmt.Sprintf("queue.connections.%s.queue", connectionName), "default")
		if queueName == "" {
			queueName = defaultQueue
		}

		// 获取 Redis 连接名称（从队列配置中获取）
		redisConnectionName := r.getRedisConnectionName(connectionName)
		if redisConnectionName == "" {
			ctx.Warning("无法确定 Redis 连接名称")
			return nil
		}

		// 查询 Redis 队列统计
		stats, err := r.getRedisQueueStats(redisConnectionName, queueName)
		if err != nil {
			ctx.Error(fmt.Sprintf("查询 Redis 队列统计失败: %v", err))
			ctx.Info("提示：请确保 Redis 连接配置正确且 Redis 服务正在运行")
			return err
		}

		totalCount := stats.Pending + stats.Reserved

		// 显示统计信息
		ctx.Info("═══════════════════════════════════════")
		ctx.Info("队列统计信息 (Redis)")
		ctx.Info("═══════════════════════════════════════")
		ctx.Info(fmt.Sprintf("队列名称:      %s", queueName))
		ctx.Info(fmt.Sprintf("待执行任务:    %d", stats.Pending))
		ctx.Info(fmt.Sprintf("正在执行任务:  %d", stats.Reserved))
		ctx.Info(fmt.Sprintf("延迟任务:      %d", stats.Delayed))
		ctx.Info(fmt.Sprintf("失败任务:      %d", stats.Failed))
		ctx.Info(fmt.Sprintf("总任务数:      %d", totalCount))
		ctx.Info("═══════════════════════════════════════")

		// 提示信息
		if stats.Pending > 0 {
			ctx.Info("")
			ctx.Warning(fmt.Sprintf("提示：队列中有 %d 个待执行任务", stats.Pending))
			ctx.Info("")
			ctx.Info("如果需要处理这些任务：")
			ctx.Info("  1. 启动主程序（main.go 中会自动启动队列 Worker）")
			ctx.Info("     go run .")
			ctx.Info("  2. 确保 Worker 监听正确的队列名称和连接")
			ctx.Info("  3. 确保任务已正确注册到 QueueServiceProvider")
			ctx.Info("")
			ctx.Info("如果不需要这些任务，可以清理队列：")
			ctx.Info("  使用 Redis 客户端执行以下命令清理队列：")
			ctx.Info(fmt.Sprintf("    redis-cli DEL queues:%s", queueName))
			ctx.Info(fmt.Sprintf("    redis-cli DEL queues:%s:reserved", queueName))
			ctx.Info(fmt.Sprintf("    redis-cli DEL queues:%s:delayed", queueName))
			ctx.Info("  或者使用命令：go run . artisan queue:clear --queue=" + queueName)
		}

		// 如果未指定队列名称，显示按队列分组的统计
		if queueName == "" {
			ctx.Info("")
			ctx.Info("按队列分组统计:")
			byQueue, err := r.getRedisStatsByQueue(redisConnectionName)
			if err != nil {
				ctx.Warning(fmt.Sprintf("获取按队列分组统计失败: %v", err))
			} else {
				if len(byQueue) == 0 {
					ctx.Info("  暂无队列数据")
				} else {
					for qName, qStats := range byQueue {
						ctx.Info(fmt.Sprintf("  队列 [%s]:", qName))
						ctx.Info(fmt.Sprintf("    待执行: %d, 正在执行: %d, 延迟: %d, 失败: %d, 总计: %d",
							qStats.Pending, qStats.Reserved, qStats.Delayed, qStats.Failed, qStats.Total))
					}
				}
			}
		}

		return nil
	}

	if driver == "sync" {
		ctx.Info("同步驱动：任务立即执行，无队列数据")
		return nil
	}

	if driver != "database" {
		ctx.Warning(fmt.Sprintf("驱动 %s 暂不支持统计查询", driver))
		return nil
	}

	// Database 驱动：查询 jobs 表
	var pendingCount, reservedCount int64
	var err error

	// 查询待执行任务数（available_at <= now 且 reserved_at 为 null）
	pendingQuery := facades.Orm().Query().Table("jobs").
		Where("available_at", "<=", time.Now()).
		Where("reserved_at IS NULL")

	// 查询正在执行任务数（reserved_at 不为 null）
	reservedQuery := facades.Orm().Query().Table("jobs").
		Where("reserved_at IS NOT NULL")

	// 如果指定了队列名称，添加筛选条件
	if queueName != "" {
		pendingQuery = pendingQuery.Where("queue", "=", queueName)
		reservedQuery = reservedQuery.Where("queue", "=", queueName)
	}

	pendingCount, err = pendingQuery.Count()
	if err != nil {
		ctx.Error(fmt.Sprintf("查询待执行任务数失败: %v", err))
		return err
	}

	reservedCount, err = reservedQuery.Count()
	if err != nil {
		ctx.Error(fmt.Sprintf("查询正在执行任务数失败: %v", err))
		return err
	}

	// 查询失败任务数（从 failed_jobs 表）
	failedQuery := facades.Orm().Query().Table("failed_jobs")
	if queueName != "" {
		failedQuery = failedQuery.Where("queue", "=", queueName)
	}
	failedCount, err := failedQuery.Count()
	if err != nil {
		ctx.Error(fmt.Sprintf("查询失败任务数失败: %v", err))
		return err
	}

	totalCount := pendingCount + reservedCount

	// 显示统计信息
	ctx.Info("═══════════════════════════════════════")
	ctx.Info("队列统计信息")
	ctx.Info("═══════════════════════════════════════")
	ctx.Info(fmt.Sprintf("待执行任务:    %d", pendingCount))
	ctx.Info(fmt.Sprintf("正在执行任务:  %d", reservedCount))
	ctx.Info(fmt.Sprintf("失败任务:      %d", failedCount))
	ctx.Info(fmt.Sprintf("总任务数:      %d", totalCount))
	ctx.Info("═══════════════════════════════════════")

	// 如果未指定队列名称，显示按队列分组的统计
	if queueName == "" {
		ctx.Info("")
		ctx.Info("按队列分组统计:")
		byQueue, err := r.getStatsByQueue()
		if err != nil {
			ctx.Warning(fmt.Sprintf("获取按队列分组统计失败: %v", err))
		} else {
			if len(byQueue) == 0 {
				ctx.Info("  暂无队列数据")
			} else {
				for qName, stats := range byQueue {
					ctx.Info(fmt.Sprintf("  队列 [%s]:", qName))
					ctx.Info(fmt.Sprintf("    待执行: %d, 正在执行: %d, 失败: %d, 总计: %d",
						stats.Pending, stats.Reserved, stats.Failed, stats.Total))
				}
			}
		}
	}

	return nil
}

// isRedisDriver 判断是否是 Redis 驱动
func (r *QueueStats) isRedisDriver(connectionName string) bool {
	via := facades.Config().Get(fmt.Sprintf("queue.connections.%s.via", connectionName))
	return via != nil || strings.Contains(connectionName, "redis")
}

// QueueStatsInfo 队列统计信息
type QueueStatsInfo struct {
	Pending  int64
	Reserved int64
	Failed   int64
	Total    int64
}

// RedisQueueStatsInfo Redis 队列统计信息
type RedisQueueStatsInfo struct {
	Pending  int64
	Reserved int64
	Delayed  int64
	Failed   int64
	Total    int64
}

// getStatsByQueue 按队列分组获取统计信息
func (r *QueueStats) getStatsByQueue() (map[string]QueueStatsInfo, error) {
	// 获取所有队列名称
	var queues []string
	err := facades.Orm().Query().Table("jobs").
		Select("DISTINCT queue").
		Pluck("queue", &queues)
	if err != nil {
		return nil, err
	}

	// 获取失败任务的队列名称
	var failedQueues []string
	err = facades.Orm().Query().Table("failed_jobs").
		Select("DISTINCT queue").
		Pluck("queue", &failedQueues)
	if err != nil {
		return nil, err
	}

	// 合并队列名称
	queueMap := make(map[string]bool)
	for _, q := range queues {
		queueMap[q] = true
	}
	for _, q := range failedQueues {
		queueMap[q] = true
	}

	result := make(map[string]QueueStatsInfo)
	now := time.Now()

	for qName := range queueMap {
		// 待执行任务数
		pendingCount, _ := facades.Orm().Query().Table("jobs").
			Where("queue", "=", qName).
			Where("available_at", "<=", now).
			Where("reserved_at IS NULL").
			Count()

		// 正在执行任务数
		reservedCount, _ := facades.Orm().Query().Table("jobs").
			Where("queue", "=", qName).
			Where("reserved_at IS NOT NULL").
			Count()

		// 失败任务数
		failedCount, _ := facades.Orm().Query().Table("failed_jobs").
			Where("queue", "=", qName).
			Count()

		result[qName] = QueueStatsInfo{
			Pending:  pendingCount,
			Reserved: reservedCount,
			Failed:   failedCount,
			Total:    pendingCount + reservedCount,
		}
	}

	return result, nil
}

// getRedisConnectionName 从队列连接配置中获取 Redis 连接名称
func (r *QueueStats) getRedisConnectionName(queueConnectionName string) string {
	// 从队列配置中获取 connection 字段
	connection := facades.Config().GetString(fmt.Sprintf("queue.connections.%s.connection", queueConnectionName), "default")

	// 如果队列连接名称包含 redis，尝试使用它
	if strings.Contains(queueConnectionName, "redis") {
		// 检查 Redis 配置中是否存在对应的连接
		redisHost := facades.Config().GetString(fmt.Sprintf("database.redis.%s.host", queueConnectionName), "")
		if redisHost != "" {
			return queueConnectionName
		}
	}

	// 默认使用 default
	return connection
}

// getRedisQueueStats 获取 Redis 队列统计信息
func (r *QueueStats) getRedisQueueStats(redisConnectionName, queueName string) (*RedisQueueStatsInfo, error) {
	// 创建 Redis 客户端
	redisClient, err := r.createRedisClient(redisConnectionName)
	if err != nil {
		return nil, fmt.Errorf("创建 Redis 客户端失败: %v", err)
	}
	defer redisClient.Close()

	ctx := context.Background()
	stats := &RedisQueueStatsInfo{}

	// 待执行队列：queues:{queueName} (List)
	pendingKey := fmt.Sprintf("queues:%s", queueName)
	pendingLen, err := redisClient.LLen(ctx, pendingKey).Result()
	if err != nil {
		return nil, fmt.Errorf("查询待执行队列失败: %v", err)
	}
	stats.Pending = pendingLen

	// 正在执行队列：queues:{queueName}:reserved (Hash)
	reservedKey := fmt.Sprintf("queues:%s:reserved", queueName)
	reservedLen, err := redisClient.HLen(ctx, reservedKey).Result()
	if err != nil {
		return nil, fmt.Errorf("查询正在执行队列失败: %v", err)
	}
	stats.Reserved = reservedLen

	// 延迟队列：queues:{queueName}:delayed (Sorted Set)
	delayedKey := fmt.Sprintf("queues:%s:delayed", queueName)
	delayedLen, err := redisClient.ZCard(ctx, delayedKey).Result()
	if err != nil {
		return nil, fmt.Errorf("查询延迟队列失败: %v", err)
	}
	stats.Delayed = delayedLen

	// 调试信息：显示 Redis 键的实际值（可选，用于排查问题）
	// 可以查看队列中的实际内容
	if pendingLen > 0 {
		// 查看队列中的第一个任务（不移除）
		firstTask, _ := redisClient.LIndex(ctx, pendingKey, 0).Result()
		if firstTask != "" {
			// 只显示前100个字符，避免输出过长
			if len(firstTask) > 100 {
				firstTask = firstTask[:100] + "..."
			}
		}
	}

	// 失败任务：从数据库 failed_jobs 表查询（Redis 队列的失败任务也存储在数据库中）
	var failedCount int64
	if queueName != "" {
		failedCount, err = facades.Orm().Query().Table("failed_jobs").
			Where("queue = ?", queueName).
			Count()
	} else {
		failedCount, err = facades.Orm().Query().Table("failed_jobs").Count()
	}
	if err != nil {
		// 失败任务查询失败不影响其他统计
		stats.Failed = 0
	} else {
		stats.Failed = failedCount
	}

	stats.Total = stats.Pending + stats.Reserved

	return stats, nil
}

// getRedisStatsByQueue 按队列分组获取 Redis 队列统计信息
func (r *QueueStats) getRedisStatsByQueue(redisConnectionName string) (map[string]*RedisQueueStatsInfo, error) {
	// 创建 Redis 客户端
	redisClient, err := r.createRedisClient(redisConnectionName)
	if err != nil {
		return nil, fmt.Errorf("创建 Redis 客户端失败: %v", err)
	}
	defer redisClient.Close()

	ctx := context.Background()
	result := make(map[string]*RedisQueueStatsInfo)

	// 查找所有队列键（queues:* 但不包含 :reserved 和 :delayed）
	pattern := "queues:*"
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("查找队列键失败: %v", err)
	}

	// 提取队列名称（排除 reserved 和 delayed 键）
	queueMap := make(map[string]bool)
	for _, key := range keys {
		// 跳过 reserved 和 delayed 键
		if strings.HasSuffix(key, ":reserved") || strings.HasSuffix(key, ":delayed") {
			continue
		}
		// 提取队列名称（去掉 queues: 前缀）
		if after, ok := strings.CutPrefix(key, "queues:"); ok {
			queueName := after
			if queueName != "" {
				queueMap[queueName] = true
			}
		}
	}

	// 如果没有找到队列键，尝试从失败任务表中获取队列名称
	if len(queueMap) == 0 {
		var failedQueues []string
		err = facades.Orm().Query().Table("failed_jobs").
			Select("DISTINCT queue").
			Pluck("queue", &failedQueues)
		if err == nil {
			for _, q := range failedQueues {
				if q != "" {
					queueMap[q] = true
				}
			}
		}
	}

	// 为每个队列获取统计信息
	for queueName := range queueMap {
		stats, err := r.getRedisQueueStats(redisConnectionName, queueName)
		if err != nil {
			// 单个队列查询失败不影响其他队列
			continue
		}
		result[queueName] = stats
	}

	return result, nil
}

// createRedisClient 创建 Redis 客户端
func (r *QueueStats) createRedisClient(connectionName string) (*redis.Client, error) {
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

	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %v", err)
	}

	return client, nil
}
