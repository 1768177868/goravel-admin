package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type SystemLogController struct {
}

func NewSystemLogController() *SystemLogController {
	return &SystemLogController{}
}

// Index 系统日志列表
func (r *SystemLogController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	level := ctx.Request().Query("level", "")
	module := ctx.Request().Query("module", "")
	traceID := ctx.Request().Query("trace_id", "")
	message := ctx.Request().Query("message", "")
	orderBy := ctx.Request().Query("order_by", "")
	// 使用辅助函数自动转换时区
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	query := facades.Orm().Query().Model(&models.SystemLog{})

	if level != "" {
		query = query.Where("level = ?", level)
	}
	if module != "" {
		query = query.Where("module LIKE ?", "%"+module+"%")
	}
	if traceID != "" {
		query = query.Where("trace_id LIKE ?", "%"+traceID+"%")
	}
	if message != "" {
		query = query.Where("message LIKE ?", "%"+message+"%")
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var logs []models.SystemLog
	offset := (page - 1) * pageSize
	// 应用排序（默认按id倒序）
	query = helpers.ApplySort(query, orderBy, "id:desc")
	if err = query.Offset(offset).Limit(pageSize).Get(&logs); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", logs, total, page, pageSize)
}

// Show 系统日志详情
func (r *SystemLogController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var log models.SystemLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"log": log,
	})
}

// Destroy 删除系统日志
func (r *SystemLogController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var log models.SystemLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&log); err != nil {
		errorlog.RecordHTTP(ctx, "system-log", "Failed to delete system log", map[string]any{
			"error": err.Error(),
			"log_id": log.ID,
		}, "Delete system log error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// SystemLogBatchDestroyRequest 批量删除请求结构
type SystemLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除系统日志
func (r *SystemLogController) BatchDestroy(ctx http.Context) http.Response {
	var req SystemLogBatchDestroyRequest

	// 使用结构体绑定
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "params_error")
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "ids_required")
	}

	ids := req.IDs

	// 转换为 []any
	idsAny := make([]any, len(ids))
	for i, id := range ids {
		idsAny[i] = id
	}

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.SystemLog{}); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// Clean 清理系统日志
func (r *SystemLogController) Clean(ctx http.Context) http.Response {
	days := cast.ToInt(ctx.Request().Input("days", "30"))
	if days <= 0 {
		days = 30
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := facades.Orm().Query().Model(&models.SystemLog{}).Where("created_at < ?", cutoffTime).Delete(&models.SystemLog{}); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "clean_failed")
	}

	return response.Success(ctx, "clean_success")
}
