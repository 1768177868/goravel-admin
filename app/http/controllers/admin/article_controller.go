package admin

import (
	"fmt"
	"time"

	"strings"

	"github.com/goravel/framework/contracts/http"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
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
	return services.BuildArticleFiltersFromHTTP(ctx)
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		ArticleService: services.NewArticleService(),
	}
}

// Index lists Article records.
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

// Show returns Article details.
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

// Store creates a new Article.
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

// Update modifies an existing Article.
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

// Destroy deletes a Article.
func (c *ArticleController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.ArticleService.Delete(id); err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "delete_success", http.Json{})
}

// Export exports Article records.
func (c *ArticleController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code)
	}

	lockKey := fmt.Sprintf("export:articles:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, apperrors.ErrGetLockFailed.Code)
	}

	filters := c.buildArticleFilters(ctx)

	list, err := c.ArticleService.GetAllArticleForExport(filters)
	if err != nil {
		return response.ErrorWithLog(ctx, "article", err, map[string]any{
			"action":   "export_articles",
			"admin_id": adminID,
		})
	}

	headers := []string{
		"admin_id",
		"title",
		"content",
		"status",
		"created_at",
		"updated_at",
	}

	timezone := helpers.GetCurrentTimezone(ctx)
	var data [][]string
	for _, row := range list {
		r := []string{
			func() string {
				return cast.ToString(row.AdminId)
			}(),
			func() string {
				return row.Title
			}(),
			func() string {
				return row.Content
			}(),
			func() string {
				return cast.ToString(row.Status)
			}(),
			func() string {
				return helpers.FormatCarbonWithTimezone(row.CreatedAt, timezone)
			}(),
			func() string {
				return helpers.FormatCarbonWithTimezone(row.UpdatedAt, timezone)
			}(),
		}
		data = append(data, r)
	}

	ctx.WithValue("export_type", "articles")

	return response.Export(ctx, "exported", headers, data, "articles")
}
