package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
)

type LoginLogController struct {
}

func NewLoginLogController() *LoginLogController {
	return &LoginLogController{}
}

// Index 登录日志列表
func (r *LoginLogController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	adminID := ctx.Request().Query("admin_id", "")
	username := ctx.Request().Query("username", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	startTime := ctx.Request().Query("start_time", "")
	endTime := ctx.Request().Query("end_time", "")

	query := facades.Orm().Query().Model(&models.LoginLog{})

	if adminID != "" {
		query = query.Where("admin_id", adminID)
	}
	if username != "" {
		query = query.Where("username", "like", "%"+username+"%")
	}
	if ip != "" {
		query = query.Where("ip", "like", "%"+ip+"%")
	}
	if status != "" {
		query = query.Where("status", status)
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

	var logs []models.LoginLog
	offset := (page - 1) * pageSize
	// 使用 With 预加载关联，避免 N+1 查询问题
	if err = query.With("Admin").Offset(offset).Limit(pageSize).Order("id desc").Get(&logs); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", logs, total, page, pageSize)
}

// Show 登录日志详情
func (r *LoginLogController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var log models.LoginLog
	// 使用 With 预加载关联
	if err := facades.Orm().Query().With("Admin").Where("id", id).First(&log); err != nil {
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"log": log,
	})
}

// Destroy 删除登录日志
func (r *LoginLogController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var log models.LoginLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&log); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// LoginLogBatchDestroyRequest 批量删除请求结构
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

	// 转换为 []any
	idsAny := make([]any, len(ids))
	for i, id := range ids {
		idsAny[i] = id
	}

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.LoginLog{}); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// Clean 清理登录日志
func (r *LoginLogController) Clean(ctx http.Context) http.Response {
	days := cast.ToInt(ctx.Request().Input("days", "30"))
	if days <= 0 {
		days = 30
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := facades.Orm().Query().Model(&models.LoginLog{}).Where("created_at < ?", cutoffTime).Delete(&models.LoginLog{}); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "clean_failed")
	}

	return response.Success(ctx, "clean_success")
}
