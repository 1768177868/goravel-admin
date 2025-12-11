package admin

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"goravel/app/http/response"
	"goravel/app/utils/errorlog"
)

// 监控数据缓存（短期缓存，减少系统调用）
var (
	monitorCache     map[string]any
	monitorCacheLock sync.RWMutex
	monitorCacheTime time.Time
	cacheDuration    = 1 * time.Second // 缓存1秒，SSE推送间隔2秒时可以减少一半的系统调用
)

// 网络带宽监控缓存（用于计算速度和峰值）
var (
	lastNetBytesSent  uint64
	lastNetBytesRecv  uint64
	lastNetSampleTime time.Time
	peakSentSpeed     float64 // 峰值发送速度
	peakRecvSpeed     float64 // 峰值接收速度
	peakTotalSpeed    float64 // 峰值总速度
	netSpeedLock      sync.RWMutex
)

// getNetworkSpeed 计算网络速度并记录峰值（需要两次采样）
func getNetworkSpeed(currentBytesSent, currentBytesRecv uint64) (sentSpeed, recvSpeed, totalSpeed, peakSent, peakRecv, peakTotal float64) {
	netSpeedLock.Lock()
	defer netSpeedLock.Unlock()

	now := time.Now()

	// 如果是第一次采样，保存当前值并返回0
	if lastNetSampleTime.IsZero() {
		lastNetBytesSent = currentBytesSent
		lastNetBytesRecv = currentBytesRecv
		lastNetSampleTime = now
		return 0, 0, 0, 0, 0, 0
	}

	// 计算时间差（秒）
	timeDiff := now.Sub(lastNetSampleTime).Seconds()
	if timeDiff <= 0 {
		return sentSpeed, recvSpeed, totalSpeed, peakSentSpeed, peakRecvSpeed, peakTotalSpeed
	}

	// 计算速度（字节/秒）
	var sentSpeedBps, recvSpeedBps float64
	if currentBytesSent >= lastNetBytesSent {
		sentSpeedBps = float64(currentBytesSent-lastNetBytesSent) / timeDiff
	}
	if currentBytesRecv >= lastNetBytesRecv {
		recvSpeedBps = float64(currentBytesRecv-lastNetBytesRecv) / timeDiff
	}

	// 更新缓存
	lastNetBytesSent = currentBytesSent
	lastNetBytesRecv = currentBytesRecv
	lastNetSampleTime = now

	// 转换为Mbps（兆比特每秒）：字节/秒 * 8 / 1024 / 1024
	sentSpeed = sentSpeedBps * 8 / 1024 / 1024
	recvSpeed = recvSpeedBps * 8 / 1024 / 1024
	totalSpeed = sentSpeed + recvSpeed

	// 更新峰值
	if sentSpeed > peakSentSpeed {
		peakSentSpeed = sentSpeed
	}
	if recvSpeed > peakRecvSpeed {
		peakRecvSpeed = recvSpeed
	}
	if totalSpeed > peakTotalSpeed {
		peakTotalSpeed = totalSpeed
	}

	return sentSpeed, recvSpeed, totalSpeed, peakSentSpeed, peakRecvSpeed, peakTotalSpeed
}

type MonitorController struct{}

func NewMonitorController() *MonitorController {
	return &MonitorController{}
}

// convertProcessStatus 将 Linux 进程状态代码转换为友好的状态文本
func convertProcessStatus(statusCode string, processName string) string {
	// Linux 进程状态代码：
	// R = Running (运行中)
	// S = Sleeping (可中断的睡眠)
	// D = Disk sleep (不可中断的睡眠，等待I/O)
	// Z = Zombie (僵尸进程)
	// T = Stopped (停止)
	// I = Idle (空闲)

	statusCode = strings.ToUpper(strings.TrimSpace(statusCode))

	switch statusCode {
	case "R", "RUNNING":
		return "running"
	case "S", "SLEEPING", "SLEEP":
		// 对于服务进程（MySQL、PostgreSQL、Redis、应用），sleep 状态是正常的，显示为 running
		if processName == "mysql" || processName == "postgresql" || processName == "redis" || processName == "app" {
			return "running"
		}
		return "sleep"
	case "D", "DISK SLEEP":
		return "running" // 等待I/O也是运行状态的一种
	case "Z", "ZOMBIE":
		return "zombie"
	case "T", "STOPPED":
		return "stopped"
	case "I", "IDLE":
		return "running" // 空闲也是运行状态
	default:
		// 如果无法识别，对于服务进程默认显示 running
		if processName == "mysql" || processName == "postgresql" || processName == "redis" || processName == "app" {
			return "running"
		}
		return statusCode
	}
}

// getProcessInfo 获取指定进程的 CPU 和内存信息
func getProcessInfo(ctx http.Context, processName string, pid int32) map[string]any {
	result := map[string]any{
		"name":   processName,
		"pid":    pid,
		"cpu":    0.0,
		"memory": 0,
		"status": "not_found",
		"rss":    0, // 物理内存占用
		"vms":    0, // 虚拟内存占用
	}

	if pid <= 0 {
		return result
	}

	proc, err := process.NewProcess(pid)
	if err != nil {
		return result
	}

	// 获取 CPU 使用率
	cpuPercent, err := proc.CPUPercent()
	if err == nil {
		result["cpu"] = cpuPercent
	}

	// 获取内存信息
	memInfo, err := proc.MemoryInfo()
	if err == nil {
		result["memory"] = memInfo.RSS // 物理内存占用（字节）
		result["rss"] = memInfo.RSS
		result["vms"] = memInfo.VMS
	}

	// 获取进程状态并转换
	status, err := proc.Status()
	if err == nil && len(status) > 0 {
		// 转换状态代码为友好的文本
		result["status"] = convertProcessStatus(status[0], processName)
	} else {
		// 如果无法获取状态，对于服务进程默认显示 running
		if processName == "mysql" || processName == "postgresql" || processName == "redis" || processName == "app" {
			result["status"] = "running"
		} else {
			result["status"] = "unknown"
		}
	}

	// 获取进程创建时间
	createTime, err := proc.CreateTime()
	if err == nil {
		result["create_time"] = createTime
	}

	// 获取进程名
	name, err := proc.Name()
	if err == nil {
		result["process_name"] = name
	}

	return result
}

// findProcessByName 根据进程名查找进程 PID
// 优先匹配更具体的进程名（如 mysqld 优先于 mysql）
func findProcessByName(ctx http.Context, processNames []string) int32 {
	processes, err := process.Processes()
	if err != nil {
		return 0
	}

	// 先尝试精确匹配（更具体的进程名）
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}
		nameLower := strings.ToLower(name)

		// 优先匹配更具体的进程名
		for _, targetName := range processNames {
			targetLower := strings.ToLower(targetName)
			// 精确匹配或包含匹配
			if nameLower == targetLower || strings.Contains(nameLower, targetLower) {
				// 对于 MySQL，优先选择 mysqld 而不是 mysql
				if targetLower == "mysqld" && nameLower == "mysqld" {
					return proc.Pid
				}
				// 对于 Redis，优先选择 redis-server
				if targetLower == "redis-server" && strings.Contains(nameLower, "redis-server") {
					return proc.Pid
				}
			}
		}
	}

	// 如果精确匹配失败，尝试通过命令行参数匹配
	for _, proc := range processes {
		cmdline, err := proc.Cmdline()
		if err != nil {
			continue
		}
		cmdlineLower := strings.ToLower(cmdline)

		for _, targetName := range processNames {
			targetLower := strings.ToLower(targetName)
			if strings.Contains(cmdlineLower, targetLower) {
				// 排除掉一些明显不是目标进程的情况
				if targetLower == "mysql" && strings.Contains(cmdlineLower, "mysqladmin") {
					continue
				}
				return proc.Pid
			}
		}
	}

	return 0
}

