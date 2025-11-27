package admin

import (
	"os"
	"runtime"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"goravel/app/http/response"
	"goravel/app/utils/logger"
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
		logger.ErrorfHTTP(ctx, "Get CPU info error: %v", err)
		cpuPercent = []float64{0}
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		logger.ErrorfHTTP(ctx, "Get CPU info error: %v", err)
		cpuInfo = []cpu.InfoStat{}
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		logger.ErrorfHTTP(ctx, "Get memory info error: %v", err)
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
		logger.ErrorfHTTP(ctx, "Get disk info error: %v", err)
		diskInfo = &disk.UsageStat{}
	}

	// 网络信息
	netIO, err := net.IOCounters(false)
	if err != nil {
		logger.ErrorfHTTP(ctx, "Get network info error: %v", err)
		netIO = []net.IOCountersStat{}
	}

	var netStats map[string]interface{}
	if len(netIO) > 0 {
		netStats = map[string]interface{}{
			"bytes_sent":   netIO[0].BytesSent,
			"bytes_recv":   netIO[0].BytesRecv,
			"packets_sent": netIO[0].PacketsSent,
			"packets_recv": netIO[0].PacketsRecv,
			"errin":        netIO[0].Errin,
			"errout":       netIO[0].Errout,
			"dropin":       netIO[0].Dropin,
			"dropout":      netIO[0].Dropout,
		}
	} else {
		netStats = map[string]interface{}{
			"bytes_sent":   0,
			"bytes_recv":   0,
			"packets_sent": 0,
			"packets_recv": 0,
			"errin":        0,
			"errout":       0,
			"dropin":       0,
			"dropout":      0,
		}
	}

	var cpuModel string
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return response.Success(ctx, "get_success", http.Json{
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
		"net": netStats,
	})
}
