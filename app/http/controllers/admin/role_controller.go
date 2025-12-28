package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
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

// findRoleByID 根据ID查找角色，如果不存在则返回错误响应
// withRelations 为 true 时会预加载 Permissions 和 Menus 关联
func (r *RoleController) findRoleByID(ctx http.Context, id uint, withRelations bool) (*models.Role, http.Response) {
	role, err := r.roleService.GetByID(id, withRelations)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrRoleNotFound.Code)
	}
	return role, nil
}

// buildFilters 构建查询过滤器
func (r *RoleController) buildFilters(ctx http.Context) services.RoleFilters {
	name := ctx.Request().Query("name", "")
	status := ctx.Request().Query("status", "")
	// 使用辅助函数自动转换时区
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.RoleFilters{
		Name:      name,
		Status:    status,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Index 角色列表
func (r *RoleController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "20"))

	filters := r.buildFilters(ctx)

	roles, total, err := r.roleService.GetList(filters, page, pageSize)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      roles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 角色详情
func (r *RoleController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	role, resp := r.findRoleByID(ctx, id, true) // 预加载关联
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"role": *role,
	})
}

// Store 创建角色
func (r *RoleController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var roleCreate adminrequests.RoleCreate
	errors, err := ctx.Request().ValidateRequest(&roleCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 检查名称是否已存在
	exists, err := facades.Orm().Query().Model(&models.Role{}).Where("name", roleCreate.Name).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrRoleNameExists.Code)
	}

	// 检查标识是否已存在
	exists, err = facades.Orm().Query().Model(&models.Role{}).Where("slug", roleCreate.Slug).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrRoleSlugExists.Code)
	}

	now := carbon.Now()
	roleData := map[string]any{
		"name":        roleCreate.Name,
		"slug":        roleCreate.Slug,
		"description": roleCreate.Description,
		"status":      roleCreate.Status,
		"sort":        roleCreate.Sort,
		"created_at":  now,
		"updated_at":  now,
	}

	if err := facades.Orm().Query().Table("roles").Create(roleData); err != nil {
		return response.ErrorWithLog(ctx, "role", err, map[string]any{
			"name": roleCreate.Name,
			"slug": roleCreate.Slug,
		})
	}

	var role models.Role
	if err := facades.Orm().Query().Where("slug", roleCreate.Slug).First(&role); err != nil {
		return response.ErrorWithLog(ctx, "role", err, map[string]any{
			"slug": roleCreate.Slug,
		})
	}

	// 处理权限关联
	permissionIDs := r.roleService.ParseIDsFromRequest(ctx, "permission_ids")
	if len(permissionIDs) > 0 {
		if err := r.roleService.SyncPermissions(&role, permissionIDs); err != nil {
			return response.ErrorWithLog(ctx, "role", err, map[string]any{
				"role_id":        role.ID,
				"permission_ids": permissionIDs,
			})
		}
	}

	// 处理菜单关联
	menuIDs := r.roleService.ParseIDsFromRequest(ctx, "menu_ids")
	if len(menuIDs) > 0 {
		if err := r.roleService.SyncMenus(&role, menuIDs); err != nil {
			return response.ErrorWithLog(ctx, "role", err, map[string]any{
				"role_id":  role.ID,
				"menu_ids": menuIDs,
			})
		}
	}

	return response.Success(ctx, http.Json{
		"role": role,
	})
}

// parseProtectedRoleSlugs 解析受保护的角色标识字符串（支持逗号分隔）
func (r *RoleController) parseProtectedRoleSlugs(slugsStr string) []string {
	var slugs []string
	if slugsStr == "" {
		return slugs
	}

	parts := str.Of(slugsStr).Split(",")
	for _, part := range parts {
		part = str.Of(part).Trim().String()
		if !str.Of(part).IsEmpty() {
			slugs = append(slugs, part)
		}
	}

	return slugs
}

// isProtectedRole 检查角色是否是受保护的（通过slug判断）
func (r *RoleController) isProtectedRole(roleSlug string) bool {
	protectedSlugsStr := facades.Config().GetString("role.protected_slugs", "super-admin")
	protectedSlugs := r.parseProtectedRoleSlugs(protectedSlugsStr)
	for _, protectedSlug := range protectedSlugs {
		if roleSlug == protectedSlug {
			return true
		}
	}
	return false
}

