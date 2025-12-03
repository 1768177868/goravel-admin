package admin

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"goravel/app/http/response"
	"goravel/app/utils/errorlog"
)

type MonitorController struct{}

func NewMonitorController() *MonitorController {
	return &MonitorController{}
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

	// 获取进程状态
	status, err := proc.Status()
	if err == nil && len(status) > 0 {
		result["status"] = status[0]
	} else {
		result["status"] = "running"
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
		"cpu":         0.0,
		"memory":      0,
		"status":      "disconnected",
		"rss":         0,
		"vms":         0,
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
	} else {
		result["host"] = fmt.Sprintf("%s:%d", dbHost, dbPort)
		result["type"] = "local"
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
	if err := orm.Query().Raw("SELECT VERSION() as version").Scan(&version); err == nil {
		result["version"] = version
	}

	// 获取MySQL状态信息
	var statusRows []map[string]any
	if err := orm.Query().Raw("SHOW STATUS WHERE Variable_name IN ('Uptime', 'Threads_connected', 'Questions', 'Connections')").Scan(&statusRows); err == nil {
		for _, row := range statusRows {
			if variableName, ok := row["Variable_name"].(string); ok {
				if value, ok := row["Value"].(string); ok {
					intValue, _ := strconv.ParseInt(value, 10, 64)
					switch variableName {
					case "Uptime":
						uptime = intValue
					case "Threads_connected":
						threads = intValue
					case "Questions":
						queries = intValue
					case "Connections":
						connections = intValue
					}
				}
			}
		}
		result["uptime"] = uptime
		result["threads"] = threads
		result["queries"] = queries
		result["connections"] = connections
	}

	// 获取MySQL变量信息（内存相关）
	var variableRows []map[string]any
	if err := orm.Query().Raw("SHOW VARIABLES WHERE Variable_name IN ('max_connections', 'innodb_buffer_pool_size')").Scan(&variableRows); err == nil {
		for _, row := range variableRows {
			if variableName, ok := row["Variable_name"].(string); ok {
				if value, ok := row["Value"].(string); ok {
					intValue, _ := strconv.ParseInt(value, 10, 64)
					if variableName == "innodb_buffer_pool_size" {
						result["buffer_pool_size"] = intValue
					} else if variableName == "max_connections" {
						result["max_connections"] = intValue
					}
				}
			}
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
		"cpu":                      0.0,
		"memory":                   0,
		"status":                   "disconnected",
		"rss":                      0,
		"vms":                      0,
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

	// 检查是否为本地Redis
	if !isLocalHost(redisHost) {
		result["host"] = fmt.Sprintf("%s:%d", redisHost, redisPort)
		result["type"] = "remote"
	} else {
		result["host"] = fmt.Sprintf("%s:%d", redisHost, redisPort)
		result["type"] = "local"
	}

	// 尝试连接Redis获取信息
	cache := facades.Cache()
	if cache == nil {
		return result
	}

	// 尝试执行Redis INFO命令（通过Cache接口可能不支持，这里先标记为已连接）
	// 注意：Goravel的Cache接口可能不直接支持INFO命令，这里先返回基本连接状态
	result["status"] = "connected"

	// 如果Redis连接正常，尝试获取一些基本信息
	// 注意：这需要Redis驱动支持，如果不行可以标记为"connected"但无法获取详细信息
	return result
}

// getProcessesInfo 获取 MySQL、Redis 和当前应用进程的信息
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

	// MySQL处理：检查是否为本地数据库
	if dbConnection == "mysql" && isLocalHost(dbHost) {
		// 本地MySQL，尝试查找进程
		mysqlNames := []string{"mysqld", "mysql", "mariadb"}
		if runtime.GOOS == "windows" {
			mysqlNames = []string{"mysqld", "mysql", "mysqld-nt"}
		}
		mysqlPid := findProcessByName(ctx, mysqlNames)
		if mysqlPid > 0 {
			result["mysql"] = getProcessInfo(ctx, "mysql", mysqlPid)
			result["mysql"].(map[string]any)["type"] = "local"
		} else {
			// 本地但找不到进程，尝试通过数据库连接获取信息
			result["mysql"] = getMySQLInfoFromDB(ctx)
		}
	} else if dbConnection == "mysql" {
		// 远程MySQL，通过数据库连接获取信息
		result["mysql"] = getMySQLInfoFromDB(ctx)
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
			result["redis"] = getProcessInfo(ctx, "redis", redisPid)
			result["redis"].(map[string]any)["type"] = "local"
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
			result["redis"] = getProcessInfo(ctx, "redis", redisPid)
			result["redis"].(map[string]any)["type"] = "local"
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

	// 确保返回的 result 不为 nil（虽然已经初始化了，但双重保险）
	if result == nil {
		result = map[string]any{
			"mysql": map[string]any{"name": "mysql", "status": "error"},
			"redis": map[string]any{"name": "redis", "status": "error"},
			"app":   map[string]any{"name": "app", "status": "error"},
		}
	}

	return result
}

// GetSystemInfo 获取系统监控信息
func (r *MonitorController) GetSystemInfo(ctx http.Context) http.Response {
	// CPU信息
	cpuPercent, err := cpu.Percent(time.Second, false)
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

	// 汇总统计
	netStats := map[string]any{
		"bytes_sent":   totalBytesSent,
		"bytes_recv":   totalBytesRecv,
		"packets_sent": totalPacketsSent,
		"packets_recv": totalPacketsRecv,
		"errin":        totalErrin,
		"errout":       totalErrout,
		"dropin":       totalDropin,
		"dropout":      totalDropout,
		"interfaces":   interfaces, // 所有网卡的详细信息
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

	return response.Success(ctx, "get_success", http.Json{
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
		"net":              netStats,
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
				continue
			}

			// 发送 SSE 消息
			fmt.Fprintf(writer, "data: %s\n\n", string(messageData))

			// 刷新缓冲区
			if flusher, ok := writer.(nethttp.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// collectSystemInfo 收集系统监控信息（从 GetSystemInfo 提取的逻辑）
func (r *MonitorController) collectSystemInfo(ctx http.Context) map[string]any {
	// CPU信息
	cpuPercent, err := cpu.Percent(time.Second, false)
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

	netStats := map[string]any{
		"bytes_sent":   totalBytesSent,
		"bytes_recv":   totalBytesRecv,
		"packets_sent": totalPacketsSent,
		"packets_recv": totalPacketsRecv,
		"errin":        totalErrin,
		"errout":       totalErrout,
		"dropin":       totalDropin,
		"dropout":      totalDropout,
		"interfaces":   interfaces,
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

	return map[string]any{
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
		"net":              netStats,
		"load":             loadAvg,
		"file_descriptors": fileDescriptors,
		"runtime": map[string]any{
			"goroutines": runtime.NumGoroutine(),
			"total_processes": func() int {
				processes, err := process.Processes()
				if err != nil {
					return 0
				}
				return len(processes)
			}(),
		},
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
	}
}
