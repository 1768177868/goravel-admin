package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
)

type OperationLogController struct {
}

func NewOperationLogController() *OperationLogController {
	return &OperationLogController{}
}

// Index 操作日志列表
func (r *OperationLogController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	adminID := ctx.Request().Query("admin_id", "")
	username := ctx.Request().Query("username", "")
	method := ctx.Request().Query("method", "")
	path := ctx.Request().Query("path", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	startTime := ctx.Request().Query("start_time", "")
	endTime := ctx.Request().Query("end_time", "")
	orderBy := ctx.Request().Query("order_by", "")

	query := facades.Orm().Query().Model(&models.OperationLog{})

	if adminID != "" {
		query = query.Where("admin_id", adminID)
	}
	if username != "" {
		// 先查询匹配的管理员ID
		var adminIDs []uint
		var admins []models.Admin
		if err := facades.Orm().Query().Where("username", "like", "%"+username+"%").Get(&admins); err == nil {
			for _, admin := range admins {
				adminIDs = append(adminIDs, admin.ID)
			}
			if len(adminIDs) > 0 {
				idsAny := make([]any, len(adminIDs))
				for i, id := range adminIDs {
					idsAny[i] = id
				}
				query = query.WhereIn("admin_id", idsAny)
			} else {
				// 如果没有匹配的管理员，返回空结果
				query = query.Where("admin_id", 0)
			}
		}
	}
	if method != "" {
		query = query.Where("method", method)
	}
	if path != "" {
		query = query.Where("path", "like", "%"+path+"%")
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

	var logs []models.OperationLog
	offset := (page - 1) * pageSize
	// 应用排序（默认按创建时间倒序）
	query = helpers.ApplySort(query, orderBy, "created_at:desc")
	// 使用 With 预加载关联，避免 N+1 查询问题
	if err = query.With("Admin").Offset(offset).Limit(pageSize).Get(&logs); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", logs, total, page, pageSize)
}

// Show 操作日志详情
func (r *OperationLogController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var log models.OperationLog
	// 使用 With 预加载关联
	if err := facades.Orm().Query().With("Admin").Where("id", id).First(&log); err != nil {
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"log": log,
	})
}

// Destroy 删除操作日志
func (r *OperationLogController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var log models.OperationLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&log); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// OperationLogBatchDestroyRequest 批量删除请求结构
type OperationLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除操作日志
func (r *OperationLogController) BatchDestroy(ctx http.Context) http.Response {
	var req OperationLogBatchDestroyRequest
	
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

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.OperationLog{}); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// Clean 清理操作日志
func (r *OperationLogController) Clean(ctx http.Context) http.Response {
	days := cast.ToInt(ctx.Request().Input("days", "30"))
	if days <= 0 {
		days = 30
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	if _, err := facades.Orm().Query().Model(&models.OperationLog{}).Where("created_at < ?", cutoffTime).Delete(&models.OperationLog{}); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "clean_failed")
	}

	return response.Success(ctx, "clean_success")
}
