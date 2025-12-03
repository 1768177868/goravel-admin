package admin

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
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
	var interfaces []map[string]interface{}
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

		interfaces = append(interfaces, map[string]interface{}{
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
	netStats := map[string]interface{}{
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
	var loadAvg map[string]interface{}
	if runtime.GOOS != "windows" {
		avg, err := load.Avg()
		if err != nil {
			errorlog.RecordHTTP(ctx, "monitor", "Get load average error", map[string]any{
				"error": err.Error(),
			}, "Get load average error: %v", err)
			loadAvg = map[string]interface{}{
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

			loadAvg = map[string]interface{}{
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
		loadAvg = map[string]interface{}{
			"load1":          0.0,
			"load5":          0.0,
			"load15":         0.0,
			"load1_percent":  0.0,
			"load5_percent":  0.0,
			"load15_percent": 0.0,
		}
	}

	// 文件描述符信息（仅Linux/Unix系统，获取系统全局的）
	var fileDescriptors map[string]interface{}
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
				// 9223372036854775807 是 uint64 的最大值，说明解析失败
				if tempMax > 0 && tempMax < 1000000000 {
					max = tempMax
				} else {
					// 值异常，记录警告但不中断流程
					errorlog.RecordHTTP(ctx, "monitor", "Invalid max value from file-nr", map[string]any{
						"value": tempMax,
						"data":  dataStr,
					}, "Invalid max value from file-nr: %d, data: %s", tempMax, dataStr)
				}
				// 已使用 = 已分配（第一个数字是已分配的文件描述符数，代表系统已使用的）
				if allocated > 0 && allocated < 1000000000 {
					used = allocated
				} else if allocated >= 1000000000 {
					// 值异常，记录警告但不中断流程
					errorlog.RecordHTTP(ctx, "monitor", "Invalid used value from file-nr", map[string]any{
						"value": allocated,
						"data":  dataStr,
					}, "Invalid used value from file-nr: %d, data: %s", allocated, dataStr)
				}
			} else {
				errorlog.RecordHTTP(ctx, "monitor", "Parse file-nr error", map[string]any{
					"error": err.Error(),
					"data":  dataStr,
					"n":     n,
				}, "Parse file-nr error: %v, data: %s, n: %d", err, dataStr, n)
			}
		} else {
			errorlog.RecordHTTP(ctx, "monitor", "Read file-nr error", map[string]any{
				"error": err.Error(),
			}, "Read /proc/sys/fs/file-nr error: %v", err)
		}

		// 如果无法读取file-nr中的max，尝试单独读取最大限制
		if max == 0 {
			if data, err := os.ReadFile("/proc/sys/fs/file-max"); err == nil {
				// 清理数据：去除换行符和空白字符
				dataStr := strings.TrimSpace(string(data))
				var tempMax uint64
				n, err := fmt.Sscanf(dataStr, "%d", &tempMax)
				if err == nil && n == 1 {
					// 验证值的合理性：最大文件描述符数不应该超过 10^9 (1 billion)
					// 9223372036854775807 是 uint64 的最大值，说明解析失败
					// 正常的系统值通常在 65536 到几百万之间
					if tempMax > 0 && tempMax < 1000000000 {
						max = tempMax
					} else {
						// 值异常，记录警告但不中断流程
						errorlog.RecordHTTP(ctx, "monitor", "Invalid file-max value", map[string]any{
							"value": tempMax,
							"data":  dataStr,
						}, "Invalid file-max value: %d, data: %s", tempMax, dataStr)
					}
				} else {
					errorlog.RecordHTTP(ctx, "monitor", "Parse file-max error", map[string]any{
						"error": err.Error(),
						"data":  dataStr,
						"n":     n,
					}, "Parse file-max error: %v, data: %s, n: %d", err, dataStr, n)
				}
			} else {
				errorlog.RecordHTTP(ctx, "monitor", "Read file-max error", map[string]any{
					"error": err.Error(),
				}, "Read /proc/sys/fs/file-max error: %v", err)
			}
		}

		// 验证 max 值的合理性
		if max > 1000000000 {
			errorlog.RecordHTTP(ctx, "monitor", "File descriptor max value too large, using default", map[string]any{
				"max": max,
			}, "File descriptor max value too large: %d, using default", max)
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

		fileDescriptors = map[string]interface{}{
			"max":     max,
			"used":    used,
			"free":    free,
			"percent": percent,
		}
	} else {
		// Windows系统不支持文件描述符限制
		fileDescriptors = map[string]interface{}{
			"max":     0,
			"used":    0,
			"free":    0,
			"percent": 0.0,
		}
	}

	return response.Success(ctx, "get_success", http.Json{
		"os": runtime.GOOS, // 操作系统类型
		"cpu": map[string]interface{}{
			"percent": cpuPercent[0],
			"model":   cpuModel,
			"cores":   len(cpuInfo),
		},
		"memory": map[string]interface{}{
			"total":     memInfo.Total,
			"available": memInfo.Available,
			"used":      memInfo.Used,
			"free":      memInfo.Free,
			"percent":   memInfo.UsedPercent,
			"cached":    memInfo.Cached,
			"buffers":   memInfo.Buffers,
		},
		"disk": map[string]interface{}{
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
		"runtime": map[string]interface{}{
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
		"system": map[string]interface{}{
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
	initMsg := map[string]interface{}{
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
			message := map[string]interface{}{
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
func (r *MonitorController) collectSystemInfo(ctx http.Context) map[string]interface{} {
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
	var interfaces []map[string]interface{}

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

		interfaces = append(interfaces, map[string]interface{}{
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

	netStats := map[string]interface{}{
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
	var loadAvg map[string]interface{}
	if runtime.GOOS != "windows" {
		avg, err := load.Avg()
		if err != nil {
			loadAvg = map[string]interface{}{
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

			loadAvg = map[string]interface{}{
				"load1":          avg.Load1,
				"load5":          avg.Load5,
				"load15":         avg.Load15,
				"load1_percent":  loadPercent1,
				"load5_percent":  loadPercent5,
				"load15_percent": loadPercent15,
			}
		}
	} else {
		loadAvg = map[string]interface{}{
			"load1":          0.0,
			"load5":          0.0,
			"load15":         0.0,
			"load1_percent":  0.0,
			"load5_percent":  0.0,
			"load15_percent": 0.0,
		}
	}

	// 文件描述符信息（简化版，避免重复代码）
	var fileDescriptors map[string]interface{}
	if runtime.GOOS != "windows" {
		used := uint64(0)
		max := uint64(0)

		if data, err := os.ReadFile("/proc/sys/fs/file-nr"); err == nil {
			dataStr := strings.TrimSpace(string(data))
			var allocated, unused, tempMax uint64
			if n, err := fmt.Sscanf(dataStr, "%d %d %d", &allocated, &unused, &tempMax); err == nil && n == 3 {
				if tempMax > 0 && tempMax < 1000000000 {
					max = tempMax
				}
				if allocated > 0 && allocated < 1000000000 {
					used = allocated
				}
			}
		}

		if max == 0 {
			if data, err := os.ReadFile("/proc/sys/fs/file-max"); err == nil {
				dataStr := strings.TrimSpace(string(data))
				var tempMax uint64
				if n, err := fmt.Sscanf(dataStr, "%d", &tempMax); err == nil && n == 1 {
					if tempMax > 0 && tempMax < 1000000000 {
						max = tempMax
					}
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

		fileDescriptors = map[string]interface{}{
			"max":     max,
			"used":    used,
			"free":    free,
			"percent": percent,
		}
	} else {
		fileDescriptors = map[string]interface{}{
			"max":     0,
			"used":    0,
			"free":    0,
			"percent": 0.0,
		}
	}

	return map[string]interface{}{
		"os": runtime.GOOS,
		"cpu": map[string]interface{}{
			"percent": cpuPercent[0],
			"model":   cpuModel,
			"cores":   len(cpuInfo),
		},
		"memory": map[string]interface{}{
			"total":     memInfo.Total,
			"available": memInfo.Available,
			"used":      memInfo.Used,
			"free":      memInfo.Free,
			"percent":   memInfo.UsedPercent,
			"cached":    memInfo.Cached,
			"buffers":   memInfo.Buffers,
		},
		"disk": map[string]interface{}{
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
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"total_processes": func() int {
				processes, err := process.Processes()
				if err != nil {
					return 0
				}
				return len(processes)
			}(),
		},
		"system": map[string]interface{}{
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
	}
}
