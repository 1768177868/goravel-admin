package admin

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type PositionController struct{}

func NewPositionController() *PositionController {
	return &PositionController{}
}

func (c *PositionController) buildPositionFilters(ctx http.Context) services.PositionFilters {
	return services.BuildPositionFiltersFromHTTP(ctx)
}

func (c *PositionController) PositionService(ctx http.Context) services.PositionService {
	return services.NewPositionService(ctx)
}

func (c *PositionController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)
	filters := c.buildPositionFilters(ctx)

	list, total, err := c.PositionService(ctx).GetList(filters, page, pageSize)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "position", http.StatusInternalServerError, err, nil)
	}

	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (c *PositionController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	position, err := c.PositionService(ctx).GetByID(id)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "position", http.StatusNotFound, err, map[string]any{"id": id})
	}

	return response.Success(ctx, http.Json{
		"position": position,
	})
}

func (c *PositionController) Store(ctx http.Context) http.Response {
	var req adminrequests.PositionCreate
	if resp := ValidateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	position, err := c.PositionService(ctx).Create(&req)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "position", http.StatusInternalServerError, err, nil)
	}

	return response.Success(ctx, http.Json{
		"position": position,
	})
}

func (c *PositionController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.PositionUpdate
	if resp := ValidateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	position, err := c.PositionService(ctx).Update(id, &req)
	if err != nil {
		return HandleGeneratedServiceError(ctx, "position", http.StatusInternalServerError, err, map[string]any{"id": id})
	}

	return response.Success(ctx, http.Json{
		"position": position,
	})
}

func (c *PositionController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.PositionService(ctx).Delete(id); err != nil {
		return HandleGeneratedServiceError(ctx, "position", http.StatusInternalServerError, err, map[string]any{"id": id})
	}

	return response.Success(ctx, "delete_success", http.Json{})
}
