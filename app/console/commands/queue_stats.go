package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
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
		ctx.Warning("Redis 驱动暂不支持统计查询")
		ctx.Info("提示：Redis 队列数据存储在 Redis 中，需要直接访问 Redis 客户端来查询")
		ctx.Info("可以使用以下 Redis 命令查询：")
		defaultQueue := facades.Config().GetString(fmt.Sprintf("queue.connections.%s.queue", connectionName), "default")
		if queueName == "" {
			queueName = defaultQueue
		}
		ctx.Info(fmt.Sprintf("  LLEN queues:%s (待执行队列长度)", queueName))
		ctx.Info(fmt.Sprintf("  HLEN queues:%s:reserved (正在执行队列长度)", queueName))
		ctx.Info(fmt.Sprintf("  ZCARD queues:%s:delayed (延迟队列长度)", queueName))
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
