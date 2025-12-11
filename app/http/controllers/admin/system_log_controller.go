package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/constants"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
)

type SystemLogController struct {
}

func NewSystemLogController() *SystemLogController {
	return &SystemLogController{}
}

// findSystemLogByID 根据ID查找系统日志，如果不存在则返回错误响应
func (r *SystemLogController) findSystemLogByID(ctx http.Context, id uint) (*models.SystemLog, http.Response) {
	if id == 0 {
		return nil, response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var log models.SystemLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	if log.ID == 0 {
		return nil, response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	return &log, nil
}

// Index 获取系统日志列表
func (r *SystemLogController) Index(ctx http.Context) http.Response {
	// 验证并规范化分页参数
	page, pageSize := helpers.ValidatePagination(
		helpers.GetIntQuery(ctx, "page", 1),
		helpers.GetIntQuery(ctx, "page_size", 10),
	)
	level := ctx.Request().Query("level", "")
	module := ctx.Request().Query("module", "")
	traceID := ctx.Request().Query("trace_id", "")
	message := ctx.Request().Query("message", "")
	orderBy := ctx.Request().Query("order_by", "")
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
		return response.ErrorWithLog(ctx, "system-log", err)
	}

	var logs []models.SystemLog
	offset := (page - 1) * pageSize
	query = helpers.ApplySort(query, orderBy, "id:desc")
	if err = query.Offset(offset).Limit(pageSize).Get(&logs); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err)
	}

	return response.Paginate(ctx, logs, total, page, pageSize)
}

// Show 获取系统日志详情
func (r *SystemLogController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findSystemLogByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, "get_success", http.Json{
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

	if _, err := facades.Orm().Query().Delete(log); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err, map[string]any{
			"log_id": log.ID,
		})
	}

	return response.Success(ctx, "delete_success")
}

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

	// 使用工具函数转换为 []any
	idsAny := helpers.ConvertUintSliceToAny(ids)

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.SystemLog{}); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err, map[string]any{
			"ids": ids,
		})
	}

	return response.Success(ctx, "delete_success")
}

// Clean 清理系统日志
// 删除指定天数之前的日志，默认删除30天前的日志
func (r *SystemLogController) Clean(ctx http.Context) http.Response {
	days := helpers.GetIntQuery(ctx, "days", constants.DefaultCleanLogDays)
	if days <= 0 {
		days = constants.DefaultCleanLogDays
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := facades.Orm().Query().Model(&models.SystemLog{}).Where("created_at < ?", cutoffTime).Delete(&models.SystemLog{}); err != nil {
		return response.ErrorWithLog(ctx, "system-log", err, map[string]any{
			"days": days,
		})
	}

	return response.Success(ctx, "clean_success")
}
