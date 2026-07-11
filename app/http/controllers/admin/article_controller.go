package admin

import (
	"encoding/json"
	appfacades "goravel/app/facades"

	"strings"

	"github.com/goravel/framework/contracts/http"

	"github.com/goravel/framework/contracts/queue"

	"github.com/goravel/framework/facades"

	"goravel/app/jobs"
	"goravel/app/models"
	"goravel/app/utils"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type ArticleController struct {}


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
	return services.BuildArticleFiltersFromHTTP(ctx)
}

func NewArticleController() *ArticleController {
	return &ArticleController{}
}

func (c *ArticleController) ArticleService(ctx http.Context) services.ArticleService {
	return services.NewArticleService(ctx)
}


// Index lists Article records.
func (c *ArticleController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := c.buildArticleFilters(ctx)

	list, total, err := c.ArticleService(ctx).GetList(filters, page, pageSize)
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

// Show returns Article details.
func (c *ArticleController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	item, err := c.ArticleService(ctx).GetByID(id)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusNotFound, err)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Store creates a new Article.
func (c *ArticleController) Store(ctx http.Context) http.Response {
	var req adminrequests.ArticleCreate
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.ArticleService(ctx).Create(&req)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Update modifies an existing Article.
func (c *ArticleController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.ArticleUpdate
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.ArticleService(ctx).Update(id, &req)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Destroy deletes a Article.
func (c *ArticleController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.ArticleService(ctx).Delete(id); err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "delete_success", http.Json{})
}

// Export exports Article records.
func (c *ArticleController) Export(ctx http.Context) http.Response {
	lock := helpers.AcquireExportLock(ctx, "articles")
	if lock.Unauthorized {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code)
	}
	if lock.Blocked {
		return response.Error(ctx, http.StatusTooManyRequests, apperrors.ErrGetLockFailed.Code)
	}
	adminID := lock.AdminID

	filters := c.buildArticleFilters(ctx)
	filtersMap := utils.ExportFiltersToMap(filters)
	lang := utils.GetCurrentLanguage(ctx)
	timezone := helpers.GetCurrentTimezone(ctx)

	exportRecord := models.Export{
		AdminID: adminID,
		Type:    "articles",
		Status:  models.ExportStatusProcessing,
		Disk:    helpers.ResolveExportDisk(ctx),
		Path:    "",
	}
	if err := appfacades.OrmQuery(ctx).Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgsStruct := jobs.ExportArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "articles",
		Language: lang,
		Timezone: timezone,
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportArticles{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   "queued",
	})
}
