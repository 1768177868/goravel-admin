package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	adminrequests "goravel/app/http/requests/admin"
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

// findMenuByID 根据ID查找菜单，如果不存在则返回错误响应
func (r *MenuController) findMenuByID(ctx http.Context, id uint) (*models.Menu, http.Response) {
	if id == 0 {
		return nil, response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var menu models.Menu
	if err := facades.Orm().Query().Where("id", id).First(&menu); err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, "menu_not_found")
	}

	if menu.ID == 0 {
		return nil, response.Error(ctx, http.StatusNotFound, "menu_not_found")
	}

	return &menu, nil
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
	menu, resp := r.findMenuByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, "get_success", http.Json{
		"menu": *menu,
	})
}

// Store 创建菜单
func (r *MenuController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var menuCreate adminrequests.MenuCreate
	errors, err := ctx.Request().ValidateRequest(&menuCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 检查 slug 是否已存在
	exists, err := facades.Orm().Query().Model(&models.Menu{}).Where("slug", menuCreate.Slug).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, "menu_slug_exists")
	}

	now := carbon.Now()
	menuData := map[string]any{
		"parent_id":  menuCreate.ParentID,
		"title":      menuCreate.Title,
		"slug":       menuCreate.Slug,
		"icon":       menuCreate.Icon,
		"path":       menuCreate.Path,
		"component":  menuCreate.Component,
		"permission": menuCreate.Permission,
		"type":       menuCreate.Type,
		"status":     menuCreate.Status,
		"sort":       menuCreate.Sort,
		"is_hidden":  menuCreate.IsHidden,
		"created_at": now,
		"updated_at": now,
	}

	if err := facades.Orm().Query().Table("menus").Create(menuData); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to create menu", map[string]any{
			"error": err.Error(),
			"title": menuCreate.Title,
			"slug":  menuCreate.Slug,
		}, "Create menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	var menu models.Menu
	if err := facades.Orm().Query().Where("slug", menuCreate.Slug).First(&menu); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to query created menu", map[string]any{
			"error": err.Error(),
			"slug":  menuCreate.Slug,
		}, "Query created menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return response.Success(ctx, "create_success", http.Json{
		"menu": menu,
	})
}

func (r *MenuController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	menu, resp := r.findMenuByID(ctx, id)
	if resp != nil {
		return resp
	}

	// 使用请求验证
	var menuUpdate adminrequests.MenuUpdate
	errors, err := ctx.Request().ValidateRequest(&menuUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用 All() 方法检查字段是否存在
	allInputs := ctx.Request().All()

	if _, exists := allInputs["title"]; exists {
		menu.Title = menuUpdate.Title
	}
	if _, exists := allInputs["slug"]; exists {
		// 检查 slug 是否已被其他菜单使用
		exists, err := facades.Orm().Query().Model(&models.Menu{}).Where("slug", menuUpdate.Slug).Where("id != ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, "menu_slug_exists")
		}
		menu.Slug = menuUpdate.Slug
	}
	if _, exists := allInputs["parent_id"]; exists {
		menu.ParentID = menuUpdate.ParentID
	}
	if _, exists := allInputs["icon"]; exists {
		menu.Icon = menuUpdate.Icon
	}
	if _, exists := allInputs["path"]; exists {
		menu.Path = menuUpdate.Path
	}
	if _, exists := allInputs["component"]; exists {
		menu.Component = menuUpdate.Component
	}
	if _, exists := allInputs["permission"]; exists {
		menu.Permission = menuUpdate.Permission
	}
	if _, exists := allInputs["type"]; exists {
		menu.Type = menuUpdate.Type
	}
	if _, exists := allInputs["status"]; exists {
		menu.Status = menuUpdate.Status
	}
	if _, exists := allInputs["sort"]; exists {
		menu.Sort = menuUpdate.Sort
	}
	if _, exists := allInputs["is_hidden"]; exists {
		menu.IsHidden = menuUpdate.IsHidden
	}

	if err := facades.Orm().Query().Save(menu); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to update menu", map[string]any{
			"error":   err.Error(),
			"menu_id": menu.ID,
		}, "Update menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"menu": *menu,
	})
}

func (r *MenuController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	menu, resp := r.findMenuByID(ctx, id)
	if resp != nil {
		return resp
	}

	hasChildren, err := r.treeService.HasMenuChildren(id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}
	if hasChildren {
		return response.Error(ctx, http.StatusBadRequest, "menu_has_children")
	}

	if _, err := facades.Orm().Query().Delete(menu); err != nil {
		errorlog.RecordHTTP(ctx, "menu", "Failed to delete menu", map[string]any{
			"error":   err.Error(),
			"menu_id": menu.ID,
		}, "Delete menu error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
