package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type RoleController struct {
	roleService services.RoleService
}

func NewRoleController() *RoleController {
	return &RoleController{
		roleService: services.NewRoleServiceImpl(),
	}
}

// Index 角色列表
func (r *RoleController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	name := ctx.Request().Query("name", "")
	status := ctx.Request().Query("status", "")
	orderBy := ctx.Request().Query("order_by", "")

	query := facades.Orm().Query().Model(&models.Role{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var roles []models.Role
	offset := (page - 1) * pageSize
	// 应用排序（默认按排序字段升序，创建时间倒序）
	query = helpers.ApplySort(query, orderBy, "sort:asc,created_at:desc")
	if err := query.Offset(offset).Limit(pageSize).Get(&roles); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", roles, total, page, pageSize)
}

// Show 角色详情
func (r *RoleController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var role models.Role
	// 使用 With 预加载关联
	if err := facades.Orm().Query().With("Permissions").With("Menus").Where("id", id).First(&role); err != nil {
		return response.Error(ctx, http.StatusNotFound, "role_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"role": role,
	})
}

// Store 创建角色
func (r *RoleController) Store(ctx http.Context) http.Response {
	name := ctx.Request().Input("name")
	slug := ctx.Request().Input("slug")
	description := ctx.Request().Input("description")
	statusInput := ctx.Request().Input("status")
	var status uint8 = 1 // 默认启用
	if statusInput != "" {
		status = cast.ToUint8(statusInput)
	}
	facades.Log().Infof("Create role - statusInput: %s, status: %d", statusInput, status)
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))

	if name == "" || slug == "" {
		return response.Error(ctx, http.StatusBadRequest, "role_name_and_slug_required")
	}

	// 检查名称是否已存在
	nameCount, err := facades.Orm().Query().Model(&models.Role{}).Where("name", name).Count()
	if err == nil && nameCount > 0 {
		return response.Error(ctx, http.StatusBadRequest, "role_name_exists")
	}
	
	// 检查标识是否已存在
	slugCount, err := facades.Orm().Query().Model(&models.Role{}).Where("slug", slug).Count()
	if err == nil && slugCount > 0 {
		return response.Error(ctx, http.StatusBadRequest, "role_slug_exists")
	}

	role := models.Role{
		Name:        name,
		Slug:        slug,
		Description: description,
		Status:      status,
		Sort:        sort,
	}

	if err := facades.Orm().Query().Create(&role); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	// 处理权限关联
	permissionIDs := r.roleService.ParseIDsFromRequest(ctx, "permission_ids")
	if len(permissionIDs) > 0 {
		if err := r.roleService.SyncPermissions(&role, permissionIDs); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "create_failed")
		}
	}

	// 处理菜单关联
	menuIDs := r.roleService.ParseIDsFromRequest(ctx, "menu_ids")
	if len(menuIDs) > 0 {
		if err := r.roleService.SyncMenus(&role, menuIDs); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "create_failed")
		}
	}

	return response.Success(ctx, "create_success", http.Json{
		"role": role,
	})
}

// Update 更新角色
func (r *RoleController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var role models.Role
	if err := facades.Orm().Query().Where("id", id).First(&role); err != nil {
		return response.Error(ctx, http.StatusNotFound, "role_not_found")
	}

	name := ctx.Request().Input("name")
	slug := ctx.Request().Input("slug")
	description := ctx.Request().Input("description")
	status := ctx.Request().Input("status", "")
	sort := ctx.Request().Input("sort", "")

	if name != "" {
		// 检查名称是否已被其他角色使用（排除当前角色）
		count, err := facades.Orm().Query().Model(&models.Role{}).Where("name", name).Where("id != ?", id).Count()
		if err == nil && count > 0 {
			return response.Error(ctx, http.StatusBadRequest, "role_name_exists")
		}
		role.Name = name
	}
	if slug != "" {
		// 检查标识是否已被其他角色使用（排除当前角色）
		count, err := facades.Orm().Query().Model(&models.Role{}).Where("slug", slug).Where("id != ?", id).Count()
		if err == nil && count > 0 {
			return response.Error(ctx, http.StatusBadRequest, "role_slug_exists")
		}
		role.Slug = slug
	}
	if description != "" {
		role.Description = description
	}
	if status != "" {
		role.Status = cast.ToUint8(status)
	}
	if sort != "" {
		role.Sort = cast.ToInt(sort)
	}

	if err := facades.Orm().Query().Save(&role); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	// 处理权限关联
	if ctx.Request().Input("permission_ids") != "" {
		permissionIDs := r.roleService.ParseIDsFromRequest(ctx, "permission_ids")
		if err := r.roleService.SyncPermissions(&role, permissionIDs); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
	}

	// 处理菜单关联
	if ctx.Request().Input("menu_ids") != "" {
		menuIDs := r.roleService.ParseIDsFromRequest(ctx, "menu_ids")
		if err := r.roleService.SyncMenus(&role, menuIDs); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
	}

	return response.Success(ctx, "update_success", http.Json{
		"role": role,
	})
}

// Destroy 删除角色
func (r *RoleController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var role models.Role
	if err := facades.Orm().Query().Where("id", id).First(&role); err != nil {
		return response.Error(ctx, http.StatusNotFound, "role_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&role); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
