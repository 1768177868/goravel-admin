package admin

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type PermissionController struct {
	treeService services.TreeService
}

func NewPermissionController() *PermissionController {
	return &PermissionController{
		treeService: services.NewTreeServiceImpl(),
	}
}

// findPermissionByID 根据ID查找权限，如果不存在则返回错误响应
// withMenu 为 true 时会预加载 Menu 关联
func (r *PermissionController) findPermissionByID(ctx http.Context, id uint, withMenu bool) (*models.Permission, http.Response) {
	var relations []string
	if withMenu {
		relations = append(relations, "Menu")
	}
	return response.FindByID[models.Permission](ctx, id, &response.FindByIDOptions{
		WithRelations:      relations,
		NotFoundMessageKey: "permission_not_found",
	})
}

// buildQuery 构建权限查询
func (r *PermissionController) buildQuery(ctx http.Context) orm.Query {
	name := ctx.Request().Query("name", "")
	slug := ctx.Request().Query("slug", "")
	method := ctx.Request().Query("method", "")
	path := ctx.Request().Query("path", "")
	status := ctx.Request().Query("status", "")
	menuID := ctx.Request().Query("menu_id", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

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
	if status != "" {
		query = query.Where("status", status)
	}
	if menuID != "" {
		// 获取菜单及其所有子菜单的ID列表
		menuIDs, err := r.treeService.GetMenuChildrenIDs(cast.ToUint(menuID))
		if err != nil {
			// 如果获取菜单ID失败，返回空查询
			return query.Where("1 = 0")
		}
		// 使用 IN 查询，查询该菜单及其所有子菜单的权限
		if len(menuIDs) > 0 {
			idsAny := helpers.ConvertUintSliceToAny(menuIDs)
			query = query.WhereIn("menu_id", idsAny)
		}
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 sort asc, id desc
	query = helpers.ApplySort(query, orderBy, "sort:asc,id:desc")

	return query
}

// Index 权限列表
func (r *PermissionController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var permissions []models.Permission
	return response.PaginateQuery(ctx, query, &permissions, &response.PaginateQueryOptions{
		WithRelations: []string{"Menu"},
	})
}

// Show 权限详情
func (r *PermissionController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
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
	status := cast.ToUint8(ctx.Request().Input("status", "0"))
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))
	menuID := cast.ToUint(ctx.Request().Input("menu_id", "0"))

	if name == "" || slug == "" {
		return response.Error(ctx, http.StatusBadRequest, "permission_name_and_slug_required")
	}

	exists, err := facades.Orm().Query().Model(&models.Permission{}).
		Where("name", name).
		OrWhere("slug", slug).
		Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, "permission_name_or_slug_exists")
	}

	now := carbon.Now()
	permissionData := map[string]any{
		"name":        name,
		"slug":        slug,
		"method":      method,
		"path":        path,
		"description": description,
		"status":      status,
		"sort":        sort,
		"menu_id":     menuID,
		"created_at":  now,
		"updated_at":  now,
	}

	if err := facades.Orm().Query().Table("permissions").Create(permissionData); err != nil {
		return response.ErrorWithLog(ctx, "permission", err, map[string]any{
			"name": name,
			"slug": slug,
		})
	}

	var permission models.Permission
	if err := facades.Orm().Query().Where("slug", slug).First(&permission); err != nil {
		return response.ErrorWithLog(ctx, "permission", err, map[string]any{
			"slug": slug,
		})
	}

	return response.Success(ctx, http.Json{
		"permission": permission,
	})
}

func (r *PermissionController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
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
		exists, err := facades.Orm().Query().Model(&models.Permission{}).Where("name", name).Where("id <> ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, "permission_name_exists")
		}
		permission.Name = name
	}
	if slug != "" {
		exists, err := facades.Orm().Query().Model(&models.Permission{}).Where("slug", slug).Where("id <> ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
		if exists {
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
	if menuIDStr != "" {
		permission.MenuID = cast.ToUint(menuIDStr)
	}

	if err := facades.Orm().Query().Save(permission); err != nil {
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
	id := cast.ToUint(ctx.Request().Route("id"))
	permission, resp := r.findPermissionByID(ctx, id, false) // 不需要预加载关联
	if resp != nil {
		return resp
	}

	if _, err := facades.Orm().Query().Delete(permission); err != nil {
		return response.ErrorWithLog(ctx, "permission", err, map[string]any{
			"permission_id": permission.ID,
		})
	}

	return response.Success(ctx)
}
