package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/constants"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type LoginLogController struct {
}

func NewLoginLogController() *LoginLogController {
	return &LoginLogController{}
}

// Index 获取登录日志列表
func (r *LoginLogController) Index(ctx http.Context) http.Response {
	// 验证并规范化分页参数
	page, pageSize := helpers.ValidatePagination(
		helpers.GetIntQuery(ctx, "page", 1),
		helpers.GetIntQuery(ctx, "page_size", 10),
	)
	adminID := ctx.Request().Query("admin_id", "")
	username := ctx.Request().Query("username", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	orderBy := ctx.Request().Query("order_by", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	query := facades.Orm().Query().Model(&models.LoginLog{})

	if adminID != "" {
		query = query.Where("admin_id", adminID)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	total, err := query.Count()
	if err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Failed to count login logs", map[string]any{
			"error": err.Error(),
		}, "Count login logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var logs []models.LoginLog
	offset := (page - 1) * pageSize
	query = helpers.ApplySort(query, orderBy, "id:desc")
	if err = query.With("Admin").Offset(offset).Limit(pageSize).Get(&logs); err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Failed to query login logs", map[string]any{
			"error": err.Error(),
		}, "Query login logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", logs, total, page, pageSize)
}

// Show 获取登录日志详情
func (r *LoginLogController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var log models.LoginLog
	if err := facades.Orm().Query().With("Admin").Where("id", id).First(&log); err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Login log not found", map[string]any{
			"error":  err.Error(),
			"log_id": id,
		}, "Login log not found: %v", err)
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"log": log,
	})
}

// Destroy 删除登录日志
func (r *LoginLogController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var log models.LoginLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Login log not found for delete", map[string]any{
			"error":  err.Error(),
			"log_id": id,
		}, "Login log not found: %v", err)
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&log); err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Failed to delete login log", map[string]any{
			"error":  err.Error(),
			"log_id": log.ID,
		}, "Delete login log error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

type LoginLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除登录日志
func (r *LoginLogController) BatchDestroy(ctx http.Context) http.Response {
	var req LoginLogBatchDestroyRequest

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

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.LoginLog{}); err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Failed to batch delete login logs", map[string]any{
			"error": err.Error(),
			"ids":   ids,
		}, "Batch delete login logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// Clean 清理登录日志
// 删除指定天数之前的日志，默认删除30天前的日志
func (r *LoginLogController) Clean(ctx http.Context) http.Response {
	days := helpers.GetIntQuery(ctx, "days", constants.DefaultCleanLogDays)
	if days <= 0 {
		days = constants.DefaultCleanLogDays
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := facades.Orm().Query().Model(&models.LoginLog{}).Where("created_at < ?", cutoffTime).Delete(&models.LoginLog{}); err != nil {
		errorlog.RecordHTTP(ctx, "login-log", "Failed to clean login logs", map[string]any{
			"error": err.Error(),
			"days":  days,
		}, "Clean login logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "clean_failed")
	}

	return response.Success(ctx, "clean_success")
}
