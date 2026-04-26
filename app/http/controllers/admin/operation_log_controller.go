package admin

import (
	"sort"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/samber/lo"

	"goravel/app/constants"
	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type OperationLogController struct {
	operationLogService services.OperationLogService
}

func NewOperationLogController() *OperationLogController {
	return &OperationLogController{
		operationLogService: services.NewOperationLogService(),
	}
}

// findOperationLogByID 根据ID查找操作日志，如果不存在则返回错误响应
// withAdmin 为 true 时会预加载 Admin 关联
func (r *OperationLogController) findOperationLogByID(ctx http.Context, id uint, withAdmin bool) (*models.OperationLog, http.Response) {
	log, err := r.operationLogService.GetByID(id, withAdmin)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrLogNotFound.Code)
	}
	return log, nil
}

// buildFilters 构建查询过滤器
func (r *OperationLogController) buildFilters(ctx http.Context) services.OperationLogFilters {
	adminID := ctx.Request().Query("admin_id", "")
	traceID := ctx.Request().Query("trace_id", "")
	username := ctx.Request().Query("username", "")
	method := ctx.Request().Query("method", "")
	path := ctx.Request().Query("path", "")
	title := ctx.Request().Query("title", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	request := ctx.Request().Query("request", "")
	startTimeStr := getTimeQueryUTC(ctx, "start_time")
	endTimeStr := getTimeQueryUTC(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.OperationLogFilters{
		AdminID:   adminID,
		TraceID:   traceID,
		Username:  username,
		Method:    method,
		Path:      path,
		Title:     title,
		IP:        ip,
		Status:    status,
		Request:   request,
		StartTime: startTimeStr,
		EndTime:   endTimeStr,
		OrderBy:   orderBy,
	}
}

// Index 获取操作日志列表
func (r *OperationLogController) Index(ctx http.Context) http.Response {
	// 验证时间范围（操作日志查询限制为3个月，可通过配置修改）
	startTimeStr := getTimeQueryUTC(ctx, "start_time")
	endTimeStr := getTimeQueryUTC(ctx, "end_time")

	// 如果只填了开始时间，结束时间默认为当前时间
	if startTimeStr != "" {
		startTime, resp := parseOptionalTimeFromQuery(ctx, "start_time", "invalid_start_time")
		if resp != nil {
			return resp
		}

		endTime := time.Now().UTC()
		if endTimeStr != "" {
			parsedEndTime, endResp := parseOptionalTimeFromQuery(ctx, "end_time", "invalid_end_time")
			if endResp != nil {
				return endResp
			}
			endTime = parsedEndTime
		}

		if resp := validateTimeRangeResponse(ctx, startTime, endTime, 3); resp != nil {
			return resp
		}
	}

	filters := r.buildFilters(ctx)

	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	logs, total, err := r.operationLogService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "operation-log", err)
	}

	return response.Success(ctx, http.Json{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 获取操作日志详情
func (r *OperationLogController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findOperationLogByID(ctx, id, true) // 预加载 Admin 关联
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"log": *log,
	})
}

// Destroy 删除操作日志
func (r *OperationLogController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findOperationLogByID(ctx, id, false)
	if resp != nil {
		return resp
	}

	if _, err := facades.Orm().Query().Delete(log); err != nil {
		return response.ErrorWithLog(ctx, "operation-log", err, map[string]any{
			"log_id": log.ID,
		})
	}

	return response.Success(ctx)
}

type OperationLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除操作日志
func (r *OperationLogController) BatchDestroy(ctx http.Context) http.Response {
	var req OperationLogBatchDestroyRequest

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

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.OperationLog{}); err != nil {
		return response.ErrorWithLog(ctx, "operation-log", err, map[string]any{
			"ids": ids,
		})
	}

	return response.Success(ctx)
}

// Clean 清理操作日志
// 删除指定天数之前的日志，默认删除30天前的日志
func (r *OperationLogController) Clean(ctx http.Context) http.Response {
	days := helpers.GetIntQuery(ctx, "days", constants.DefaultCleanLogDays)
	if days <= 0 {
		days = constants.DefaultCleanLogDays
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := facades.Orm().Query().Model(&models.OperationLog{}).Where("created_at < ?", cutoffTime).Delete(&models.OperationLog{}); err != nil {
		return response.ErrorWithLog(ctx, "operation-log", err, map[string]any{
			"days": days,
		})
	}

	return response.Success(ctx)
}

// GetTitleOptions 获取所有可用的操作标题选项
func (r *OperationLogController) GetTitleOptions(ctx http.Context) http.Response {
	// 从数据库查询已存在的标题（现在标题直接存权限标识 slug，如 admin.update）
	var dbTitles []string
	_ = facades.Orm().Query().Model(&models.OperationLog{}).
		Select("DISTINCT title").
		Where("title IS NOT NULL AND title != ''"). // 排除空标题
		Order("title ASC").
		Pluck("title", &dbTitles)

	// 同时读取权限表中的启用 slug，避免前端依赖手写默认值。
	var permissionSlugs []string
	_ = facades.Orm().Query().Model(&models.Permission{}).
		Select("slug").
		Where("status = 1").
		Order("slug ASC").
		Pluck("slug", &permissionSlugs)

	// 过滤并去重标题（权限标识），忽略旧的 operation.xxx 配置
	result := lo.Uniq(lo.Filter(append(dbTitles, permissionSlugs...), func(title string, _ int) bool {
		// 排除空标题、未知标题以及旧的 operation.xxx 标题
		return title != "" && title != "operation.unknown" && !strings.HasPrefix(title, "operation.")
	}))

	sort.Strings(result)

	return response.Success(ctx, http.Json{
		"titles": result,
	})
}
