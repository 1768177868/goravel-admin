package admin

import (
	"sort"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/constants"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
)

type OperationLogController struct {
}

func NewOperationLogController() *OperationLogController {
	return &OperationLogController{}
}

// findOperationLogByID 根据ID查找操作日志，如果不存在则返回错误响应
// withAdmin 为 true 时会预加载 Admin 关联
func (r *OperationLogController) findOperationLogByID(ctx http.Context, id uint, withAdmin bool) (*models.OperationLog, http.Response) {
	var relations []string
	if withAdmin {
		relations = append(relations, "Admin")
	}
	return response.FindByID[models.OperationLog](ctx, id, &response.FindByIDOptions{
		WithRelations:      relations,
		NotFoundMessageKey: "log_not_found",
	})
}

// Index 获取操作日志列表
func (r *OperationLogController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var logs []models.OperationLog
	return response.PaginateQuery(ctx, query, &logs, &response.PaginateQueryOptions{
		WithRelations: []string{"Admin"},
		ErrorModule:   "operation-log",
	})
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

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 id desc
	query = helpers.ApplySort(query, orderBy, "id:desc")

	return query
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
		return response.Error(ctx, http.StatusBadRequest, "params_error")
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "ids_required")
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

	// 合并数据库标题和配置标题，去重
	uniqueTitles := make(map[string]bool)
	var result []string

	// 只使用数据库中存在的标题（权限标识），忽略旧的 operation.xxx 配置
	for _, title := range dbTitles {
		// 排除空标题、未知标题以及旧的 operation.xxx 标题
		if title == "" || title == "operation.unknown" || strings.HasPrefix(title, "operation.") {
			continue
		}
		if !uniqueTitles[title] {
			uniqueTitles[title] = true
			result = append(result, title)
		}
	}

	sort.Strings(result)

	return response.Success(ctx, http.Json{
		"titles": result,
	})
}
