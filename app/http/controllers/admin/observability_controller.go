package admin

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

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

func NewObservabilityController() *ObservabilityController {
	return &ObservabilityController{
		slowQueryService: services.NewSlowQueryService(),
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
	if err := facades.Orm().Query().
		Model(&models.OperationLog{}).
		Where("trace_id", traceID).
		With("Admin").
		Order("id asc").
		Limit(50).
		Get(&operations); err != nil {
		return response.ErrorWithLog(ctx, "observability", err, map[string]any{"trace_id": traceID})
	}

	var systemLogs []models.SystemLog
	if err := facades.Orm().Query().
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

	events, err := r.collectAuditEvents(traceID, keyword, adminID, startTime, endTime)
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

func (r *ObservabilityController) collectAuditEvents(traceID, keyword string, adminID int, startTime, endTime string) ([]auditEvent, error) {
	opQuery := facades.Orm().Query().Model(&models.OperationLog{}).With("Admin").Order("id desc").Limit(500)
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

	sysQuery := facades.Orm().Query().Model(&models.SystemLog{}).Order("id desc").Limit(500)
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
