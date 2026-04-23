package admin

import (
	"runtime"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/response"
	wsnotifications "goravel/app/websocket/notifications"
)

type ObservabilityLiteController struct{}

func NewObservabilityLiteController() *ObservabilityLiteController {
	return &ObservabilityLiteController{}
}

func (r *ObservabilityLiteController) Summary(ctx http.Context) http.Response {
	wsAdmins, wsConnections := wsnotifications.Hub().Stats()
	monitorController := NewMonitorController()
	systemSnapshot := monitorController.collectSystemInfo(ctx)
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	alertCount := countAlerts(systemSnapshot)
	healthStatus := "ok"
	if alertCount > 0 {
		healthStatus = "warning"
	}

	return response.Success(ctx, http.Json{
		"mode":      "detailed",
		"timestamp": time.Now().Format(time.RFC3339),
		"app": http.Json{
			"env":              facades.Config().GetString("app.env", "production"),
			"debug":            facades.Config().GetBool("app.debug", false),
			"timezone":         facades.Config().GetString("app.timezone", "UTC"),
			"queue_connection": facades.Config().GetString("queue.default", "sync"),
			"cache_store":      facades.Config().GetString("cache.default", "file"),
		},
		"runtime": http.Json{
			"goroutines": runtime.NumGoroutine(),
			"go_version": runtime.Version(),
			"num_cpu":    runtime.NumCPU(),
			"memory": http.Json{
				"alloc":        memStats.Alloc,
				"total_alloc":  memStats.TotalAlloc,
				"sys":          memStats.Sys,
				"heap_alloc":   memStats.HeapAlloc,
				"heap_inuse":   memStats.HeapInuse,
				"heap_idle":    memStats.HeapIdle,
				"heap_objects": memStats.HeapObjects,
				"num_gc":       memStats.NumGC,
			},
		},
		"websocket": http.Json{
			"online_admins": wsAdmins,
			"connections":   wsConnections,
		},
		"health": http.Json{
			"status":      healthStatus,
			"alert_count": alertCount,
		},
		"system_snapshot": systemSnapshot,
		// compatibility fields
		"app_env":          facades.Config().GetString("app.env", "production"),
		"app_debug":        facades.Config().GetBool("app.debug", false),
		"queue_connection": facades.Config().GetString("queue.default", "sync"),
		"cache_store":      facades.Config().GetString("cache.default", "file"),
		"ws": http.Json{
			"online_admins": wsAdmins,
			"connections":   wsConnections,
		},
	})
}

func countAlerts(systemSnapshot map[string]any) int {
	alertsValue, ok := systemSnapshot["alerts"]
	if !ok {
		return 0
	}

	switch alerts := alertsValue.(type) {
	case []any:
		return len(alerts)
	case []map[string]any:
		return len(alerts)
	default:
		return 0
	}
}
