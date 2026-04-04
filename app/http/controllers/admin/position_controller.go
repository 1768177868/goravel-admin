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

type PositionController struct {
	positionService services.PositionService
}

func NewPositionController() *PositionController {
	return &PositionController{
		positionService: services.NewPositionService(),
	}
}

func (r *PositionController) findPositionByID(ctx http.Context, id uint) (*models.Position, http.Response) {
	position, err := r.positionService.GetByID(id)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrPositionNotFound.Code)
	}
	return position, nil
}

func (r *PositionController) buildFilters(ctx http.Context) services.PositionFilters {
	return services.PositionFilters{
		Name:      ctx.Request().Query("name", ""),
		Status:    ctx.Request().Query("status", ""),
		StartTime: helpers.GetTimeQueryParam(ctx, "start_time"),
		EndTime:   helpers.GetTimeQueryParam(ctx, "end_time"),
		OrderBy:   ctx.Request().Query("order_by", ""),
	}
}

func (r *PositionController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)
	filters := r.buildFilters(ctx)
	list, total, err := r.positionService.GetList(filters, page, pageSize)
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

func (r *PositionController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	position, resp := r.findPositionByID(ctx, id)
	if resp != nil {
		return resp
	}
	return response.Success(ctx, http.Json{
		"position": *position,
	})
}

func (r *PositionController) Store(ctx http.Context) http.Response {
	var req adminrequests.PositionCreate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}
	position, err := r.positionService.Create(req.Name, req.Code, req.Remark, req.Status, req.Sort)
	if err != nil {
		return response.ErrorWithLog(ctx, "position", err, map[string]any{"name": req.Name})
	}
	return response.Success(ctx, http.Json{
		"position": position,
	})
}

func (r *PositionController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	position, resp := r.findPositionByID(ctx, id)
	if resp != nil {
		return resp
	}
	var req adminrequests.PositionUpdate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}
	allInputs := ctx.Request().All()
	if _, exists := allInputs["name"]; exists {
		position.Name = req.Name
	}
	if _, exists := allInputs["code"]; exists {
		position.Code = req.Code
	}
	if _, exists := allInputs["status"]; exists {
		position.Status = req.Status
	}
	if _, exists := allInputs["sort"]; exists {
		position.Sort = req.Sort
	}
	if _, exists := allInputs["remark"]; exists {
		position.Remark = req.Remark
	}
	if err := r.positionService.Update(position); err != nil {
		return response.ErrorWithLog(ctx, "position", err, map[string]any{"position_id": position.ID})
	}
	return response.Success(ctx, http.Json{
		"position": *position,
	})
}

func (r *PositionController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	position, resp := r.findPositionByID(ctx, id)
	if resp != nil {
		return resp
	}
	hasAdmins, err := r.positionService.HasAdmins(id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	if hasAdmins {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrPositionHasAdmins.Code)
	}
	if err := r.positionService.Delete(position); err != nil {
		return response.ErrorWithLog(ctx, "position", err, map[string]any{"position_id": position.ID})
	}
	return response.Success(ctx)
}
