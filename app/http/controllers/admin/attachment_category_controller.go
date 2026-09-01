package admin

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type AttachmentCategoryController struct{}

func NewAttachmentCategoryController() *AttachmentCategoryController {
	return &AttachmentCategoryController{}
}

func (c *AttachmentCategoryController) buildAttachmentCategoryFilters(ctx http.Context) services.AttachmentCategoryFilters {
	return services.BuildAttachmentCategoryFiltersFromHTTP(ctx)
}

func (c *AttachmentCategoryController) AttachmentCategoryService(ctx http.Context) services.AttachmentCategoryService {
	return services.NewAttachmentCategoryService(ctx)
}

func (c *AttachmentCategoryController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 100)
	filters := c.buildAttachmentCategoryFilters(ctx)

	list, total, err := c.AttachmentCategoryService(ctx).GetList(filters, page, pageSize)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "attachment_category", http.StatusInternalServerError, err, nil)
	}

	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (c *AttachmentCategoryController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	category, err := c.AttachmentCategoryService(ctx).GetByID(id)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "attachment_category", http.StatusNotFound, err, map[string]any{"id": id})
	}

	return response.Success(ctx, http.Json{"category": category})
}

func (c *AttachmentCategoryController) Store(ctx http.Context) http.Response {
	var req adminrequests.AttachmentCategoryCreate
	if resp := ValidateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	category, err := c.AttachmentCategoryService(ctx).Create(&req)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "attachment_category", http.StatusInternalServerError, err, nil)
	}

	return response.Success(ctx, http.Json{"category": category})
}

func (c *AttachmentCategoryController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.AttachmentCategoryUpdate
	if resp := ValidateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	category, err := c.AttachmentCategoryService(ctx).Update(id, &req)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "attachment_category", http.StatusInternalServerError, err, map[string]any{"id": id})
	}

	return response.Success(ctx, http.Json{"category": category})
}

func (c *AttachmentCategoryController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.AttachmentCategoryService(ctx).Delete(id); err != nil {
		return HandleGeneratedServiceError(ctx, "attachment_category", http.StatusInternalServerError, err, map[string]any{"id": id})
	}

	return response.Success(ctx, "delete_success", http.Json{})
}
