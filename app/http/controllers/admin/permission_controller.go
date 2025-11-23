package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
)

type PermissionController struct {
}

func NewPermissionController() *PermissionController {
	return &PermissionController{}
}

// Index 权限列表
func (r *PermissionController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	name := ctx.Request().Query("name", "")
	slug := ctx.Request().Query("slug", "")
	method := ctx.Request().Query("method", "")
	path := ctx.Request().Query("path", "")
	status := ctx.Request().Query("status", "")

	query := facades.Orm().Query().Model(&models.Permission{})

	if name != "" {
		query = query.Where("name", "like", "%"+name+"%")
	}
	if slug != "" {
		query = query.Where("slug", "like", "%"+slug+"%")
	}
	if method != "" {
		query = query.Where("method", method)
	}
	if path != "" {
		query = query.Where("path", "like", "%"+path+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var permissions []models.Permission
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort asc, id desc").Get(&permissions); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", permissions, total, page, pageSize)
}

// Show 权限详情
func (r *PermissionController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var permission models.Permission
	if err := facades.Orm().Query().Where("id", id).First(&permission); err != nil {
		return response.Error(ctx, http.StatusNotFound, "permission_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"permission": permission,
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

	if name == "" || slug == "" {
		return response.Error(ctx, http.StatusBadRequest, "permission_name_and_slug_required")
	}

	// 检查名称或标识是否已存在
	var existPermission models.Permission
	if err := facades.Orm().Query().Where("name", name).OrWhere("slug", slug).First(&existPermission); err == nil {
		return response.Error(ctx, http.StatusBadRequest, "permission_name_or_slug_exists")
	}

	permission := models.Permission{
		Name:        name,
		Slug:        slug,
		Method:      method,
		Path:        path,
		Description: description,
		Status:      status,
		Sort:        sort,
	}

	if err := facades.Orm().Query().Create(&permission); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return response.Success(ctx, "create_success", http.Json{
		"permission": permission,
	})
}

// Update 更新权限
func (r *PermissionController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var permission models.Permission
	if err := facades.Orm().Query().Where("id", id).First(&permission); err != nil {
		return response.Error(ctx, http.StatusNotFound, "permission_not_found")
	}

	name := ctx.Request().Input("name")
	slug := ctx.Request().Input("slug")
	method := ctx.Request().Input("method")
	path := ctx.Request().Input("path")
	description := ctx.Request().Input("description")
	status := ctx.Request().Input("status", "")
	sort := ctx.Request().Input("sort", "")

	if name != "" {
		// 检查名称是否已被其他权限使用
		var existPermission models.Permission
		if err := facades.Orm().Query().Where("name", name).Where("id <> ?", id).First(&existPermission); err == nil {
			return response.Error(ctx, http.StatusBadRequest, "permission_name_exists")
		}
		permission.Name = name
	}
	if slug != "" {
		// 检查标识是否已被其他权限使用
		var existPermission models.Permission
		if err := facades.Orm().Query().Where("slug", slug).Where("id <> ?", id).First(&existPermission); err == nil {
			return response.Error(ctx, http.StatusBadRequest, "permission_slug_exists")
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

	if err := facades.Orm().Query().Save(&permission); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"permission": permission,
	})
}

// Destroy 删除权限
func (r *PermissionController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var permission models.Permission
	if err := facades.Orm().Query().Where("id", id).First(&permission); err != nil {
		return response.Error(ctx, http.StatusNotFound, "permission_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&permission); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
