package admin

import (
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type <<.ControllerName>> struct {
	<<.ServiceName>> services.<<.ServiceName>>
}

func New<<.ControllerName>>() *<<.ControllerName>> {
	return &<<.ControllerName>>{
		<<.ServiceName>>: services.New<<.ServiceName>>(),
	}
}

// Index <<.ModelName>>列表
func (c *<<.ControllerName>>) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page",1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

<<range .SearchableFields>>
	<<.Name>> := ctx.Request().Query("<<.Name>>", "")
<<end>>

	filters := services.<<.ModelName>>Filters{
<<range .SearchableFields>>
		<<.PascalName>>: <<.Name>>,
<<end>>
	}

	list, total, err := c.<<.ServiceName>>.GetList(filters, page, pageSize)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show <<.ModelName>>详情
func (c *<<.ControllerName>>) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	item, err := c.<<.ServiceName>>.GetByID(id)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusNotFound, businessErr.Code)
		}
		return response.Error(ctx, http.StatusNotFound, err.Error())
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
}

// Store 创建<<.ModelName>>
func (c *<<.ControllerName>>) Store(ctx http.Context) http.Response {
	var req adminrequests.<<.RequestCreateName>>
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	item, err := c.<<.ServiceName>>.Create(&req)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
}

// Update 更新<<.ModelName>>
func (c *<<.ControllerName>>) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.<<.RequestUpdateName>>
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	item, err := c.<<.ServiceName>>.Update(id, &req)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
}

// Destroy 删除<<.ModelName>>
func (c *<<.ControllerName>>) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.<<.ServiceName>>.Delete(id); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, "delete_success", http.Json{})
}
