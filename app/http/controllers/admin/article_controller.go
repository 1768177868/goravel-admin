package admin

import (
	"encoding/json"
	"fmt"

	"strings"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"

	"goravel/app/jobs"
	"goravel/app/models"

	"goravel/app/services"

	"goravel/app/utils"
)

type ArticleController struct {
	ArticleService services.ArticleService
}

func handleGeneratedServiceError(ctx http.Context, status int, err error) http.Response {
	if businessErr, ok := apperrors.GetBusinessError(err); ok {
		if businessErr.Code == "params_error" || businessErr.Code == "invalid_argument" {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		if businessErr.Code == "record_not_found" || strings.HasSuffix(businessErr.Code, "_not_found") {
			return response.Error(ctx, http.StatusNotFound, businessErr.Code)
		}
		return response.Error(ctx, status, businessErr.Code)
	}
	return response.Error(ctx, status, err.Error())
}

func validateGeneratedRequest(ctx http.Context, req http.FormRequest) http.Response {
	validationErrors, err := ctx.Request().ValidateRequest(req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if validationErrors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", validationErrors.All())
	}
	return nil
}

func (c *ArticleController) buildArticleFilters(ctx http.Context) services.ArticleFilters {
	return services.ArticleFilters{
		AdminId:        ctx.Request().Query("admin_id", ""),
		Title:          ctx.Request().Query("title", ""),
		Content:        ctx.Request().Query("content", ""),
		Status:         ctx.Request().Query("status", ""),
		CreatedAt:      ctx.Request().Query("created_at", ""),
		CreatedAtStart: ctx.Request().Query("created_at_start", ""),
		CreatedAtEnd:   ctx.Request().Query("created_at_end", ""),
		UpdatedAt:      ctx.Request().Query("updated_at", ""),
		UpdatedAtStart: ctx.Request().Query("updated_at_start", ""),
		UpdatedAtEnd:   ctx.Request().Query("updated_at_end", ""),
	}
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		ArticleService: services.NewArticleService(),
	}
}

// Index Article列表
func (c *ArticleController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := c.buildArticleFilters(ctx)

	list, total, err := c.ArticleService.GetList(filters, page, pageSize)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show Article详情
func (c *ArticleController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	item, err := c.ArticleService.GetByID(id)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusNotFound, err)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Store 创建Article
func (c *ArticleController) Store(ctx http.Context) http.Response {
	var req adminrequests.ArticleCreate
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.ArticleService.Create(&req)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Update 更新Article
func (c *ArticleController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.ArticleUpdate
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.ArticleService.Update(id, &req)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Destroy 删除Article
func (c *ArticleController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.ArticleService.Delete(id); err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "delete_success", http.Json{})
}

// Export 导出Article
func (c *ArticleController) Export(ctx http.Context) http.Response {
	filters := c.buildArticleFilters(ctx)
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	lockKey := fmt.Sprintf("export:articles:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10)
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "export_in_progress")
	}

	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = utils.GetConfigValue("storage", "export_disk", "")
	}
	if disk == "" {
		disk = "local"
	}

	exportRecord := models.Export{
		AdminID: adminID,
		Type:    "articles",
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "",
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		lock.Release()
		return response.ErrorWithLog(ctx, "export", err)
	}

	filtersMap := map[string]any{}
	if filters.AdminId != "" {
		filtersMap["admin_id"] = filters.AdminId
	}
	if filters.Title != "" {
		filtersMap["title"] = filters.Title
	}
	if filters.Content != "" {
		filtersMap["content"] = filters.Content
	}
	if filters.Status != "" {
		filtersMap["status"] = filters.Status
	}
	if filters.CreatedAt != "" {
		filtersMap["created_at"] = filters.CreatedAt
	}
	if filters.CreatedAtStart != "" {
		filtersMap["created_at_start"] = filters.CreatedAtStart
	}
	if filters.CreatedAtEnd != "" {
		filtersMap["created_at_end"] = filters.CreatedAtEnd
	}
	if filters.UpdatedAt != "" {
		filtersMap["updated_at"] = filters.UpdatedAt
	}
	if filters.UpdatedAtStart != "" {
		filtersMap["updated_at_start"] = filters.UpdatedAtStart
	}
	if filters.UpdatedAtEnd != "" {
		filtersMap["updated_at_end"] = filters.UpdatedAtEnd
	}

	exportArgsStruct := jobs.ExportGenericArgs{
		ExportArgs: jobs.ExportArgs{
			ExportID: exportRecord.ID,
			AdminID:  adminID,
			Filters:  filtersMap,
			Type:     "articles",
			Language: utils.GetCurrentLanguage(ctx),
			Timezone: helpers.GetCurrentTimezone(ctx),
		},
		Table:      "articles",
		FilePrefix: "articles",
		HeaderKeys: []string{
			"admin_id",
			"title",
			"content",
			"status",
			"created_at",
			"updated_at",
		},
		Columns: []string{
			"admin_id",
			"title",
			"content",
			"status",
			"created_at",
			"updated_at",
		},
		SearchTypes: map[string]string{
			"admin_id":   "=",
			"title":      "like",
			"content":    "like",
			"status":     "=",
			"created_at": "=",
			"updated_at": "=",
		},
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		_ = facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportGeneric{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		_ = facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	lock.Release()
	exportID := exportRecord.ID

	return response.Success(ctx, "export_task_submitted", http.Json{
		"export_id": exportID,
	})
}
