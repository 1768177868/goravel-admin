package admin

import (
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/constants"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
)

type LoginLogController struct {
}

func NewLoginLogController() *LoginLogController {
	return &LoginLogController{}
}

// findLoginLogByID 根据ID查找登录日志，如果不存在则返回错误响应
// withAdmin 为 true 时会预加载 Admin 关联
func (r *LoginLogController) findLoginLogByID(ctx http.Context, id uint, withAdmin bool) (*models.LoginLog, http.Response) {
	var relations []string
	if withAdmin {
		relations = append(relations, "Admin")
	}
	return response.FindByID[models.LoginLog](ctx, id, &response.FindByIDOptions{
		WithRelations:      relations,
		NotFoundMessageKey: "log_not_found",
	})
}

// buildQuery 构建登录日志查询
func (r *LoginLogController) buildQuery(ctx http.Context) orm.Query {
	adminID := ctx.Request().Query("admin_id", "")
	username := ctx.Request().Query("username", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
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

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 id desc
	query = helpers.ApplySort(query, orderBy, "id:desc")

	return query
}

// Index 获取登录日志列表
func (r *LoginLogController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var logs []models.LoginLog
	return response.PaginateQuery(ctx, query, &logs, &response.PaginateQueryOptions{
		WithRelations: []string{"Admin"},
		ErrorModule:   "login-log",
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
		return response.Error(ctx, http.StatusBadRequest, "params_error")
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "ids_required")
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
