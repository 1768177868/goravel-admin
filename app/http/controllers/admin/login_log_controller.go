package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/constants"
	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type LoginLogController struct {
	loginLogService services.LoginLogService
}

func NewLoginLogController() *LoginLogController {
	return &LoginLogController{
		loginLogService: services.NewLoginLogService(),
	}
}

// findLoginLogByID 根据ID查找登录日志，如果不存在则返回错误响应
// withAdmin 为 true 时会预加载 Admin 关联
func (r *LoginLogController) findLoginLogByID(ctx http.Context, id uint, withAdmin bool) (*models.LoginLog, http.Response) {
	log, err := r.loginLogService.GetByID(id, withAdmin)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrLogNotFound.Code)
	}
	return log, nil
}

// buildFilters 构建查询过滤器
func (r *LoginLogController) buildFilters(ctx http.Context) services.LoginLogFilters {
	adminID := ctx.Request().Query("admin_id", "")
	username := ctx.Request().Query("username", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.LoginLogFilters{
		AdminID:   adminID,
		Username:  username,
		IP:        ip,
		Status:    status,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Index 获取登录日志列表
func (r *LoginLogController) Index(ctx http.Context) http.Response {
	filters := r.buildFilters(ctx)

	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	logs, total, err := r.loginLogService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "login-log", err)
	}

	return response.Success(ctx, http.Json{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 获取登录日志详情
func (r *LoginLogController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findLoginLogByID(ctx, id, true) // 预加载 Admin 关联
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"log": *log,
	})
}

// Destroy 删除登录日志
func (r *LoginLogController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	log, resp := r.findLoginLogByID(ctx, id, false)
	if resp != nil {
		return resp
	}

	if _, err := facades.Orm().Query().Delete(log); err != nil {
		return response.ErrorWithLog(ctx, "login-log", err, map[string]any{
			"log_id": log.ID,
		})
	}

	return response.Success(ctx)
}

type LoginLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除登录日志
func (r *LoginLogController) BatchDestroy(ctx http.Context) http.Response {
	var req LoginLogBatchDestroyRequest

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

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.LoginLog{}); err != nil {
		return response.ErrorWithLog(ctx, "login-log", err, map[string]any{
			"ids": ids,
		})
	}

	return response.Success(ctx)
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
		return response.ErrorWithLog(ctx, "login-log", err, map[string]any{
			"days": days,
		})
	}

	return response.Success(ctx)
}
