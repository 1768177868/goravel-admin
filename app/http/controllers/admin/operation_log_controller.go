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
	"goravel/app/utils/errorlog"
)

type OperationLogController struct {
}

func NewOperationLogController() *OperationLogController {
	return &OperationLogController{}
}

// Index 获取操作日志列表
func (r *OperationLogController) Index(ctx http.Context) http.Response {
	// 验证并规范化分页参数
	page, pageSize := helpers.ValidatePagination(
		helpers.GetIntQuery(ctx, "page", 1),
		helpers.GetIntQuery(ctx, "page_size", 10),
	)

	// 构建查询
	query := r.buildQuery(ctx)

	// 获取总数
	total, err := query.Count()
	if err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Failed to count operation logs", map[string]any{
			"error": err.Error(),
		}, "Count operation logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	// 应用排序和分页
	orderBy := ctx.Request().Query("order_by", "")
	query = helpers.ApplySort(query, orderBy, "id:desc")
	offset := (page - 1) * pageSize

	var logs []models.OperationLog
	if err = query.With("Admin").Offset(offset).Limit(pageSize).Get(&logs); err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Failed to query operation logs", map[string]any{
			"error": err.Error(),
		}, "Query operation logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", logs, total, page, pageSize)
}

// buildQuery 构建操作日志查询
func (r *OperationLogController) buildQuery(ctx http.Context) orm.Query {
	query := facades.Orm().Query().Model(&models.OperationLog{})

	adminID := ctx.Request().Query("admin_id", "")
	username := ctx.Request().Query("username", "")
	method := ctx.Request().Query("method", "")
	path := ctx.Request().Query("path", "")
	title := ctx.Request().Query("title", "")
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	if adminID != "" {
		query = query.Where("admin_id", adminID)
	}
	if username != "" {
		var adminIDs []uint
		var admins []models.Admin
		if err := facades.Orm().Query().Where("username LIKE ?", "%"+username+"%").Get(&admins); err == nil {
			for _, admin := range admins {
				adminIDs = append(adminIDs, admin.ID)
			}
			if len(adminIDs) > 0 {
				idsAny := helpers.ConvertUintSliceToAny(adminIDs)
				query = query.WhereIn("admin_id", idsAny)
			} else {
				query = query.Where("admin_id", 0)
			}
		}
	}
	if method != "" {
		query = query.Where("method = ?", method)
	}
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
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

	return query
}

// Show 获取操作日志详情
func (r *OperationLogController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var log models.OperationLog
	if err := facades.Orm().Query().With("Admin").Where("id", id).First(&log); err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Operation log not found", map[string]any{
			"error":  err.Error(),
			"log_id": id,
		}, "Operation log not found: %v", err)
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"log": log,
	})
}

// Destroy 删除操作日志
func (r *OperationLogController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var log models.OperationLog
	if err := facades.Orm().Query().Where("id", id).First(&log); err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Operation log not found for delete", map[string]any{
			"error":  err.Error(),
			"log_id": id,
		}, "Operation log not found: %v", err)
		return response.Error(ctx, http.StatusNotFound, "log_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&log); err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Failed to delete operation log", map[string]any{
			"error":  err.Error(),
			"log_id": log.ID,
		}, "Delete operation log error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

type OperationLogBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除操作日志
func (r *OperationLogController) BatchDestroy(ctx http.Context) http.Response {
	var req OperationLogBatchDestroyRequest

	// 使用结构体绑定
	if err := ctx.Request().Bind(&req); err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Failed to bind batch delete request", map[string]any{
			"error": err.Error(),
		}, "Bind batch delete request error: %v", err)
		return response.Error(ctx, http.StatusBadRequest, "params_error")
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "ids_required")
	}

	ids := req.IDs

	// 使用工具函数转换为 []any
	idsAny := helpers.ConvertUintSliceToAny(ids)

	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.OperationLog{}); err != nil {
		errorlog.RecordHTTP(ctx, "operation-log", "Failed to batch delete operation logs", map[string]any{
			"error": err.Error(),
			"ids":   ids,
		}, "Batch delete operation logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
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
		errorlog.RecordHTTP(ctx, "operation-log", "Failed to clean operation logs", map[string]any{
			"error": err.Error(),
			"days":  days,
		}, "Clean operation logs error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "clean_failed")
	}

	return response.Success(ctx, "clean_success")
}

// GetTitleOptions 获取所有可用的操作标题选项
func (r *OperationLogController) GetTitleOptions(ctx http.Context) http.Response {
	// 从数据库查询已存在的标题
	var dbTitles []string
	_ = facades.Orm().Query().Model(&models.OperationLog{}).
		Select("DISTINCT title").
		Where("title IS NOT NULL AND title != ''").
		Order("title ASC").
		Pluck("title", &dbTitles)

	// 从配置中获取所有可能的操作标题
	allTitleKeysInterface := facades.Config().Get("operation_log.all_title_keys", []string{})
	var configTitles []string
	if allTitleKeys, ok := allTitleKeysInterface.([]interface{}); ok {
		for _, keyInterface := range allTitleKeys {
			if key, ok := keyInterface.(string); ok {
				configTitles = append(configTitles, key)
			}
		}
	} else if allTitleKeys, ok := allTitleKeysInterface.([]string); ok {
		configTitles = allTitleKeys
	}

	// 合并数据库标题和配置标题，去重
	uniqueTitles := make(map[string]bool)
	var result []string

	// 先添加配置中的所有标题（确保所有可能的标题都在列表中）
	for _, title := range configTitles {
		if title != "" && !uniqueTitles[title] {
			uniqueTitles[title] = true
			result = append(result, title)
		}
	}

	// 再添加数据库中存在的标题（可能有一些不在配置中的标题）
	for _, title := range dbTitles {
		if title != "" && !uniqueTitles[title] {
			uniqueTitles[title] = true
			result = append(result, title)
		}
	}

	// 排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return response.Success(ctx, "get_success", http.Json{
		"titles": result,
	})
}
