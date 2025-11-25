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
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if slug != "" {
		query = query.Where("slug LIKE ?", "%"+slug+"%")
	}
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
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
	// 处理状态字段：需要正确处理 0 值
	// 处理状态字段：需要正确处理 0 值
	// 使用 All() 方法获取所有输入数据，确保能正确获取 JSON 数据中的 0 值
	allInputs := ctx.Request().All()
	var status uint8 = 1 // 默认启用
	if statusVal, exists := allInputs["status"]; exists {
		if statusVal != nil {
			status = cast.ToUint8(statusVal)
		}
	}
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))
	menuID := cast.ToUint(ctx.Request().Input("menu_id", "0"))

	if name == "" || slug == "" {
		return response.Error(ctx, http.StatusBadRequest, "permission_name_and_slug_required")
	}

	// 检查名称或标识是否已存在
	var existPermission models.Permission
	if err := facades.Orm().Query().Where("name", name).OrWhere("slug", slug).First(&existPermission); err == nil {
		return response.Error(ctx, http.StatusBadRequest, "permission_name_or_slug_exists")
	}

	// 使用 map 方式创建，确保零值字段（status=0）也能被正确保存
	permissionData := map[string]interface{}{
		"name":        name,
		"slug":        slug,
		"method":      method,
		"path":        path,
		"description": description,
		"status":      status, // 明确设置 status，即使是 0 也会被保存
		"sort":        sort,
		"menu_id":     menuID,
	}

	if err := facades.Orm().Query().Table("permissions").Create(permissionData); err != nil {
		facades.Log().Errorf("Create permission error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	// 获取创建后的记录
	var permission models.Permission
	if err := facades.Orm().Query().Where("slug", slug).First(&permission); err != nil {
		facades.Log().Errorf("Get created permission error: %v", err)
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
	menuID := ctx.Request().Input("menu_id", "")

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
	if menuID != "" {
		permission.MenuID = cast.ToUint(menuID)
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
