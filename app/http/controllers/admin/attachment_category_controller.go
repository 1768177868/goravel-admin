package admin

import (
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type AttachmentCategoryController struct{}

func NewAttachmentCategoryController() *AttachmentCategoryController {
	return &AttachmentCategoryController{}
}

func (r *AttachmentCategoryController) categoryService(ctx http.Context) services.AttachmentCategoryService {
	return services.NewAttachmentCategoryService(ctx)
}

func (r *AttachmentCategoryController) findByID(ctx http.Context, id uint) (*models.AttachmentCategory, http.Response) {
	category, err := r.categoryService(ctx).GetByID(id)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrAttachmentCategoryNotFound.Code)
	}
	return category, nil
}

func (r *AttachmentCategoryController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 100)
	filters := services.AttachmentCategoryFilters{
		Name:    ctx.Request().Query("name", ""),
		Status:  ctx.Request().Query("status", ""),
		OrderBy: ctx.Request().Query("order_by", ""),
	}
	list, total, err := r.categoryService(ctx).GetList(filters, page, pageSize)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (r *AttachmentCategoryController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	category, resp := r.findByID(ctx, id)
	if resp != nil {
		return resp
	}
	return response.Success(ctx, http.Json{"category": *category})
}

func (r *AttachmentCategoryController) Store(ctx http.Context) http.Response {
	var req adminrequests.AttachmentCategoryCreate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}
	status := req.Status
	if _, exists := ctx.Request().All()["status"]; !exists {
		status = 1
	}
	category, err := r.categoryService(ctx).Create(req.Name, req.Remark, status, req.Sort)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment_category", err, map[string]any{"name": req.Name})
	}
	return response.Success(ctx, http.Json{"category": category})
}

func (r *AttachmentCategoryController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	category, resp := r.findByID(ctx, id)
	if resp != nil {
		return resp
	}
	var req adminrequests.AttachmentCategoryUpdate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}
	allInputs := ctx.Request().All()
	if _, exists := allInputs["name"]; exists {
		category.Name = req.Name
	}
	if _, exists := allInputs["status"]; exists {
		category.Status = req.Status
	}
	if _, exists := allInputs["sort"]; exists {
		category.Sort = req.Sort
	}
	if _, exists := allInputs["remark"]; exists {
		category.Remark = req.Remark
	}
	if err := r.categoryService(ctx).Update(category); err != nil {
		if businessErr, ok := err.(*apperrors.BusinessError); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr)
		}
		return response.ErrorWithLog(ctx, "attachment_category", err, map[string]any{"id": id})
	}
	return response.Success(ctx, http.Json{"category": category})
}

func (r *AttachmentCategoryController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	category, resp := r.findByID(ctx, id)
	if resp != nil {
		return resp
	}
	if err := r.categoryService(ctx).Delete(category); err != nil {
		if businessErr, ok := err.(*apperrors.BusinessError); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr)
		}
		return response.ErrorWithLog(ctx, "attachment_category", err, map[string]any{"id": id})
	}
	return response.Success(ctx)
}
