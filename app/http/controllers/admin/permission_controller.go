package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"
	appfacades "goravel/app/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type PermissionController struct {
	permissionService services.PermissionService
	treeService       services.TreeService
}

func NewPermissionController() *PermissionController {
	return &PermissionController{
		permissionService: services.NewPermissionService(),
		treeService:       services.NewTreeServiceImpl(),
	}
}

// findPermissionByID 根据ID查找权限，如果不存在则返回错误响应
// withMenu 为 true 时会预加载 Menu 关联
func (r *PermissionController) findPermissionByID(ctx http.Context, id uint, withMenu bool) (*models.Permission, http.Response) {
	permission, err := r.permissionService.GetByID(id, withMenu)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrPermissionNotFound.Code)
	}
	return permission, nil
}

// buildFilters 构建查询过滤器
func (r *PermissionController) buildFilters(ctx http.Context) services.PermissionFilters {
	name := ctx.Request().Query("name", "")
	slug := ctx.Request().Query("slug", "")
	method := ctx.Request().Query("method", "")
	path := ctx.Request().Query("path", "")
	status := ctx.Request().Query("status", "")
	menuID := ctx.Request().Query("menu_id", "")
	startTime := getTimeQueryUTC(ctx, "start_time")
	endTime := getTimeQueryUTC(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.PermissionFilters{
		Name:      name,
		Slug:      slug,
		Method:    method,
		Path:      path,
		Status:    status,
		MenuID:    menuID,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Index 权限列表
func (r *PermissionController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := r.buildFilters(ctx)

	permissions, total, err := r.permissionService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "permission", err)
	}

	return response.Success(ctx, http.Json{
		"list":      permissions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 权限详情
func (r *PermissionController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	permission, resp := r.findPermissionByID(ctx, id, true) // 预加载 Menu 关联
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"permission": *permission,
	})
}

// Store 创建权限
func (r *PermissionController) Store(ctx http.Context) http.Response {
	name := ctx.Request().Input("name")
	slug := ctx.Request().Input("slug")
	method := ctx.Request().Input("method")
	path := ctx.Request().Input("path")
	description := ctx.Request().Input("description")
	status := cast.ToUint8(ctx.Request().Input("status", "1"))
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))
	menuID := cast.ToUint(ctx.Request().Input("menu_id", "0"))

	if name == "" || slug == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrPermissionNameAndSlugRequired.Code)
	}

	exists, err := appfacades.OrmQuery(ctx).Model(&models.Permission{}).
		Where("name", name).
		OrWhere("slug", slug).
		Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrPermissionNameOrSlugExists.Code)
	}

	permission, err := r.permissionService.Create(
		name,
		slug,
		method,
		path,
		description,
		status,
		sort,
		menuID,
	)
	if err != nil {
		return response.ErrorWithLog(ctx, "permission", err, map[string]any{
			"name": name,
			"slug": slug,
		})
	}

	return response.Success(ctx, http.Json{
		"permission": *permission,
	})
}

func (r *PermissionController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	permission, resp := r.findPermissionByID(ctx, id, false)
	if resp != nil {
		return resp
	}

	name := ctx.Request().Input("name")
	slug := ctx.Request().Input("slug")
	method := ctx.Request().Input("method")
	path := ctx.Request().Input("path")
	description := ctx.Request().Input("description")
	status := ctx.Request().Input("status", "")
	sort := ctx.Request().Input("sort", "")
	menuIDStr := ctx.Request().Input("menu_id", "")

	if name != "" {
		exists, err := appfacades.OrmQuery(ctx).Model(&models.Permission{}).Where("name", name).Where("id <> ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrPermissionNameExists.Code)
		}
		permission.Name = name
	}
	if slug != "" {
		exists, err := appfacades.OrmQuery(ctx).Model(&models.Permission{}).Where("slug", slug).Where("id <> ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrPermissionSlugExists.Code)
		}
		permission.Slug = slug
	}
	if method != "" {
		permission.Method = method
	}
	if path != "" {
		permission.Path = path
	}
	if description != "" {
		permission.Description = description
	}
	if status != "" {
		permission.Status = cast.ToUint8(status)
	}
	if sort != "" {
		permission.Sort = cast.ToInt(sort)
	}
	if menuIDStr != "" {
		permission.MenuID = cast.ToUint(menuIDStr)
	}

	if err := r.permissionService.Update(permission); err != nil {
		return response.ErrorWithLog(ctx, "permission", err, map[string]any{
			"permission_id": permission.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"permission": *permission,
	})
}

// Destroy 删除权限
func (r *PermissionController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	permission, resp := r.findPermissionByID(ctx, id, false) // 不需要预加载关联
	if resp != nil {
		return resp
	}

	if err := r.permissionService.Delete(permission); err != nil {
		return response.ErrorWithLog(ctx, "permission", err, map[string]any{
			"permission_id": permission.ID,
		})
	}

	return response.Success(ctx)
}