// Update 更新角色
func (r *RoleController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	role, resp := r.findRoleByID(ctx, id, false)
	if resp != nil {
		return resp
	}

	// 检查是否是受保护的角色（通过slug判断）
	isProtected := r.isProtectedRole(role.Slug)

	// 使用请求验证
	var roleUpdate adminrequests.RoleUpdate
	errors, err := ctx.Request().ValidateRequest(&roleUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用 All() 方法检查字段是否存在
	allInputs := ctx.Request().All()

	if _, exists := allInputs["name"]; exists {
		// 检查名称是否已被其他角色使用（排除当前角色）
		exists, err := facades.Orm().Query().Model(&models.Role{}).Where("name", roleUpdate.Name).Where("id != ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrRoleNameExists.Code)
		}
		role.Name = roleUpdate.Name
	}
	if _, exists := allInputs["slug"]; exists {
		// 只有当 slug 值真正改变时才检查
		if roleUpdate.Slug != role.Slug {
			// 受保护角色的标识不能修改
			if isProtected {
				return response.Error(ctx, http.StatusForbidden, apperrors.ErrRoleProtectedCannotModifySlug.Code)
			}
			// 检查标识是否已被其他角色使用（排除当前角色）
			exists, err := facades.Orm().Query().Model(&models.Role{}).Where("slug", roleUpdate.Slug).Where("id != ?", id).Exists()
			if err != nil {
				return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
			}
			if exists {
				return response.Error(ctx, http.StatusBadRequest, apperrors.ErrRoleSlugExists.Code)
			}
			role.Slug = roleUpdate.Slug
		}
		// 如果 slug 值未改变，跳过更新（允许其他字段正常更新）
	}
	if _, exists := allInputs["description"]; exists {
		// 描述字段允许修改，包括受保护角色（如 super-admin）也可以修改描述
		role.Description = roleUpdate.Description
	}
	if _, exists := allInputs["status"]; exists {
		// 受保护角色不能禁用
		if isProtected && roleUpdate.Status == 0 {
			return response.Error(ctx, http.StatusForbidden, apperrors.ErrRoleProtectedCannotDisable.Code)
		}
		role.Status = roleUpdate.Status
	}
	if _, exists := allInputs["sort"]; exists {
		role.Sort = roleUpdate.Sort
	}

	if err := facades.Orm().Query().Save(role); err != nil {
		return response.ErrorWithLog(ctx, "role", err, map[string]any{
			"role_id": role.ID,
		})
	}

	// super-admin 角色拥有所有权限，不需要设置菜单和权限
	// 处理权限关联
	if !isProtected && ctx.Request().Input("permission_ids") != "" {
		permissionIDs := r.roleService.ParseIDsFromRequest(ctx, "permission_ids")
		if err := r.roleService.SyncPermissions(role, permissionIDs); err != nil {
			return response.ErrorWithLog(ctx, "role", err, map[string]any{
				"role_id":        role.ID,
				"permission_ids": permissionIDs,
			})
		}
	}

	// 处理菜单关联
	if !isProtected && ctx.Request().Input("menu_ids") != "" {
		menuIDs := r.roleService.ParseIDsFromRequest(ctx, "menu_ids")
		if err := r.roleService.SyncMenus(role, menuIDs); err != nil {
			return response.ErrorWithLog(ctx, "role", err, map[string]any{
				"role_id":  role.ID,
				"menu_ids": menuIDs,
			})
		}
	}

	return response.Success(ctx, http.Json{
		"role": *role,
	})
}

// Destroy 删除角色
func (r *RoleController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	role, resp := r.findRoleByID(ctx, id, false)
	if resp != nil {
		return resp
	}

	// 检查是否是受保护的角色（通过slug判断）
	if r.isProtectedRole(role.Slug) {
		return response.Error(ctx, http.StatusForbidden, apperrors.ErrRoleProtectedCannotDelete.Code)
	}

	if _, err := facades.Orm().Query().Delete(role); err != nil {
		return response.ErrorWithLog(ctx, "role", err, map[string]any{
			"role_id": role.ID,
		})
	}

	return response.Success(ctx)
}