// isLocalHost 检查地址是否为本地地址
func isLocalHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "" || host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0.0.0.0"
}

// getMySQLInfoFromDB 通过数据库连接获取MySQL信息
func getMySQLInfoFromDB(ctx http.Context) map[string]any {
	result := map[string]any{
		"name":        "mysql",
		"type":        "remote",
		"status":      "disconnected",
		"version":     "",
		"uptime":      0,
		"threads":     0,
		"queries":     0,
		"connections": 0,
	}

	defer func() {
		if r := recover(); r != nil {
			result["status"] = "error"
		}
	}()

	// 获取数据库连接配置
	dbConnection := facades.Config().GetString("database.default", "sqlite")
	dbHost := facades.Config().GetString(fmt.Sprintf("database.connections.%s.host", dbConnection), "127.0.0.1")
	dbPort := facades.Config().GetInt(fmt.Sprintf("database.connections.%s.port", dbConnection), 3306)

	// 检查是否为本地数据库
	if !isLocalHost(dbHost) {
		result["host"] = fmt.Sprintf("%s:%d", dbHost, dbPort)
		result["type"] = "remote"
		// 云数据库无法获取进程信息（CPU、内存、PID等），不设置这些字段
	} else {
		result["host"] = fmt.Sprintf("%s:%d", dbHost, dbPort)
		result["type"] = "local"
		// 本地数据库可能会通过进程监控获取CPU、内存等信息，但这里先不设置
	}

	// 尝试连接数据库获取信息
	orm := facades.Orm()
	if orm == nil {
		return result
	}

	// 检查连接类型是否为MySQL
	if dbConnection != "mysql" {
		result["status"] = "not_mysql"
		return result
	}

	// 执行MySQL状态查询
	var version string
	var uptime, threads, queries, connections int64

	// 获取MySQL版本
	query := orm.Query()
	if query != nil {
		if err := query.Raw("SELECT VERSION() as version").Scan(&version); err == nil {
			result["version"] = version
		}

		// 获取MySQL状态信息
		var statusRows []map[string]any
		if err := query.Raw("SHOW STATUS WHERE Variable_name IN ('Uptime', 'Threads_connected', 'Questions')").Scan(&statusRows); err == nil {
			for _, row := range statusRows {
				if variableName, ok := row["Variable_name"].(string); ok {
					if value, ok := row["Value"].(string); ok {
						intValue, _ := strconv.ParseInt(value, 10, 64)
						switch variableName {
						case "Uptime":
							uptime = intValue
						case "Threads_connected":
							// Threads_connected 是当前连接数，应该用作 connections
							connections = intValue
							threads = intValue // 线程数也使用当前连接数
						case "Questions":
							queries = intValue
						}
					}
				}
			}
			result["uptime"] = uptime
			result["threads"] = threads
			result["queries"] = queries
			result["connections"] = connections // 当前连接数
		}

		// 获取MySQL变量信息（内存相关）
		var variableRows []map[string]any
		if err := query.Raw("SHOW VARIABLES WHERE Variable_name IN ('max_connections', 'innodb_buffer_pool_size', 'slow_query_log', 'long_query_time')").Scan(&variableRows); err == nil {
			for _, row := range variableRows {
				if variableName, ok := row["Variable_name"].(string); ok {
					if value, ok := row["Value"].(string); ok {
						intValue, _ := strconv.ParseInt(value, 10, 64)
						if variableName == "innodb_buffer_pool_size" {
							result["buffer_pool_size"] = intValue
						} else if variableName == "max_connections" {
							result["max_connections"] = intValue
						} else if variableName == "slow_query_log" {
							result["slow_query_log"] = value
						} else if variableName == "long_query_time" {
							if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
								result["long_query_time"] = floatValue
							}
						}
					}
				}
			}
		}

		// 获取MySQL更多状态信息（慢查询、锁等待等）
		var moreStatusRows []map[string]any
		if err := query.Raw("SHOW STATUS WHERE Variable_name IN ('Slow_queries', 'Table_locks_waited', 'Innodb_row_lock_waits', 'Innodb_row_lock_time_avg', 'Threads_running', 'Threads_created', 'Aborted_connects')").Scan(&moreStatusRows); err == nil {
			for _, row := range moreStatusRows {
				if variableName, ok := row["Variable_name"].(string); ok {
					if value, ok := row["Value"].(string); ok {
						intValue, _ := strconv.ParseInt(value, 10, 64)
						switch variableName {
						case "Slow_queries":
							result["slow_queries"] = intValue
						case "Table_locks_waited":
							result["table_locks_waited"] = intValue
						case "Innodb_row_lock_waits":
							result["innodb_row_lock_waits"] = intValue
						case "Innodb_row_lock_time_avg":
							result["innodb_row_lock_time_avg"] = intValue
						case "Threads_running":
							result["threads_running"] = intValue
						case "Threads_created":
							result["threads_created"] = intValue
						case "Aborted_connects":
							result["aborted_connects"] = intValue
						}
					}
				}
			}
		}
	}

	result["status"] = "connected"
	return result
}

// getPostgreSQLInfoFromDB 通过数据库连接获取PostgreSQL信息
func getPostgreSQLInfoFromDB(ctx http.Context) map[string]any {
	result := map[string]any{
		"name":            "postgresql",
		"type":            "remote",
		"status":          "disconnected",
		"version":         "",
		"uptime":          0,
		"connections":     0,
		"max_connections": 0,
	}

	defer func() {
		if r := recover(); r != nil {
			result["status"] = "error"
		}
	}()

	// 获取数据库连接配置
	dbConnection := facades.Config().GetString("database.default", "sqlite")
	dbHost := facades.Config().GetString(fmt.Sprintf("database.connections.%s.host", dbConnection), "127.0.0.1")
	dbPort := facades.Config().GetInt(fmt.Sprintf("database.connections.%s.port", dbConnection), 5432)

	// 检查是否为本地数据库
	if !isLocalHost(dbHost) {
		result["host"] = fmt.Sprintf("%s:%d", dbHost, dbPort)
		result["type"] = "remote"
	} else {
		result["host"] = fmt.Sprintf("%s:%d", dbHost, dbPort)
		result["type"] = "local"
	}

	// 尝试连接数据库获取信息
	orm := facades.Orm()
	if orm == nil {
		return result
	}

	// 检查连接类型是否为PostgreSQL
	if dbConnection != "postgres" {
		result["status"] = "not_postgres"
		return result
	}

	// 执行PostgreSQL查询
	query := orm.Query()
	if query != nil {
		// 获取PostgreSQL版本
		var version string
		if err := query.Raw("SELECT version() as version").Scan(&version); err == nil {
			// 提取版本号（例如：PostgreSQL 14.5 on x86_64-pc-linux-gnu）
			if strings.Contains(version, "PostgreSQL") {
				parts := strings.Fields(version)
				if len(parts) >= 2 {
					result["version"] = parts[1]
				} else {
					result["version"] = version
				}
			} else {
				result["version"] = version
			}
		}

		// 获取PostgreSQL运行时间（秒）
		var uptime int64
		if err := query.Raw("SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))::bigint as uptime").Scan(&uptime); err == nil {
			result["uptime"] = uptime
		}

		// 获取当前连接数
		var connections int64
		if err := query.Raw("SELECT count(*) FROM pg_stat_activity").Scan(&connections); err == nil {
			result["connections"] = connections
		}

		// 获取最大连接数
		var maxConnections int64
		if err := query.Raw("SELECT setting::bigint FROM pg_settings WHERE name = 'max_connections'").Scan(&maxConnections); err == nil {
			result["max_connections"] = maxConnections
		}

		// 获取数据库大小
		var dbSize int64
		if err := query.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize); err == nil {
			result["database_size"] = dbSize
		}

		// 获取活跃连接数
		var activeConnections int64
		if err := query.Raw("SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConnections); err == nil {
			result["active_connections"] = activeConnections
		}

		// 获取空闲连接数
		var idleConnections int64
		if err := query.Raw("SELECT count(*) FROM pg_stat_activity WHERE state = 'idle'").Scan(&idleConnections); err == nil {
			result["idle_connections"] = idleConnections
		}

		// 获取总查询数（从启动开始）
		var totalQueries int64
		if err := query.Raw("SELECT sum(xact_commit + xact_rollback) FROM pg_stat_database WHERE datname = current_database()").Scan(&totalQueries); err == nil {
			result["queries"] = totalQueries
		}
	}

	result["status"] = "connected"
	return result
}

