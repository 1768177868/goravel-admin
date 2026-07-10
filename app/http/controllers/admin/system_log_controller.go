package admin

import (
	"context"
	appfacades "goravel/app/facades"
	"sort"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/constants"
	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type SystemLogController struct {
	systemLogService services.SystemLogService
}

func NewSystemLogController() *SystemLogController {
	return &SystemLogController{
		systemLogService: services.NewSystemLogService(context.Background()),
	}
}

// findSystemLogByID 根据ID查找系统日志，如果不存在则返回错误响应
func (r *SystemLogController) findSystemLogByID(ctx http.Context, id uint) (*models.SystemLog, http.Response) {
	log, err := r.systemLogService.GetByID(id)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrLogNotFound.Code)
	}
	return log, nil
}

// buildFilters 构建查询过滤器
func (r *SystemLogController) buildFilters(ctx http.Context) services.SystemLogFilters {
	level := ctx.Request().Query("level", "")
	module := ctx.Request().Query("module", "")
	traceID := ctx.Request().Query("trace_id", "")
	message := ctx.Request().Query("message", "")
	startTime := getTimeQueryUTC(ctx, "start_time")
	endTime := getTimeQueryUTC(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.SystemLogFilters{
		Level:     level,
		Module:    module,
		TraceID:   traceID,
		Message:   message,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Index 获取系统日志列表
func (r *SystemLogController) Index(ctx http.Context) http.Response {
	filters := r.buildFilters(ctx)

	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	logs, total, err := r.systemLogService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "system-log", err)
	}

	return response.Success(ctx, http.Json{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 获取系统日志详情
func (r *SystemLogController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findSystemLogByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"log": *log,
	})
}

// Destroy 删除系统日志
func (r *SystemLogController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findSystemLogByID(ctx, id)
	if resp != nil {
		return resp
	}

	if _, err := appfacades.OrmQuery(ctx).Delete(log); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err, map[string]any{
			"log_id": log.ID,
		})
	}

	return response.Success(ctx)
}

type SystemLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除系统日志
func (r *SystemLogController) BatchDestroy(ctx http.Context) http.Response {
	var req SystemLogBatchDestroyRequest

	// 使用结构体绑定
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsError.Code)
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDsRequired.Code)
	}

	ids := req.IDs

	// 使用工具函数转换为 []any
	idsAny := helpers.ConvertUintSliceToAny(ids)

	if _, err := appfacades.OrmQuery(ctx).WhereIn("id", idsAny).Delete(&models.SystemLog{}); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err, map[string]any{
			"ids": ids,
		})
	}

	return response.Success(ctx)
}

// Clean 清理系统日志
// 删除指定天数之前的日志，默认删除30天前的日志
func (r *SystemLogController) Clean(ctx http.Context) http.Response {
	days := helpers.GetIntQuery(ctx, "days", constants.DefaultCleanLogDays)
	if days <= 0 {
		days = constants.DefaultCleanLogDays
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := appfacades.OrmQuery(ctx).Model(&models.SystemLog{}).Where("created_at < ?", cutoffTime).Delete(&models.SystemLog{}); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err, map[string]any{
			"days": days,
		})
	}

	return response.Success(ctx)
}

// GetModuleOptions 获取系统日志模块选项（用于前端筛选下拉）
func (r *SystemLogController) GetModuleOptions(ctx http.Context) http.Response {
	var modules []string
	_ = appfacades.OrmQuery(ctx).Model(&models.SystemLog{}).
		Select("DISTINCT module").
		Where("module IS NOT NULL AND module != ''").
		Order("module ASC").
		Pluck("module", &modules)

	// 去重并排序，避免数据库方言差异导致顺序不稳定
	moduleSet := make(map[string]struct{}, len(modules))
	uniqueModules := make([]string, 0, len(modules))
	for _, module := range modules {
		if module == "" {
			continue
		}
		if _, exists := moduleSet[module]; exists {
			continue
		}
		moduleSet[module] = struct{}{}
		uniqueModules = append(uniqueModules, module)
	}
	sort.Strings(uniqueModules)

	return response.Success(ctx, http.Json{
		"modules": uniqueModules,
	})
}
