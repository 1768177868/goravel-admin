package admin

import (
	"bytes"
	"encoding/json"
	appfacades "goravel/app/facades"
	"net/http"
	rpprof "runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	pprofprofile "github.com/google/pprof/profile"
	ghttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type ObservabilityController struct {
	slowQueryService services.SlowQueryService
	systemLogService services.SystemLogService
	apiMetricService services.ApiMetricService
}

type auditEvent struct {
	Time       string `json:"time"`
	SortAt     int64  `json:"-"`
	Type       string `json:"type"`
	TraceID    string `json:"trace_id"`
	Level      string `json:"level,omitempty"`
	Module     string `json:"module,omitempty"`
	Title      string `json:"title,omitempty"`
	Method     string `json:"method,omitempty"`
	Path       string `json:"path,omitempty"`
	AdminID    uint   `json:"admin_id,omitempty"`
	AdminName  string `json:"admin_name,omitempty"`
	Status     uint8  `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
	Context    any    `json:"context,omitempty"`
	DurationMS int    `json:"duration_ms,omitempty"`
}

type pprofHotspot struct {
	Function    string  `json:"function"`
	FlatValue   int64   `json:"flat_value"`
	FlatPercent float64 `json:"flat_percent"`
	CumValue    int64   `json:"cum_value"`
	CumPercent  float64 `json:"cum_percent"`
}

func NewObservabilityController() *ObservabilityController {
	return &ObservabilityController{
		slowQueryService: services.NewSlowQueryService(),
		systemLogService: services.NewSystemLogService(),
		apiMetricService: services.NewApiMetricService(),
	}
}

// TraceAggregate 请求追踪聚合（按 trace_id 串联请求、异常、SQL）
func (r *ObservabilityController) TraceAggregate(ctx ghttp.Context) ghttp.Response {
	traceID := strings.TrimSpace(ctx.Request().Query("trace_id", ""))
	if traceID == "" {
		return response.Error(ctx, http.StatusBadRequest, "trace_id_required")
	}

	_ = r.slowQueryService.CollectFromLatestLog(200)

	var operations []models.OperationLog
	if err := appfacades.OrmQuery(ctx).
		Model(&models.OperationLog{}).
		Where("trace_id", traceID).
		With("Admin").
		Order("id asc").
		Limit(50).
		Get(&operations); err != nil {
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{"trace_id": traceID})
	}

	var systemLogs []models.SystemLog
	if err := appfacades.OrmQuery(ctx).
		Model(&models.SystemLog{}).
		Where("trace_id", traceID).
		Order("id asc").
		Limit(200).
		Get(&systemLogs); err != nil {
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{"trace_id": traceID})
	}

	slowSQL, _ := r.slowQueryService.GetByTraceID(traceID, 100)

	return response.Success(ctx, ghttp.Json{
		"trace_id":    traceID,
		"request":     firstOperationRequest(operations),
		"operations":  operations,
		"exceptions":  filterExceptionLogs(systemLogs),
		"system_logs": systemLogs,
		"slow_sql":    slowSQL,
	})
}

// SlowSQLTopN 慢 SQL TopN 聚合
func (r *ObservabilityController) SlowSQLTopN(ctx ghttp.Context) ghttp.Response {
	minDurationMS := float64(helpers.GetIntQuery(ctx, "min_duration_ms", 200))
	hours := helpers.GetIntQuery(ctx, "hours", 24)
	limit := helpers.GetIntQuery(ctx, "limit", 20)

	_ = r.slowQueryService.CollectFromLatestLog(minDurationMS)
	top, err := r.slowQueryService.GetTopN(hours, limit, minDurationMS)
	if err != nil {
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"hours":           hours,
			"limit":           limit,
			"min_duration_ms": minDurationMS,
		})
	}

	return response.Success(ctx, ghttp.Json{
		"hours":           hours,
		"limit":           limit,
		"min_duration_ms": minDurationMS,
		"list":            top,
	})
}

// APIPerformanceOverview 接口性能总览（慢接口、错误率、QPS、P95/P99）
func (r *ObservabilityController) APIPerformanceOverview(ctx ghttp.Context) ghttp.Response {
	hours := helpers.GetIntQuery(ctx, "hours", 24)
	limit := helpers.GetIntQuery(ctx, "limit", 20)
	overview, err := r.apiMetricService.GetOverview(hours, limit)
	if err != nil {
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"hours": hours,
			"limit": limit,
		})
	}
	return response.Success(ctx, overview)
}

// APIPerformanceTraces 接口性能下钻（查看某接口最近 trace）
func (r *ObservabilityController) APIPerformanceTraces(ctx ghttp.Context) ghttp.Response {
	method := strings.TrimSpace(ctx.Request().Query("method", ""))
	routeTemplate := strings.TrimSpace(ctx.Request().Query("route_template", ""))
	hours := helpers.GetIntQuery(ctx, "hours", 24)
	limit := helpers.GetIntQuery(ctx, "limit", 20)
	list, err := r.apiMetricService.GetRecentTraces(method, routeTemplate, hours, limit)
	if err != nil {
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"method":         method,
			"route_template": routeTemplate,
			"hours":          hours,
			"limit":          limit,
		})
	}
	return response.Success(ctx, ghttp.Json{
		"method":         method,
		"route_template": routeTemplate,
		"hours":          hours,
		"limit":          limit,
		"list":           list,
	})
}

// AuditTimeline 审计事件聚合时间线（操作日志 + 系统日志）
func (r *ObservabilityController) AuditTimeline(ctx ghttp.Context) ghttp.Response {
	traceID := strings.TrimSpace(ctx.Request().Query("trace_id", ""))
	keyword := strings.TrimSpace(ctx.Request().Query("keyword", ""))
	adminID := helpers.GetIntQuery(ctx, "admin_id", 0)
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 20)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	startTime := getTimeQueryUTC(ctx, "start_time")
	endTime := getTimeQueryUTC(ctx, "end_time")

	events, err := r.collectAuditEvents(ctx, traceID, keyword, adminID, startTime, endTime)
	if err != nil {
		return response.ErrorWithLog(ctx, "observability", err)
	}

	total := len(events)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return response.Success(ctx, ghttp.Json{
		"list":      events[start:end],
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// QueueDashboard 轻量队列看板（database / Redis 列表 / Redis Stream 三类可统计，其余标记不支持）
func (r *ObservabilityController) QueueDashboard(ctx ghttp.Context) ghttp.Response {
	reader := services.NewQueueStatsReader()
	panels, defaultConn := reader.BuildQueueDashboard()
	return response.Success(ctx, ghttp.Json{
		"default_connection": defaultConn,
		"connections":        panels,
	})
}

// PprofStatus 返回 pprof 功能开关与 token 要求
func (r *ObservabilityController) PprofStatus(ctx ghttp.Context) ghttp.Response {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	var adminID uint
	if admin, ok := adminValue.(models.Admin); ok {
		adminID = admin.ID
	} else if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
		adminID = adminPtr.ID
	}

	isDeveloper := false
	if adminID > 0 {
		developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
		parts := strings.Split(developerIDsStr, ",")
		for _, part := range parts {
			idStr := strings.TrimSpace(part)
			if idStr == "" {
				continue
			}
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err == nil && uint(id) == adminID {
				isDeveloper = true
				break
			}
		}
	}

	pprofEnabled := facades.Config().GetBool("pprof.enabled", false) || facades.Config().GetBool("app.debug", false)
	tokenRequired := strings.TrimSpace(facades.Config().GetString("pprof.token", "")) != ""

	return response.Success(ctx, ghttp.Json{
		"enabled":        pprofEnabled,
		"is_developer":   isDeveloper,
		"token_required": tokenRequired,
	})
}

// PprofVerify 验证 pprof token 是否可用（仅开发者）
func (r *ObservabilityController) PprofVerify(ctx ghttp.Context) ghttp.Response {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	var adminID uint
	if admin, ok := adminValue.(models.Admin); ok {
		adminID = admin.ID
	} else if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
		adminID = adminPtr.ID
	}

	if adminID == 0 {
		return response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	isDeveloper := false
	for _, part := range strings.Split(developerIDsStr, ",") {
		idStr := strings.TrimSpace(part)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err == nil && uint(id) == adminID {
			isDeveloper = true
			break
		}
	}
	if !isDeveloper {
		return response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	pprofEnabled := facades.Config().GetBool("pprof.enabled", false) || facades.Config().GetBool("app.debug", false)
	if !pprofEnabled {
		return response.Error(ctx, http.StatusBadRequest, "pprof_disabled")
	}

	needToken := strings.TrimSpace(facades.Config().GetString("pprof.token", "")) != ""
	configuredToken := strings.TrimSpace(facades.Config().GetString("pprof.token", ""))
	token := strings.TrimSpace(ctx.Request().Input("token"))
	if needToken && token == "" {
		return response.Error(ctx, http.StatusBadRequest, "pprof_token_required")
	}
	if needToken && token != configuredToken {
		return response.Error(ctx, http.StatusBadRequest, "pprof_token_invalid")
	}

	return response.Success(ctx, ghttp.Json{
		"verified": true,
	})
}

// PprofCPUHotspots CPU 热点函数 TopN（按自身 CPU 时间排序）
func (r *ObservabilityController) PprofCPUHotspots(ctx ghttp.Context) ghttp.Response {
	adminID, resp := r.ensurePprofAccess(ctx, true)
	if resp != nil {
		return resp
	}

	seconds := parsePositiveIntInput(ctx, "seconds", 10)
	if seconds <= 0 {
		seconds = 10
	}
	if seconds > 120 {
		seconds = 120
	}

	topN := parsePositiveIntInput(ctx, "top_n", 15)
	if topN <= 0 {
		topN = 15
	}
	if topN > 100 {
		topN = 100
	}
	r.recordPprofSamplingLog(ctx, "pprof_cpu_sampling_started", map[string]any{
		"admin_id": adminID,
		"seconds":  seconds,
		"top_n":    topN,
	})

	var buf bytes.Buffer
	if err := rpprof.StartCPUProfile(&buf); err != nil {
		r.recordPprofSamplingLog(ctx, "pprof_cpu_sampling_failed", map[string]any{
			"admin_id": adminID,
			"seconds":  seconds,
			"top_n":    topN,
			"error":    err.Error(),
		})
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"seconds": seconds,
			"top_n":   topN,
		})
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	rpprof.StopCPUProfile()

	prof, err := pprofprofile.ParseData(buf.Bytes())
	if err != nil {
		r.recordPprofSamplingLog(ctx, "pprof_cpu_sampling_failed", map[string]any{
			"admin_id": adminID,
			"seconds":  seconds,
			"top_n":    topN,
			"error":    err.Error(),
		})
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"seconds": seconds,
			"top_n":   topN,
		})
	}

	valueIdx := selectProfileValueIndex(prof, "nanoseconds")
	list, total := buildHotspotsFromProfile(prof, valueIdx, topN)
	r.recordPprofSamplingLog(ctx, "pprof_cpu_sampling_completed", map[string]any{
		"admin_id":    adminID,
		"seconds":     seconds,
		"top_n":       topN,
		"total_value": total,
		"result_size": len(list),
	})

	return response.Success(ctx, ghttp.Json{
		"seconds":     seconds,
		"top_n":       topN,
		"total_value": total,
		"unit":        "nanoseconds",
		"list":        list,
	})
}

// PprofMemoryHotspots 内存分配热点 TopN（按分配量排序）
func (r *ObservabilityController) PprofMemoryHotspots(ctx ghttp.Context) ghttp.Response {
	adminID, resp := r.ensurePprofAccess(ctx, true)
	if resp != nil {
		return resp
	}

	topN := parsePositiveIntInput(ctx, "top_n", 25)
	if topN <= 0 {
		topN = 25
	}
	if topN > 100 {
		topN = 100
	}
	r.recordPprofSamplingLog(ctx, "pprof_memory_sampling_started", map[string]any{
		"admin_id": adminID,
		"top_n":    topN,
	})

	var buf bytes.Buffer
	profiler := rpprof.Lookup("allocs")
	if profiler == nil {
		r.recordPprofSamplingLog(ctx, "pprof_memory_sampling_failed", map[string]any{
			"admin_id": adminID,
			"top_n":    topN,
			"error":    "allocs_profiler_not_available",
		})
		return response.Error(ctx, http.StatusInternalServerError, "operation_failed")
	}
	if err := profiler.WriteTo(&buf, 0); err != nil {
		r.recordPprofSamplingLog(ctx, "pprof_memory_sampling_failed", map[string]any{
			"admin_id": adminID,
			"top_n":    topN,
			"error":    err.Error(),
		})
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"top_n": topN,
		})
	}

	prof, err := pprofprofile.ParseData(buf.Bytes())
	if err != nil {
		r.recordPprofSamplingLog(ctx, "pprof_memory_sampling_failed", map[string]any{
			"admin_id": adminID,
			"top_n":    topN,
			"error":    err.Error(),
		})
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{
			"top_n": topN,
		})
	}

	valueIdx := selectProfileValueIndex(prof, "bytes")
	list, total := buildHotspotsFromProfile(prof, valueIdx, topN)
	r.recordPprofSamplingLog(ctx, "pprof_memory_sampling_completed", map[string]any{
		"admin_id":    adminID,
		"top_n":       topN,
		"total_value": total,
		"result_size": len(list),
	})

	return response.Success(ctx, ghttp.Json{
		"top_n":       topN,
		"total_value": total,
		"unit":        "bytes",
		"list":        list,
	})
}

func (r *ObservabilityController) ensurePprofAccess(ctx ghttp.Context, checkToken bool) (uint, ghttp.Response) {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return 0, response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	var adminID uint
	if admin, ok := adminValue.(models.Admin); ok {
		adminID = admin.ID
	} else if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
		adminID = adminPtr.ID
	}
	if adminID == 0 {
		return 0, response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	isDeveloper := false
	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	for _, part := range strings.Split(developerIDsStr, ",") {
		idStr := strings.TrimSpace(part)
		if idStr == "" {
			continue
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err == nil && uint(id) == adminID {
			isDeveloper = true
			break
		}
	}
	if !isDeveloper {
		return 0, response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	pprofEnabled := facades.Config().GetBool("pprof.enabled", false) || facades.Config().GetBool("app.debug", false)
	if !pprofEnabled {
		return 0, response.Error(ctx, http.StatusBadRequest, "pprof_disabled")
	}
	if checkToken {
		needToken := strings.TrimSpace(facades.Config().GetString("pprof.token", "")) != ""
		configuredToken := strings.TrimSpace(facades.Config().GetString("pprof.token", ""))
		token := strings.TrimSpace(ctx.Request().Input("token"))
		if needToken && token == "" {
			return 0, response.Error(ctx, http.StatusBadRequest, "pprof_token_required")
		}
		if needToken && token != configuredToken {
			return 0, response.Error(ctx, http.StatusBadRequest, "pprof_token_invalid")
		}
	}

	return adminID, nil
}

func selectProfileValueIndex(prof *pprofprofile.Profile, preferredUnit string) int {
	if prof == nil || len(prof.SampleType) == 0 {
		return 0
	}
	for i, sampleType := range prof.SampleType {
		if strings.EqualFold(sampleType.Unit, preferredUnit) {
			return i
		}
	}
	return len(prof.SampleType) - 1
}

func buildHotspotsFromProfile(prof *pprofprofile.Profile, valueIdx, topN int) ([]pprofHotspot, int64) {
	if prof == nil {
		return []pprofHotspot{}, 0
	}
	if valueIdx < 0 {
		valueIdx = 0
	}
	flat := map[string]int64{}
	cum := map[string]int64{}
	var total int64

	for _, sample := range prof.Sample {
		if valueIdx >= len(sample.Value) {
			continue
		}
		value := sample.Value[valueIdx]
		if value <= 0 {
			continue
		}
		total += value

		cumSeen := map[string]bool{}
		flatAdded := false
		for _, location := range sample.Location {
			for _, line := range location.Line {
				if line.Function == nil {
					continue
				}
				name := line.Function.Name
				if name == "" {
					continue
				}
				if !cumSeen[name] {
					cum[name] += value
					cumSeen[name] = true
				}
				if !flatAdded {
					flat[name] += value
					flatAdded = true
				}
			}
		}
	}

	items := make([]pprofHotspot, 0, len(flat))
	for fn, flatValue := range flat {
		cumValue := cum[fn]
		item := pprofHotspot{
			Function:    fn,
			FlatValue:   flatValue,
			CumValue:    cumValue,
			FlatPercent: 0,
			CumPercent:  0,
		}
		if total > 0 {
			item.FlatPercent = float64(flatValue) * 100 / float64(total)
			item.CumPercent = float64(cumValue) * 100 / float64(total)
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].FlatValue == items[j].FlatValue {
			return items[i].CumValue > items[j].CumValue
		}
		return items[i].FlatValue > items[j].FlatValue
	})
	if topN > 0 && len(items) > topN {
		items = items[:topN]
	}

	return items, total
}

func parsePositiveIntInput(ctx ghttp.Context, key string, defaultValue int) int {
	raw := strings.TrimSpace(ctx.Request().Input(key))
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func (r *ObservabilityController) recordPprofSamplingLog(ctx ghttp.Context, message string, attrs map[string]any) {
	if r.systemLogService == nil {
		return
	}
	_ = r.systemLogService.RecordHTTP(ctx, "info", "pprof", message, attrs)
}

func (r *ObservabilityController) collectAuditEvents(ctx ghttp.Context, traceID, keyword string, adminID int, startTime, endTime string) ([]auditEvent, error) {
	opQuery := appfacades.OrmQuery(ctx).Model(&models.OperationLog{}).With("Admin").Order("id desc").Limit(500)
	if traceID != "" {
		opQuery = opQuery.Where("trace_id = ?", traceID)
	}
	if adminID > 0 {
		opQuery = opQuery.Where("admin_id = ?", adminID)
	}
	if keyword != "" {
		opQuery = opQuery.Where("path LIKE ? OR title LIKE ? OR request LIKE ? OR error_msg LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if startTime != "" {
		opQuery = opQuery.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		opQuery = opQuery.Where("created_at <= ?", endTime)
	}
	var opLogs []models.OperationLog
	if err := opQuery.Get(&opLogs); err != nil {
		return nil, err
	}

	sysQuery := appfacades.OrmQuery(ctx).Model(&models.SystemLog{}).Order("id desc").Limit(500)
	if traceID != "" {
		sysQuery = sysQuery.Where("trace_id = ?", traceID)
	}
	if keyword != "" {
		sysQuery = sysQuery.Where("message LIKE ? OR module LIKE ? OR context LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if startTime != "" {
		sysQuery = sysQuery.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		sysQuery = sysQuery.Where("created_at <= ?", endTime)
	}
	var sysLogs []models.SystemLog
	if err := sysQuery.Get(&sysLogs); err != nil {
		return nil, err
	}

	events := make([]auditEvent, 0, len(opLogs)+len(sysLogs))
	for _, item := range opLogs {
		events = append(events, auditEvent{
			Time:       formatCarbon(item.CreatedAt),
			SortAt:     toUnix(item.CreatedAt),
			Type:       "operation",
			TraceID:    item.TraceID,
			Title:      item.Title,
			Method:     item.Method,
			Path:       item.Path,
			AdminID:    item.AdminID,
			AdminName:  item.Admin.Username,
			Status:     item.Status,
			Message:    item.ErrorMsg,
			DurationMS: item.Duration,
			Context: map[string]any{
				"request": item.Request,
				"changes": safeJSON(item.Changes),
			},
		})
	}
	for _, item := range sysLogs {
		events = append(events, auditEvent{
			Time:    formatCarbon(item.CreatedAt),
			SortAt:  toUnix(item.CreatedAt),
			Type:    "system",
			TraceID: item.TraceID,
			Level:   item.Level,
			Module:  item.Module,
			Message: item.Message,
			Context: safeJSON(item.Context),
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].SortAt > events[j].SortAt
	})
	return events, nil
}

func firstOperationRequest(operations []models.OperationLog) map[string]any {
	if len(operations) == 0 {
		return map[string]any{}
	}
	first := operations[0]
	return map[string]any{
		"trace_id":    first.TraceID,
		"path":        first.Path,
		"method":      first.Method,
		"status":      first.Status,
		"request":     safeJSON(first.Request),
		"duration_ms": first.Duration,
		"created_at":  first.CreatedAt,
	}
}

func formatCarbon(dt *carbon.DateTime) string {
	if dt == nil {
		return ""
	}
	return dt.ToDateTimeString()
}

func toUnix(dt *carbon.DateTime) int64 {
	if dt == nil {
		return 0
	}
	parsed, err := time.Parse("2006-01-02 15:04:05", dt.ToDateTimeString())
	if err != nil {
		return 0
	}
	return parsed.Unix()
}

func filterExceptionLogs(logs []models.SystemLog) []models.SystemLog {
	items := make([]models.SystemLog, 0)
	for _, item := range logs {
		level := strings.ToLower(item.Level)
		if level == "error" || strings.Contains(strings.ToLower(item.Message), "panic") || strings.Contains(strings.ToLower(item.Message), "exception") {
			items = append(items, item)
		}
	}
	return items
}

func safeJSON(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
		return decoded
	}
	return raw
}
