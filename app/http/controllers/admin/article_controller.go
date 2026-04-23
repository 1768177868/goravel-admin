package admin

import (
	"github.com/goravel/framework/contracts/http"

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
		AdminId:   ctx.Request().Query("admin_id", ""),
		Title:     ctx.Request().Query("title", ""),
		Content:   ctx.Request().Query("content", ""),
		Status:    ctx.Request().Query("status", ""),
		CreatedAt: ctx.Request().Query("created_at", ""),
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
	if err := c.ArticleService.Export(filters); err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "export_task_submitted", http.Json{})
}