// getRedisInfoFromConnection 通过Redis连接获取Redis信息
func getRedisInfoFromConnection(ctx http.Context) map[string]any {
	result := map[string]any{
		"name":                     "redis",
		"type":                     "remote",
		"status":                   "disconnected",
		"version":                  "",
		"used_memory":              0,
		"used_memory_human":        "",
		"connected_clients":        0,
		"total_commands_processed": 0,
		"keyspace_hits":            0,
		"keyspace_misses":          0,
	}

	defer func() {
		if r := recover(); r != nil {
			result["status"] = "error"
		}
	}()

	// 获取Redis连接配置
	redisHost := facades.Config().GetString("database.redis.default.host", "")
	redisPort := facades.Config().GetInt("database.redis.default.port", 6379)
	redisPassword := facades.Config().GetString("database.redis.default.password", "")
	redisDB := facades.Config().GetInt("database.redis.default.database", 0)

	// 检查是否为本地Redis
	if !isLocalHost(redisHost) {
		result["host"] = fmt.Sprintf("%s:%d", redisHost, redisPort)
		result["type"] = "remote"
		// 云数据库无法获取进程信息（CPU、PID等），不设置这些字段
		// 但可以通过INFO命令获取内存使用情况
	} else {
		result["host"] = fmt.Sprintf("%s:%d", redisHost, redisPort)
		result["type"] = "local"
		// 本地Redis可能会通过进程监控获取CPU等信息，但这里先不设置
	}

	// 创建Redis客户端直接连接（用于执行INFO命令）
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	// 测试连接
	redisCtx := context.Background()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		result["status"] = "disconnected"
		return result
	}

	result["status"] = "connected"

	// 执行INFO命令获取详细信息
	infoStr, err := redisClient.Info(redisCtx).Result()
	if err == nil && infoStr != "" {
		// 解析Redis INFO命令返回的信息
		parseRedisInfo(infoStr, result)
		// 将 used_memory 也设置到 memory 字段（用于兼容）
		if usedMemory, ok := result["used_memory"].(int64); ok {
			result["memory"] = usedMemory
		}
	}

	return result
}

// parseRedisInfo 解析Redis INFO命令返回的字符串
func parseRedisInfo(infoStr string, result map[string]any) {
	lines := strings.Split(infoStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "redis_version":
			result["version"] = value
		case "used_memory":
			if mem, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["used_memory"] = mem
			}
		case "used_memory_human":
			result["used_memory_human"] = value
		case "used_memory_peak":
			if mem, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["used_memory_peak"] = mem
			}
		case "used_memory_rss":
			if mem, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["used_memory_rss"] = mem
			}
		case "connected_clients":
			if clients, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["connected_clients"] = clients
			}
		case "total_commands_processed":
			if cmds, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["total_commands_processed"] = cmds
			}
		case "instantaneous_ops_per_sec":
			if ops, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["instantaneous_ops_per_sec"] = ops
			}
		case "keyspace_hits":
			if hits, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["keyspace_hits"] = hits
			}
		case "keyspace_misses":
			if misses, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["keyspace_misses"] = misses
			}
		case "expired_keys":
			if expired, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["expired_keys"] = expired
			}
		case "evicted_keys":
			if evicted, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["evicted_keys"] = evicted
			}
		case "uptime_in_seconds":
			if uptime, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["uptime"] = uptime
			}
		case "role":
			result["role"] = value
		case "connected_slaves":
			if slaves, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["connected_slaves"] = slaves
			}
		case "rdb_last_save_time":
			if saveTime, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["rdb_last_save_time"] = saveTime
			}
		case "aof_enabled":
			if enabled, err := strconv.ParseInt(value, 10, 64); err == nil {
				result["aof_enabled"] = enabled == 1
			}
		}
	}

	// 计算命中率
	if hits, ok := result["keyspace_hits"].(int64); ok {
		if misses, ok2 := result["keyspace_misses"].(int64); ok2 {
			total := hits + misses
			if total > 0 {
				hitRate := float64(hits) / float64(total) * 100
				result["keyspace_hit_rate"] = hitRate
			}
		}
	}
}

