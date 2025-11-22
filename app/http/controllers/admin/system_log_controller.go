package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
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
	startTime := ctx.Request().Query("start_time", "")
	endTime := ctx.Request().Query("end_time", "")

	query := facades.Orm().Query().Model(&models.SystemLog{})

	if level != "" {
		query = query.Where("level", level)
	}
	if module != "" {
		query = query.Where("module", "like", "%"+module+"%")
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
	if err = query.Offset(offset).Limit(pageSize).Order("id desc").Get(&logs); err != nil {
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

// Store 创建系统日志（不支持，返回405）
func (r *SystemLogController) Store(ctx http.Context) http.Response {
	return ctx.Response().Json(http.StatusMethodNotAllowed, http.Json{
		"code":    405,
		"message": "Method Not Allowed",
	})
}

// Update 更新系统日志（不支持，返回405）
func (r *SystemLogController) Update(ctx http.Context) http.Response {
	return ctx.Response().Json(http.StatusMethodNotAllowed, http.Json{
		"code":    405,
		"message": "Method Not Allowed",
	})
}
