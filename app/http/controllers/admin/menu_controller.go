package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"
)

type MenuController struct {
	treeService services.TreeService
}

func NewMenuController() *MenuController {
	return &MenuController{
		treeService: services.NewTreeServiceImpl(),
	}
}

// Index 菜单列表（树形结构）
func (r *MenuController) Index(ctx http.Context) http.Response {
	menus, err := r.treeService.BuildMenuTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, "get_success", http.Json{
		"menus": menus,
	})
}

// Show 菜单详情
func (r *MenuController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var menu models.Menu
	if err := facades.Orm().Query().Where("id", id).First(&menu); err != nil {
		return response.Error(ctx, http.StatusNotFound, "menu_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"menu": menu,
	})
}

// Store 创建菜单
func (r *MenuController) Store(ctx http.Context) http.Response {
	parentID := cast.ToUint(ctx.Request().Input("parent_id", "0"))
	title := ctx.Request().Input("title")
	slug := ctx.Request().Input("slug")
	icon := ctx.Request().Input("icon")
	path := ctx.Request().Input("path")
	component := ctx.Request().Input("component")
	permission := ctx.Request().Input("permission")
	menuType := cast.ToUint8(ctx.Request().Input("type", "1"))
	status := cast.ToUint8(ctx.Request().Input("status", "0"))
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))
	isHidden := cast.ToUint8(ctx.Request().Input("is_hidden", "0"))

	if title == "" {
		return response.Error(ctx, http.StatusBadRequest, "menu_title_required")
	}

	if slug == "" {
		return response.Error(ctx, http.StatusBadRequest, "menu_slug_required")
	}

	// 检查标识是否已存在
	exists, err := facades.Orm().Query().Model(&models.Menu{}).Where("slug", slug).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, "menu_slug_exists")
	}

	now := carbon.Now()
	menuData := map[string]interface{}{
		"parent_id":  parentID,
		"title":      title,
		"slug":       slug,
		"icon":       icon,
		"path":       path,
		"component":  component,
		"permission": permission,
		"type":       menuType,
		"status":     status, // 明确设置 status，即使是 0 也会被保存
		"sort":       sort,
		"is_hidden":  isHidden,
		"created_at": now,
		"updated_at": now,
	}

	if err := facades.Orm().Query().Table("menus").Create(menuData); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to create menu", map[string]any{
			"error": err.Error(),
			"title": title,
			"slug":  slug,
		}, "Create menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	var menu models.Menu
	if err := facades.Orm().Query().Where("slug", slug).First(&menu); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to query created menu", map[string]any{
			"error": err.Error(),
			"slug":  slug,
		}, "Query created menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return response.Success(ctx, "create_success", http.Json{
		"menu": menu,
	})
}

// Update 更新菜单
func (r *MenuController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var menu models.Menu
	if err := facades.Orm().Query().Where("id", id).First(&menu); err != nil {
		return response.Error(ctx, http.StatusNotFound, "menu_not_found")
	}

	parentID := ctx.Request().Input("parent_id", "")
	title := ctx.Request().Input("title")
	slug := ctx.Request().Input("slug")
	icon := ctx.Request().Input("icon")
	path := ctx.Request().Input("path")
	component := ctx.Request().Input("component")
	permission := ctx.Request().Input("permission")
	menuType := ctx.Request().Input("type", "")
	status := ctx.Request().Input("status", "")
	sort := ctx.Request().Input("sort", "")
	isHidden := ctx.Request().Input("is_hidden", "")

	if title != "" {
		menu.Title = title
	}
	if slug != "" {
		// 检查标识是否已被其他菜单使用（排除当前菜单）
		exists, err := facades.Orm().Query().Model(&models.Menu{}).Where("slug", slug).Where("id != ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, "menu_slug_exists")
		}
		menu.Slug = slug
	}
	if parentID != "" {
		menu.ParentID = cast.ToUint(parentID)
	}
	menu.Icon = icon
	if path != "" {
		menu.Path = path
	}
	if component != "" {
		menu.Component = component
	}
	if permission != "" {
		menu.Permission = permission
	}
	if menuType != "" {
		menu.Type = cast.ToUint8(menuType)
	}
	if status != "" {
		menu.Status = cast.ToUint8(status)
	}
	if sort != "" {
		menu.Sort = cast.ToInt(sort)
	}
	if isHidden != "" {
		menu.IsHidden = cast.ToUint8(isHidden)
	}

	if err := facades.Orm().Query().Save(&menu); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to update menu", map[string]any{
			"error":   err.Error(),
			"menu_id": menu.ID,
		}, "Update menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"menu": menu,
	})
}

// Destroy 删除菜单
func (r *MenuController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var menu models.Menu
	if err := facades.Orm().Query().Where("id", id).First(&menu); err != nil {
		return response.Error(ctx, http.StatusNotFound, "menu_not_found")
	}

	// 检查是否有子菜单
	hasChildren, err := r.treeService.HasMenuChildren(id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}
	if hasChildren {
		return response.Error(ctx, http.StatusBadRequest, "menu_has_children")
	}

	if _, err := facades.Orm().Query().Delete(&menu); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to delete menu", map[string]any{
			"error":   err.Error(),
			"menu_id": menu.ID,
		}, "Delete menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