// getProcessesInfo 获取 MySQL、PostgreSQL、Redis 和当前应用进程的信息
func (r *MonitorController) getProcessesInfo(ctx http.Context) map[string]any {
	result := map[string]any{
		"mysql": map[string]any{
			"name":   "mysql",
			"pid":    0,
			"cpu":    0.0,
			"memory": 0,
			"status": "not_found",
			"rss":    0,
			"vms":    0,
		},
		"postgresql": map[string]any{
			"name":   "postgresql",
			"pid":    0,
			"cpu":    0.0,
			"memory": 0,
			"status": "not_found",
			"rss":    0,
			"vms":    0,
		},
		"redis": map[string]any{
			"name":   "redis",
			"pid":    0,
			"cpu":    0.0,
			"memory": 0,
			"status": "not_found",
			"rss":    0,
			"vms":    0,
		},
		"app": map[string]any{
			"name":   "app",
			"pid":    0,
			"cpu":    0.0,
			"memory": 0,
			"status": "not_found",
			"rss":    0,
			"vms":    0,
		},
	}

	// 使用 defer recover 确保即使进程查找出错也不影响整体功能
	defer func() {
		if r := recover(); r != nil {
			// 静默处理错误，返回默认值
		}
	}()

	// 获取数据库和Redis连接配置
	dbConnection := facades.Config().GetString("database.default", "sqlite")
	dbHost := facades.Config().GetString(fmt.Sprintf("database.connections.%s.host", dbConnection), "127.0.0.1")
	redisHost := facades.Config().GetString("database.redis.default.host", "")

	// PostgreSQL处理：检查是否为PostgreSQL数据库
	if dbConnection == "postgres" {
		// 无论本地还是远程，都先通过数据库连接获取统计信息
		postgresDBInfo := getPostgreSQLInfoFromDB(ctx)

		if isLocalHost(dbHost) {
			// 本地PostgreSQL，尝试查找进程获取 CPU、内存等信息
			postgresNames := []string{"postgres", "postmaster", "postgresql"}
			if runtime.GOOS == "windows" {
				postgresNames = []string{"postgres", "postgres.exe", "postmaster.exe"}
			}
			postgresPid := findProcessByName(ctx, postgresNames)
			if postgresPid > 0 {
				// 获取进程信息（CPU、内存等）
				processInfo := getProcessInfo(ctx, "postgresql", postgresPid)
				if processInfo != nil {
					// 合并进程信息和数据库统计信息
					for k, v := range postgresDBInfo {
						if _, exists := processInfo[k]; !exists {
							// 如果进程信息中没有这个字段，使用数据库信息
							processInfo[k] = v
						}
					}
					// 确保类型和状态正确
					processInfo["type"] = "local"
					if postgresDBInfo["status"] == "connected" {
						processInfo["status"] = "connected"
					}
					result["postgresql"] = processInfo
				} else {
					// 如果获取进程信息失败，使用数据库信息
					result["postgresql"] = postgresDBInfo
				}
			} else {
				// 找不到进程，使用数据库信息
				result["postgresql"] = postgresDBInfo
			}
		} else {
			// 远程PostgreSQL，直接使用数据库信息
			result["postgresql"] = postgresDBInfo
		}
	}

	// MySQL处理：检查是否为本地数据库
	if dbConnection == "mysql" {
		// 无论本地还是远程，都先通过数据库连接获取统计信息（连接数、线程数等）
		mysqlDBInfo := getMySQLInfoFromDB(ctx)

		if isLocalHost(dbHost) {
			// 本地MySQL，尝试查找进程获取 CPU、内存等信息
			mysqlNames := []string{"mysqld", "mysql", "mariadb"}
			if runtime.GOOS == "windows" {
				mysqlNames = []string{"mysqld", "mysql", "mysqld-nt"}
			}
			mysqlPid := findProcessByName(ctx, mysqlNames)
			if mysqlPid > 0 {
				// 获取进程信息（CPU、内存等）
				processInfo := getProcessInfo(ctx, "mysql", mysqlPid)
				if processInfo != nil {
					// 合并进程信息和数据库统计信息
					// 进程信息覆盖基础字段，数据库信息提供统计字段
					for k, v := range mysqlDBInfo {
						if _, exists := processInfo[k]; !exists {
							// 如果进程信息中没有这个字段，使用数据库信息
							processInfo[k] = v
						}
					}
					// 确保类型和状态正确
					processInfo["type"] = "local"
					if mysqlDBInfo["status"] == "connected" {
						processInfo["status"] = "connected"
					}
					result["mysql"] = processInfo
				} else {
					// 如果获取进程信息失败，使用数据库信息
					result["mysql"] = mysqlDBInfo
				}
			} else {
				// 找不到进程，使用数据库信息
				result["mysql"] = mysqlDBInfo
			}
		} else {
			// 远程MySQL，直接使用数据库信息
			result["mysql"] = mysqlDBInfo
		}
	}

	// Redis处理：检查是否为本地Redis
	if redisHost == "" {
		// Redis配置为空，尝试查找本地进程
		redisNames := []string{"redis-server", "redis"}
		if runtime.GOOS == "windows" {
			redisNames = []string{"redis-server", "redis-server.exe", "redis"}
		}
		redisPid := findProcessByName(ctx, redisNames)
		if redisPid > 0 {
			redisInfo := getProcessInfo(ctx, "redis", redisPid)
			if redisInfo != nil {
				redisInfo["type"] = "local"
				result["redis"] = redisInfo
			}
		}
		// 如果找不到进程，保持默认的 not_found 状态
	} else if isLocalHost(redisHost) {
		// 本地Redis，尝试查找进程
		redisNames := []string{"redis-server", "redis"}
		if runtime.GOOS == "windows" {
			redisNames = []string{"redis-server", "redis-server.exe", "redis"}
		}
		redisPid := findProcessByName(ctx, redisNames)
		if redisPid > 0 {
			redisInfo := getProcessInfo(ctx, "redis", redisPid)
			if redisInfo != nil {
				redisInfo["type"] = "local"
				result["redis"] = redisInfo
			}
		} else {
			// 本地但找不到进程，尝试通过连接获取信息
			result["redis"] = getRedisInfoFromConnection(ctx)
		}
	} else {
		// 远程Redis，通过连接获取信息
		result["redis"] = getRedisInfoFromConnection(ctx)
	}

	// 获取当前应用进程信息（总是尝试获取，因为这是当前进程）
	currentPid := int32(os.Getpid())
	if currentPid > 0 {
		appInfo := getProcessInfo(ctx, "app", currentPid)
		if appInfo != nil {
			appInfo["type"] = "local"
			// 确保应用进程总是有状态信息
			if appInfo["status"] == "not_found" {
				appInfo["status"] = "running" // 当前进程应该总是运行中
			}
			result["app"] = appInfo
		} else {
			// 如果 getProcessInfo 返回 nil，使用默认值
			result["app"] = map[string]any{
				"name":   "app",
				"pid":    currentPid,
				"cpu":    0.0,
				"memory": 0,
				"status": "running",
				"type":   "local",
			}
		}
	} else {
		// 如果无法获取 PID，至少返回基本信息
		result["app"] = map[string]any{
			"name":   "app",
			"pid":    0,
			"cpu":    0.0,
			"memory": 0,
			"status": "unknown",
			"type":   "local",
		}
	}

	return result
}

