package services

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/utils/traceid"
)

type SystemLogService interface {
	// GetByID 根据ID获取系统日志
	GetByID(id uint) (*models.SystemLog, error)
	// GetList 获取系统日志列表
	GetList(filters SystemLogFilters, page, pageSize int) ([]models.SystemLog, int64, error)
	// RecordHTTP 记录系统日志（HTTP context）
	RecordHTTP(ctx http.Context, level, module, message string, attributes map[string]any) error
	// Record 记录系统日志（标准 context）
	Record(ctx context.Context, level, module, message string, attributes map[string]any) error
}

// SystemLogFilters 系统日志查询过滤器
type SystemLogFilters struct {
	Level     string
	Module    string
	TraceID   string
	Message   string
	StartTime string
	EndTime   string
	OrderBy   string
}

type SystemLogServiceImpl struct {
}

var (
	systemLogsTraceIDColumnOnce sync.Once
	systemLogsHasTraceIDColumn  bool
)

func hasSystemLogsTraceIDColumn() bool {
	systemLogsTraceIDColumnOnce.Do(func() {
		systemLogsHasTraceIDColumn = facades.Schema().HasTable("system_logs") && facades.Schema().HasColumn("system_logs", "trace_id")
	})
	return systemLogsHasTraceIDColumn
}

func NewSystemLogService() SystemLogService {
	return &SystemLogServiceImpl{}
}

// GetByID 根据ID获取系统日志
func (s *SystemLogServiceImpl) GetByID(id uint) (*models.SystemLog, error) {
	var log models.SystemLog
	if err := facades.Orm().Query().Where("id", id).FirstOrFail(&log); err != nil {
		return nil, apperrors.ErrLogNotFound.WithError(err)
	}
	return &log, nil
}

// GetList 获取系统日志列表
func (s *SystemLogServiceImpl) GetList(filters SystemLogFilters, page, pageSize int) ([]models.SystemLog, int64, error) {
	query := facades.Orm().Query().Model(&models.SystemLog{})

	// 应用筛选条件
	if filters.Level != "" {
		query = query.Where("level = ?", filters.Level)
	}
	if filters.Module != "" {
		query = query.Where("module LIKE ?", "%"+filters.Module+"%")
	}
	if filters.TraceID != "" {
		if hasSystemLogsTraceIDColumn() {
			query = query.Where("trace_id LIKE ?", "%"+filters.TraceID+"%")
		}
	}
	if filters.Message != "" {
		query = query.Where("message LIKE ?", "%"+filters.Message+"%")
	}
	if filters.StartTime != "" {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if filters.EndTime != "" {
		query = query.Where("created_at <= ?", filters.EndTime)
	}

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "id:desc"
	}
	query = helpers.ApplySort(query, orderBy, "id:desc")

	// 分页查询
	var logs []models.SystemLog
	var total int64
	if err := query.Paginate(page, pageSize, &logs, &total); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// RecordHTTP 记录系统日志（HTTP context）
func (s *SystemLogServiceImpl) RecordHTTP(ctx http.Context, level, module, message string, attributes map[string]any) error {
	// 检查数据库连接是否可用
	orm := facades.Orm()
	if orm == nil {
		// 数据库未初始化时，静默跳过日志记录
		return nil
	}

	var contextJSON string
	if len(attributes) > 0 {
		if data, err := json.Marshal(attributes); err == nil {
			contextJSON = string(data)
		}
	}

	traceID := traceid.FromHTTPContext(ctx)
	if traceID == "" {
		traceID = traceid.EnsureHTTPContext(ctx, "")
	}

	payload := map[string]any{
		"level":      level,
		"module":     module,
		"message":    message,
		"context":    contextJSON,
		"ip":         ctx.Request().Ip(),
		"user_agent": ctx.Request().Header("User-Agent", ""),
	}
	if hasSystemLogsTraceIDColumn() {
		payload["trace_id"] = traceID
	}

	if err := orm.Query().Table("system_logs").Create(payload); err != nil {
		return err
	}
	return nil
}

// Record 记录系统日志（标准 context）
func (s *SystemLogServiceImpl) Record(ctx context.Context, level, module, message string, attributes map[string]any) error {
	// 检查数据库连接是否可用
	orm := facades.Orm()
	if orm == nil {
		// 数据库未初始化时，静默跳过日志记录
		return nil
	}

	var contextJSON string
	if len(attributes) > 0 {
		if data, err := json.Marshal(attributes); err == nil {
			contextJSON = string(data)
		}
	}

	traceID := traceid.FromContext(ctx)
	if traceID == "" {
		var newCtx context.Context
		newCtx, traceID = traceid.EnsureContext(ctx)
		ctx = newCtx
	}

	payload := map[string]any{
		"level":   level,
		"module":  module,
		"message": message,
		"context": contextJSON,
	}
	if hasSystemLogsTraceIDColumn() {
		payload["trace_id"] = traceID
	}

	if err := orm.Query().Table("system_logs").Create(payload); err != nil {
		return err
	}
	return nil
}