// GetSystemInfo 获取系统监控信息
func (r *MonitorController) GetSystemInfo(ctx http.Context) http.Response {
	// CPU信息 - 使用0秒采样，避免阻塞（gopsutil会使用上次采样的差值）
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get CPU percent error", map[string]any{
			"error": err.Error(),
		}, "Get CPU percent error: %v", err)
		cpuPercent = []float64{0}
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get CPU info error", map[string]any{
			"error": err.Error(),
		}, "Get CPU info error: %v", err)
		cpuInfo = []cpu.InfoStat{}
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get memory info error", map[string]any{
			"error": err.Error(),
		}, "Get memory info error: %v", err)
		memInfo = &mem.VirtualMemoryStat{}
	}

	// 磁盘信息（根据操作系统选择路径）
	var diskPath string
	if runtime.GOOS == "windows" {
		// Windows 系统使用当前工作目录的驱动器
		wd, _ := os.Getwd()
		if len(wd) > 0 {
			diskPath = wd[:1] + ":\\"
		} else {
			diskPath = "C:\\"
		}
	} else {
		diskPath = "/"
	}
	diskInfo, err := disk.Usage(diskPath)
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get disk info error", map[string]any{
			"error": err.Error(),
			"path":  diskPath,
		}, "Get disk info error: %v", err)
		diskInfo = &disk.UsageStat{}
	}

	// 网络信息 - 获取所有网卡的详细信息
	netIO, err := net.IOCounters(true) // true 表示获取每个网卡的详细信息
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get network info error", map[string]any{
			"error": err.Error(),
		}, "Get network info error: %v", err)
		netIO = []net.IOCountersStat{}
	}

	// 汇总所有网卡的统计信息
	var totalBytesSent, totalBytesRecv, totalPacketsSent, totalPacketsRecv uint64
	var totalErrin, totalErrout, totalDropin, totalDropout uint64

	// 每个网卡的详细信息
	var interfaces []map[string]any
	for _, io := range netIO {
		// 跳过回环接口（通常以 lo 或 Loopback 开头）
		if io.Name == "lo" || io.Name == "Loopback" || io.Name == "lo0" {
			continue
		}

		totalBytesSent += io.BytesSent
		totalBytesRecv += io.BytesRecv
		totalPacketsSent += io.PacketsSent
		totalPacketsRecv += io.PacketsRecv
		totalErrin += io.Errin
		totalErrout += io.Errout
		totalDropin += io.Dropin
		totalDropout += io.Dropout

		interfaces = append(interfaces, map[string]any{
			"name":         io.Name,
			"bytes_sent":   io.BytesSent,
			"bytes_recv":   io.BytesRecv,
			"packets_sent": io.PacketsSent,
			"packets_recv": io.PacketsRecv,
			"errin":        io.Errin,
			"errout":       io.Errout,
			"dropin":       io.Dropin,
			"dropout":      io.Dropout,
		})
	}

	// 计算网络速度（Mbps）并获取峰值
	sentSpeed, recvSpeed, totalSpeed, peakSent, peakRecv, peakTotal := getNetworkSpeed(totalBytesSent, totalBytesRecv)

	// 汇总统计
	netStats := map[string]any{
		"bytes_sent":       totalBytesSent,
		"bytes_recv":       totalBytesRecv,
		"packets_sent":     totalPacketsSent,
		"packets_recv":     totalPacketsRecv,
		"errin":            totalErrin,
		"errout":           totalErrout,
		"dropin":           totalDropin,
		"dropout":          totalDropout,
		"interfaces":       interfaces, // 所有网卡的详细信息
		"speed_sent_mbps":  sentSpeed,  // 当前发送速度（Mbps）
		"speed_recv_mbps":  recvSpeed,  // 当前接收速度（Mbps）
		"speed_total_mbps": totalSpeed, // 当前总速度（Mbps）
		"peak_sent_mbps":   peakSent,   // 峰值发送速度（Mbps）
		"peak_recv_mbps":   peakRecv,   // 峰值接收速度（Mbps）
		"peak_total_mbps":  peakTotal,  // 峰值总速度（Mbps）
	}

	var cpuModel string
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	// 负载信息（仅Linux/Unix系统）
	var loadAvg map[string]any
	if runtime.GOOS != "windows" {
		avg, err := load.Avg()
		if err != nil {
			errorlog.RecordHTTP(ctx, "monitor", "Get load average error", map[string]any{
				"error": err.Error(),
			}, "Get load average error: %v", err)
			loadAvg = map[string]any{
				"load1":  0.0,
				"load5":  0.0,
				"load15": 0.0,
			}
		} else {
			// 计算负载百分比（相对于CPU核心数）
			cores := float64(len(cpuInfo))
			if cores == 0 {
				cores = 1
			}
			loadPercent1 := (avg.Load1 / cores) * 100
			loadPercent5 := (avg.Load5 / cores) * 100
			loadPercent15 := (avg.Load15 / cores) * 100

			loadAvg = map[string]any{
				"load1":          avg.Load1,
				"load5":          avg.Load5,
				"load15":         avg.Load15,
				"load1_percent":  loadPercent1,
				"load5_percent":  loadPercent5,
				"load15_percent": loadPercent15,
			}
		}
	} else {
		// Windows系统不支持负载
		loadAvg = map[string]any{
			"load1":          0.0,
			"load5":          0.0,
			"load15":         0.0,
			"load1_percent":  0.0,
			"load5_percent":  0.0,
			"load15_percent": 0.0,
		}
	}

	// 文件描述符信息（仅Linux/Unix系统，获取系统全局的）
	var fileDescriptors map[string]any
	if runtime.GOOS != "windows" {
		// 读取系统全局文件描述符信息 /proc/sys/fs/file-nr
		// 格式：已分配 已使用但未释放 最大数量
		used := uint64(0)
		max := uint64(0)

		if data, err := os.ReadFile("/proc/sys/fs/file-nr"); err == nil {
			// 清理数据：去除换行符和空白字符
			dataStr := strings.TrimSpace(string(data))
			// 解析文件内容：例如 "1024 512 65536"
			// 格式：已分配的文件描述符数 已分配但未使用的文件描述符数 系统最大文件描述符数
			var allocated, unused, tempMax uint64
			n, err := fmt.Sscanf(dataStr, "%d %d %d", &allocated, &unused, &tempMax)
			if err == nil && n == 3 {
				// 验证值的合理性：最大文件描述符数不应该超过 10^9 (1 billion)
				if tempMax > 0 && tempMax < 1000000000 {
					max = tempMax
				}
				// 已使用 = 已分配（第一个数字是已分配的文件描述符数，代表系统已使用的）
				if allocated > 0 && allocated < 1000000000 {
					used = allocated
				}
			}
			// 解析失败或值不合理时静默处理，后续会使用默认值
		}
		// 读取失败时静默处理，后续会尝试读取 file-max 或使用默认值

		// 如果无法读取file-nr中的max，尝试单独读取最大限制
		if max == 0 {
			if data, err := os.ReadFile("/proc/sys/fs/file-max"); err == nil {
				// 清理数据：去除换行符和空白字符
				dataStr := strings.TrimSpace(string(data))
				var tempMax uint64
				n, err := fmt.Sscanf(dataStr, "%d", &tempMax)
				if err == nil && n == 1 {
					// 验证值的合理性：最大文件描述符数不应该超过 10^9 (1 billion)
					// 正常的系统值通常在 65536 到几百万之间
					if tempMax > 0 && tempMax < 1000000000 {
						max = tempMax
					}
					// 值异常时静默处理，后续会使用默认值
				}
				// 解析失败或读取失败时静默处理，后续会使用默认值
			}
		}

		// 验证 max 值的合理性，如果异常则重置为0，后续会使用默认值
		if max > 1000000000 {
			max = 0
		}

		// 如果还是无法获取或值异常，使用默认值
		if max == 0 {
			max = 65536 // Linux常见默认值
		}

		// 计算剩余文件描述符，确保不会溢出
		free := uint64(0)
		if max > used {
			free = max - used
		}

		percent := float64(0)
		if max > 0 {
			percent = (float64(used) / float64(max)) * 100
		}

		fileDescriptors = map[string]any{
			"max":     max,
			"used":    used,
			"free":    free,
			"percent": percent,
		}
	} else {
		// Windows系统不支持文件描述符限制
		fileDescriptors = map[string]any{
			"max":     0,
			"used":    0,
			"free":    0,
			"percent": 0.0,
		}
	}

	// 确保 cpuPercent 不为空
	if len(cpuPercent) == 0 {
		cpuPercent = []float64{0}
	}

	// 获取磁盘IO统计（仅Linux/Unix系统）
	var diskIO map[string]any
	if runtime.GOOS != "windows" {
		ioCounters, err := disk.IOCounters()
		if err == nil && len(ioCounters) > 0 {
			// 汇总所有磁盘的IO统计
			var totalReadBytes, totalWriteBytes, totalReadCount, totalWriteCount uint64
			var diskIOCounters []map[string]any
			for name, io := range ioCounters {
				totalReadBytes += io.ReadBytes
				totalWriteBytes += io.WriteBytes
				totalReadCount += io.ReadCount
				totalWriteCount += io.WriteCount
				diskIOCounters = append(diskIOCounters, map[string]any{
					"name":        name,
					"read_bytes":  io.ReadBytes,
					"write_bytes": io.WriteBytes,
					"read_count":  io.ReadCount,
					"write_count": io.WriteCount,
					"read_time":   io.ReadTime,
					"write_time":  io.WriteTime,
				})
			}
			diskIO = map[string]any{
				"total_read_bytes":  totalReadBytes,
				"total_write_bytes": totalWriteBytes,
				"total_read_count":  totalReadCount,
				"total_write_count": totalWriteCount,
				"disks":             diskIOCounters,
			}
		} else {
			diskIO = map[string]any{
				"total_read_bytes":  0,
				"total_write_bytes": 0,
				"total_read_count":  0,
				"total_write_count": 0,
				"disks":             []map[string]any{},
			}
		}
	} else {
		// Windows系统不支持磁盘IO统计
		diskIO = map[string]any{
			"total_read_bytes":  0,
			"total_write_bytes": 0,
			"total_read_count":  0,
			"total_write_count": 0,
			"disks":             []map[string]any{},
		}
	}

	// 获取TCP连接统计（限制处理数量，避免连接数过多时阻塞）
	var tcpConnections map[string]any
	connections, err := net.Connections("tcp")
	if err == nil {
		var established, listen, timeWait, closeWait int
		var listeningPorts []int
		portMap := make(map[int]bool)
		// 限制处理数量，避免连接数过多时阻塞（最多处理10000个连接）
		maxConnections := 10000
		processed := 0
		for _, conn := range connections {
			if processed >= maxConnections {
				break
			}
			processed++
			switch conn.Status {
			case "ESTABLISHED":
				established++
			case "LISTEN":
				listen++
				port := int(conn.Laddr.Port)
				if port > 0 && !portMap[port] {
					listeningPorts = append(listeningPorts, port)
					portMap[port] = true
				}
			case "TIME_WAIT":
				timeWait++
			case "CLOSE_WAIT":
				closeWait++
			}
		}
		tcpConnections = map[string]any{
			"total":           len(connections),
			"established":     established,
			"listen":          listen,
			"time_wait":       timeWait,
			"close_wait":      closeWait,
			"listening_ports": listeningPorts,
		}
	} else {
		tcpConnections = map[string]any{
			"total":           0,
			"established":     0,
			"listen":          0,
			"time_wait":       0,
			"close_wait":      0,
			"listening_ports": []int{},
		}
	}

	// 获取所有磁盘分区信息（限制数量，避免分区过多时阻塞）
	var diskPartitions []map[string]any
	partitions, err := disk.Partitions(false)
	if err == nil {
		maxPartitions := 20 // 限制最多处理20个分区
		count := 0
		for _, part := range partitions {
			if count >= maxPartitions {
				break
			}
			usage, err := disk.Usage(part.Mountpoint)
			if err == nil {
				diskPartitions = append(diskPartitions, map[string]any{
					"device":     part.Device,
					"mountpoint": part.Mountpoint,
					"fstype":     part.Fstype,
					"total":      usage.Total,
					"free":       usage.Free,
					"used":       usage.Used,
					"percent":    usage.UsedPercent,
				})
				count++
			}
		}
	}

	// 生成系统告警提示
	alerts := []map[string]any{}
	if memInfo.UsedPercent > 90 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "high",
			"message": fmt.Sprintf("内存使用率过高: %.2f%%", memInfo.UsedPercent),
			"metric":  "memory",
		})
	} else if memInfo.UsedPercent > 80 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "medium",
			"message": fmt.Sprintf("内存使用率较高: %.2f%%", memInfo.UsedPercent),
			"metric":  "memory",
		})
	}
	if diskInfo.UsedPercent > 90 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "high",
			"message": fmt.Sprintf("磁盘使用率过高: %.2f%%", diskInfo.UsedPercent),
			"metric":  "disk",
		})
	} else if diskInfo.UsedPercent > 80 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "medium",
			"message": fmt.Sprintf("磁盘使用率较高: %.2f%%", diskInfo.UsedPercent),
			"metric":  "disk",
		})
	}
	if cpuPercent[0] > 90 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "high",
			"message": fmt.Sprintf("CPU使用率过高: %.2f%%", cpuPercent[0]),
			"metric":  "cpu",
		})
	} else if cpuPercent[0] > 80 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "medium",
			"message": fmt.Sprintf("CPU使用率较高: %.2f%%", cpuPercent[0]),
			"metric":  "cpu",
		})
	}
	if runtime.GOOS != "windows" {
		if percent, ok := fileDescriptors["percent"].(float64); ok && percent > 90 {
			alerts = append(alerts, map[string]any{
				"type":    "warning",
				"level":   "high",
				"message": fmt.Sprintf("文件描述符使用率过高: %.2f%%", percent),
				"metric":  "file_descriptors",
			})
		}
	}

	return response.Success(ctx, http.Json{
		"os": runtime.GOOS, // 操作系统类型
		"cpu": map[string]any{
			"percent": cpuPercent[0],
			"model":   cpuModel,
			"cores":   len(cpuInfo),
		},
		"memory": map[string]any{
			"total":     memInfo.Total,
			"available": memInfo.Available,
			"used":      memInfo.Used,
			"free":      memInfo.Free,
			"percent":   memInfo.UsedPercent,
			"cached":    memInfo.Cached,
			"buffers":   memInfo.Buffers,
		},
		"disk": map[string]any{
			"total":   diskInfo.Total,
			"free":    diskInfo.Free,
			"used":    diskInfo.Used,
			"percent": diskInfo.UsedPercent,
			"fstype":  diskInfo.Fstype,
			"path":    diskInfo.Path,
		},
		"disk_partitions":  diskPartitions,
		"disk_io":          diskIO,
		"net":              netStats,
		"tcp_connections":  tcpConnections,
		"load":             loadAvg,
		"file_descriptors": fileDescriptors,
		"runtime": map[string]any{
			"goroutines": runtime.NumGoroutine(),
			"total_processes": func() int {
				processes, err := process.Processes()
				if err != nil {
					errorlog.RecordHTTP(ctx, "monitor", "Get processes error", map[string]any{
						"error": err.Error(),
					}, "Get processes error: %v", err)
					return 0
				}
				return len(processes)
			}(),
		},
		"system": map[string]any{
			"hostname": func() string {
				hostname, err := os.Hostname()
				if err != nil {
					errorlog.RecordHTTP(ctx, "monitor", "Get hostname error", map[string]any{
						"error": err.Error(),
					}, "Get hostname error: %v", err)
					return "unknown"
				}
				return hostname
			}(),
			"arch":       runtime.GOARCH,
			"os":         runtime.GOOS,
			"go_version": runtime.Version(),
		},
		"processes": r.getProcessesInfo(ctx),
		"alerts":    alerts,
	})
}

// StreamSystemInfo SSE 实时推送系统监控信息
// 每 2-3 秒推送一次系统监控数据
func (r *MonitorController) StreamSystemInfo(ctx http.Context) http.Response {
	// 设置 SSE 响应头
	writer := ctx.Response().Writer()
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 获取推送间隔（秒），默认 2 秒
	interval := 2
	if intervalStr := ctx.Request().Query("interval", ""); intervalStr != "" {
		if parsed, err := time.ParseDuration(intervalStr + "s"); err == nil {
			interval = int(parsed.Seconds())
			if interval < 1 {
				interval = 1
			}
			if interval > 10 {
				interval = 10
			}
		}
	}

	// 发送初始连接消息
	initMsg := map[string]any{
		"type":     "connected",
		"message":  "SSE连接已建立，开始推送系统监控数据",
		"interval": interval,
	}
	initData, _ := json.Marshal(initMsg)
	fmt.Fprintf(writer, "data: %s\n\n", string(initData))
	if flusher, ok := writer.(nethttp.Flusher); ok {
		flusher.Flush()
	}

	// 创建 ticker，定期推送数据
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// 检测客户端断开连接
	clientGone := ctx.Request().Origin().Context().Done()

	for {
		select {
		case <-clientGone:
			// 客户端断开连接
			return nil
		case <-ticker.C:
			// 检查客户端是否已断开
			select {
			case <-clientGone:
				return nil
			default:
			}

			// 使用 recover 捕获可能的 panic（客户端断开时 writer 可能为 nil）
			func() {
				defer func() {
					if r := recover(); r != nil {
						// 客户端断开连接，静默返回
						facades.Log().Debugf("Monitor SSE: client disconnected, error: %v", r)
					}
				}()

				// 获取系统信息（复用 GetSystemInfo 的逻辑）
				systemInfo := r.collectSystemInfo(ctx)

				// 构造 SSE 消息
				message := map[string]any{
					"type":      "system_info",
					"data":      systemInfo,
					"timestamp": time.Now().Format(time.RFC3339),
				}

				messageData, err := json.Marshal(message)
				if err != nil {
					errorlog.RecordHTTP(ctx, "monitor", "Failed to marshal system info", map[string]any{
						"error": err.Error(),
					}, "Marshal system info error: %v", err)
					return
				}

				// 发送 SSE 消息（可能因客户端断开而失败）
				if _, err := fmt.Fprintf(writer, "data: %s\n\n", string(messageData)); err != nil {
					// 写入失败，客户端可能已断开
					facades.Log().Debugf("Monitor SSE: write failed, client may have disconnected: %v", err)
					return
				}

				// 刷新缓冲区（可能因客户端断开而失败）
				if flusher, ok := writer.(nethttp.Flusher); ok {
					flusher.Flush()
				}
			}()
		}
	}
}

// collectSystemInfo 收集系统监控信息（从 GetSystemInfo 提取的逻辑）
func (r *MonitorController) collectSystemInfo(ctx http.Context) map[string]any {
	// 检查缓存（仅在SSE流中使用缓存，减少系统调用）
	monitorCacheLock.RLock()
	if monitorCache != nil && time.Since(monitorCacheTime) < cacheDuration {
		cached := monitorCache
		monitorCacheLock.RUnlock()
		return cached
	}
	monitorCacheLock.RUnlock()

	// CPU信息 - 使用0秒采样，避免阻塞（gopsutil会使用上次采样的差值）
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get CPU percent error", map[string]any{
			"error": err.Error(),
		}, "Get CPU percent error: %v", err)
		cpuPercent = []float64{0}
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get CPU info error", map[string]any{
			"error": err.Error(),
		}, "Get CPU info error: %v", err)
		cpuInfo = []cpu.InfoStat{}
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get memory info error", map[string]any{
			"error": err.Error(),
		}, "Get memory info error: %v", err)
		memInfo = &mem.VirtualMemoryStat{}
	}

	// 磁盘信息
	var diskPath string
	if runtime.GOOS == "windows" {
		wd, _ := os.Getwd()
		if len(wd) > 0 {
			diskPath = wd[:1] + ":\\"
		} else {
			diskPath = "C:\\"
		}
	} else {
		diskPath = "/"
	}
	diskInfo, err := disk.Usage(diskPath)
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get disk info error", map[string]any{
			"error": err.Error(),
			"path":  diskPath,
		}, "Get disk info error: %v", err)
		diskInfo = &disk.UsageStat{}
	}

	// 网络信息
	netIO, err := net.IOCounters(true)
	if err != nil {
		errorlog.RecordHTTP(ctx, "monitor", "Get network info error", map[string]any{
			"error": err.Error(),
		}, "Get network info error: %v", err)
		netIO = []net.IOCountersStat{}
	}

	// 汇总网络统计
	var totalBytesSent, totalBytesRecv, totalPacketsSent, totalPacketsRecv uint64
	var totalErrin, totalErrout, totalDropin, totalDropout uint64
	var interfaces []map[string]any

	for _, io := range netIO {
		if io.Name == "lo" || io.Name == "Loopback" || io.Name == "lo0" {
			continue
		}
		totalBytesSent += io.BytesSent
		totalBytesRecv += io.BytesRecv
		totalPacketsSent += io.PacketsSent
		totalPacketsRecv += io.PacketsRecv
		totalErrin += io.Errin
		totalErrout += io.Errout
		totalDropin += io.Dropin
		totalDropout += io.Dropout

		interfaces = append(interfaces, map[string]any{
			"name":         io.Name,
			"bytes_sent":   io.BytesSent,
			"bytes_recv":   io.BytesRecv,
			"packets_sent": io.PacketsSent,
			"packets_recv": io.PacketsRecv,
			"errin":        io.Errin,
			"errout":       io.Errout,
			"dropin":       io.Dropin,
			"dropout":      io.Dropout,
		})
	}

	// 计算网络速度（Mbps）并获取峰值
	sentSpeed, recvSpeed, totalSpeed, peakSent, peakRecv, peakTotal := getNetworkSpeed(totalBytesSent, totalBytesRecv)

	netStats := map[string]any{
		"bytes_sent":       totalBytesSent,
		"bytes_recv":       totalBytesRecv,
		"packets_sent":     totalPacketsSent,
		"packets_recv":     totalPacketsRecv,
		"errin":            totalErrin,
		"errout":           totalErrout,
		"dropin":           totalDropin,
		"dropout":          totalDropout,
		"interfaces":       interfaces,
		"speed_sent_mbps":  sentSpeed,  // 当前发送速度（Mbps）
		"speed_recv_mbps":  recvSpeed,  // 当前接收速度（Mbps）
		"speed_total_mbps": totalSpeed, // 当前总速度（Mbps）
		"peak_sent_mbps":   peakSent,   // 峰值发送速度（Mbps）
		"peak_recv_mbps":   peakRecv,   // 峰值接收速度（Mbps）
		"peak_total_mbps":  peakTotal,  // 峰值总速度（Mbps）
	}

	var cpuModel string
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	// 负载信息
	var loadAvg map[string]any
	if runtime.GOOS != "windows" {
		avg, err := load.Avg()
		if err != nil {
			loadAvg = map[string]any{
				"load1":  0.0,
				"load5":  0.0,
				"load15": 0.0,
			}
		} else {
			cores := float64(len(cpuInfo))
			if cores == 0 {
				cores = 1
			}
			loadPercent1 := (avg.Load1 / cores) * 100
			loadPercent5 := (avg.Load5 / cores) * 100
			loadPercent15 := (avg.Load15 / cores) * 100

			loadAvg = map[string]any{
				"load1":          avg.Load1,
				"load5":          avg.Load5,
				"load15":         avg.Load15,
				"load1_percent":  loadPercent1,
				"load5_percent":  loadPercent5,
				"load15_percent": loadPercent15,
			}
		}
	} else {
		loadAvg = map[string]any{
			"load1":          0.0,
			"load5":          0.0,
			"load15":         0.0,
			"load1_percent":  0.0,
			"load5_percent":  0.0,
			"load15_percent": 0.0,
		}
	}

	// 文件描述符信息（简化版，避免重复代码）
	var fileDescriptors map[string]any
	if runtime.GOOS != "windows" {
		used := uint64(0)
		max := uint64(0)

		if data, err := os.ReadFile("/proc/sys/fs/file-nr"); err == nil {
			dataStr := strings.TrimSpace(string(data))
			var allocated, unused, tempMax uint64
			if n, err := fmt.Sscanf(dataStr, "%d %d %d", &allocated, &unused, &tempMax); err == nil && n == 3 {
				// 验证值的合理性
				if tempMax > 0 && tempMax < 1000000000 {
					max = tempMax
				}
				if allocated > 0 && allocated < 1000000000 {
					used = allocated
				}
			}
			// 解析失败或值不合理时静默处理，后续会使用默认值
		}

		if max == 0 {
			if data, err := os.ReadFile("/proc/sys/fs/file-max"); err == nil {
				dataStr := strings.TrimSpace(string(data))
				var tempMax uint64
				if n, err := fmt.Sscanf(dataStr, "%d", &tempMax); err == nil && n == 1 {
					// 验证值的合理性
					if tempMax > 0 && tempMax < 1000000000 {
						max = tempMax
					}
					// 值异常时静默处理，后续会使用默认值
				}
			}
		}

		if max == 0 {
			max = 65536
		}

		free := uint64(0)
		if max > used {
			free = max - used
		}

		percent := float64(0)
		if max > 0 {
			percent = (float64(used) / float64(max)) * 100
		}

		fileDescriptors = map[string]any{
			"max":     max,
			"used":    used,
			"free":    free,
			"percent": percent,
		}
	} else {
		fileDescriptors = map[string]any{
			"max":     0,
			"used":    0,
			"free":    0,
			"percent": 0.0,
		}
	}

	// 确保 cpuPercent 不为空
	if len(cpuPercent) == 0 {
		cpuPercent = []float64{0}
	}

	// 获取磁盘IO统计（仅Linux/Unix系统）
	var diskIO map[string]any
	if runtime.GOOS != "windows" {
		ioCounters, err := disk.IOCounters()
		if err == nil && len(ioCounters) > 0 {
			// 汇总所有磁盘的IO统计
			var totalReadBytes, totalWriteBytes, totalReadCount, totalWriteCount uint64
			var diskIOCounters []map[string]any
			for name, io := range ioCounters {
				totalReadBytes += io.ReadBytes
				totalWriteBytes += io.WriteBytes
				totalReadCount += io.ReadCount
				totalWriteCount += io.WriteCount
				diskIOCounters = append(diskIOCounters, map[string]any{
					"name":        name,
					"read_bytes":  io.ReadBytes,
					"write_bytes": io.WriteBytes,
					"read_count":  io.ReadCount,
					"write_count": io.WriteCount,
					"read_time":   io.ReadTime,
					"write_time":  io.WriteTime,
				})
			}
			diskIO = map[string]any{
				"total_read_bytes":  totalReadBytes,
				"total_write_bytes": totalWriteBytes,
				"total_read_count":  totalReadCount,
				"total_write_count": totalWriteCount,
				"disks":             diskIOCounters,
			}
		} else {
			diskIO = map[string]any{
				"total_read_bytes":  0,
				"total_write_bytes": 0,
				"total_read_count":  0,
				"total_write_count": 0,
				"disks":             []map[string]any{},
			}
		}
	} else {
		// Windows系统不支持磁盘IO统计
		diskIO = map[string]any{
			"total_read_bytes":  0,
			"total_write_bytes": 0,
			"total_read_count":  0,
			"total_write_count": 0,
			"disks":             []map[string]any{},
		}
	}

	// 获取TCP连接统计（限制处理数量，避免连接数过多时阻塞）
	var tcpConnections map[string]any
	connections, err := net.Connections("tcp")
	if err == nil {
		var established, listen, timeWait, closeWait int
		var listeningPorts []int
		portMap := make(map[int]bool)
		// 限制处理数量，避免连接数过多时阻塞（最多处理10000个连接）
		maxConnections := 10000
		processed := 0
		for _, conn := range connections {
			if processed >= maxConnections {
				break
			}
			processed++
			switch conn.Status {
			case "ESTABLISHED":
				established++
			case "LISTEN":
				listen++
				port := int(conn.Laddr.Port)
				if port > 0 && !portMap[port] {
					listeningPorts = append(listeningPorts, port)
					portMap[port] = true
				}
			case "TIME_WAIT":
				timeWait++
			case "CLOSE_WAIT":
				closeWait++
			}
		}
		tcpConnections = map[string]any{
			"total":           len(connections),
			"established":     established,
			"listen":          listen,
			"time_wait":       timeWait,
			"close_wait":      closeWait,
			"listening_ports": listeningPorts,
		}
	} else {
		tcpConnections = map[string]any{
			"total":           0,
			"established":     0,
			"listen":          0,
			"time_wait":       0,
			"close_wait":      0,
			"listening_ports": []int{},
		}
	}

	// 获取所有磁盘分区信息（限制数量，避免分区过多时阻塞）
	var diskPartitions []map[string]any
	partitions, err := disk.Partitions(false)
	if err == nil {
		maxPartitions := 20 // 限制最多处理20个分区
		count := 0
		for _, part := range partitions {
			if count >= maxPartitions {
				break
			}
			usage, err := disk.Usage(part.Mountpoint)
			if err == nil {
				diskPartitions = append(diskPartitions, map[string]any{
					"device":     part.Device,
					"mountpoint": part.Mountpoint,
					"fstype":     part.Fstype,
					"total":      usage.Total,
					"free":       usage.Free,
					"used":       usage.Used,
					"percent":    usage.UsedPercent,
				})
				count++
			}
		}
	}

	// 生成系统告警提示
	alerts := []map[string]any{}
	if memInfo.UsedPercent > 90 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "high",
			"message": fmt.Sprintf("内存使用率过高: %.2f%%", memInfo.UsedPercent),
			"metric":  "memory",
		})
	} else if memInfo.UsedPercent > 80 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "medium",
			"message": fmt.Sprintf("内存使用率较高: %.2f%%", memInfo.UsedPercent),
			"metric":  "memory",
		})
	}
	if diskInfo.UsedPercent > 90 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "high",
			"message": fmt.Sprintf("磁盘使用率过高: %.2f%%", diskInfo.UsedPercent),
			"metric":  "disk",
		})
	} else if diskInfo.UsedPercent > 80 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "medium",
			"message": fmt.Sprintf("磁盘使用率较高: %.2f%%", diskInfo.UsedPercent),
			"metric":  "disk",
		})
	}
	if cpuPercent[0] > 90 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "high",
			"message": fmt.Sprintf("CPU使用率过高: %.2f%%", cpuPercent[0]),
			"metric":  "cpu",
		})
	} else if cpuPercent[0] > 80 {
		alerts = append(alerts, map[string]any{
			"type":    "warning",
			"level":   "medium",
			"message": fmt.Sprintf("CPU使用率较高: %.2f%%", cpuPercent[0]),
			"metric":  "cpu",
		})
	}
	if runtime.GOOS != "windows" {
		if percent, ok := fileDescriptors["percent"].(float64); ok && percent > 90 {
			alerts = append(alerts, map[string]any{
				"type":    "warning",
				"level":   "high",
				"message": fmt.Sprintf("文件描述符使用率过高: %.2f%%", percent),
				"metric":  "file_descriptors",
			})
		}
	}

	result := map[string]any{
		"os": runtime.GOOS,
		"cpu": map[string]any{
			"percent": cpuPercent[0],
			"model":   cpuModel,
			"cores":   len(cpuInfo),
		},
		"memory": map[string]any{
			"total":     memInfo.Total,
			"available": memInfo.Available,
			"used":      memInfo.Used,
			"free":      memInfo.Free,
			"percent":   memInfo.UsedPercent,
			"cached":    memInfo.Cached,
			"buffers":   memInfo.Buffers,
		},
		"disk": map[string]any{
			"total":   diskInfo.Total,
			"free":    diskInfo.Free,
			"used":    diskInfo.Used,
			"percent": diskInfo.UsedPercent,
			"fstype":  diskInfo.Fstype,
			"path":    diskInfo.Path,
		},
		"disk_partitions":  diskPartitions,
		"disk_io":          diskIO,
		"net":              netStats,
		"tcp_connections":  tcpConnections,
		"load":             loadAvg,
		"file_descriptors": fileDescriptors,
		"runtime": func() map[string]any {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			return map[string]any{
				"goroutines": runtime.NumGoroutine(),
				"num_cpu":    runtime.NumCPU(),
				"gomaxprocs": runtime.GOMAXPROCS(0), // 0表示获取当前值，不修改
				"total_processes": func() int {
					processes, err := process.Processes()
					if err != nil {
						return 0
					}
					return len(processes)
				}(),
				"memory": map[string]any{
					"alloc":          memStats.Alloc,        // 当前分配的内存
					"total_alloc":    memStats.TotalAlloc,   // 累计分配的内存
					"sys":            memStats.Sys,          // 系统内存
					"lookups":        memStats.Lookups,      // 指针查找次数
					"mallocs":        memStats.Mallocs,      // 分配次数
					"frees":          memStats.Frees,        // 释放次数
					"heap_alloc":     memStats.HeapAlloc,    // 堆内存分配
					"heap_sys":       memStats.HeapSys,      // 堆内存系统
					"heap_idle":      memStats.HeapIdle,     // 堆内存空闲
					"heap_inuse":     memStats.HeapInuse,    // 堆内存使用
					"heap_objects":   memStats.HeapObjects,  // 堆对象数
					"stack_inuse":    memStats.StackInuse,   // 栈内存使用
					"stack_sys":      memStats.StackSys,     // 栈内存系统
					"num_gc":         memStats.NumGC,        // GC次数
					"pause_total_ns": memStats.PauseTotalNs, // GC总暂停时间（纳秒）
					"last_gc":        memStats.LastGC,       // 上次GC时间
				},
			}
		}(),
		"system": map[string]any{
			"hostname": func() string {
				hostname, err := os.Hostname()
				if err != nil {
					return "unknown"
				}
				return hostname
			}(),
			"arch":       runtime.GOARCH,
			"os":         runtime.GOOS,
			"go_version": runtime.Version(),
		},
		"processes": r.getProcessesInfo(ctx),
		"alerts":    alerts,
	}

	// 更新缓存（仅在collectSystemInfo中缓存，用于SSE流）
	monitorCacheLock.Lock()
	monitorCache = result
	monitorCacheTime = time.Now()
	monitorCacheLock.Unlock()

	return result
}
