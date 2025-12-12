package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/cast"

	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
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
	return response.FindByID[models.Menu](ctx, id, &response.FindByIDOptions{
		NotFoundMessageKey: "menu_not_found",
	})
}

// Index 菜单列表（树形结构）
func (r *MenuController) Index(ctx http.Context) http.Response {
	menus, err := r.treeService.BuildMenuTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	// 检查是否需要隐藏服务监控菜单
	// 只有当配置值不为空且不等于 "0" 时才隐藏（"0" 表示不隐藏）
	monitorHidden := facades.Config().GetString("admin.monitor_hidden", "")
	if monitorHidden != "" && monitorHidden != "0" {
		// 获取当前管理员ID
		adminValue := ctx.Value("admin")
		var adminID uint
		if adminValue != nil {
			if admin, ok := adminValue.(models.Admin); ok {
				adminID = admin.ID
			} else if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
				adminID = adminPtr.ID
			}
		}

		// 检查是否是开发者管理员
		developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
		isDeveloperAdmin := r.isDeveloperAdmin(adminID, developerIDsStr)

		// 如果不是开发者管理员，则过滤掉服务监控菜单
		if !isDeveloperAdmin {
			menus = r.filterMonitorMenu(menus)
		}
	}

	return response.Success(ctx, http.Json{
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

	return response.Success(ctx, http.Json{
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
		"link_type":  menuCreate.LinkType,
		"open_type":  menuCreate.OpenType,
		"created_at": now,
		"updated_at": now,
	}

	if err := facades.Orm().Query().Table("menus").Create(menuData); err != nil {
		return response.ErrorWithLog(ctx, "menu", err, map[string]any{
			"title": menuCreate.Title,
			"slug":  menuCreate.Slug,
		})
	}

	var menu models.Menu
	if err := facades.Orm().Query().Where("slug", menuCreate.Slug).First(&menu); err != nil {
		return response.ErrorWithLog(ctx, "menu", err, map[string]any{
			"slug": menuCreate.Slug,
		})
	}

	return response.Success(ctx, http.Json{
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
	if _, exists := allInputs["link_type"]; exists {
		menu.LinkType = menuUpdate.LinkType
	}
	if _, exists := allInputs["open_type"]; exists {
		menu.OpenType = menuUpdate.OpenType
	}

	if err := facades.Orm().Query().Save(menu); err != nil {
		return response.ErrorWithLog(ctx, "menu", err, map[string]any{
			"menu_id": menu.ID,
		})
	}

	return response.Success(ctx, http.Json{
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
		return response.ErrorWithLog(ctx, "menu", err, map[string]any{
			"menu_id": menu.ID,
		})
	}

	return response.Success(ctx)
}

// isDeveloperAdmin 检查是否是开发者管理员
func (r *MenuController) isDeveloperAdmin(adminID uint, developerIDsStr string) bool {
	if developerIDsStr == "" {
		return false
	}

	// 解析开发者ID列表
	parts := str.Of(developerIDsStr).Split(",")
	for _, part := range parts {
		part = str.Of(part).Trim().String()
		if !str.Of(part).IsEmpty() {
			if id := cast.ToUint(part); id > 0 && id == adminID {
				return true
			}
		}
	}

	return false
}

// filterMonitorMenu 递归过滤掉服务监控菜单
func (r *MenuController) filterMonitorMenu(menus []models.Menu) []models.Menu {
	var filteredMenus []models.Menu
	for _, menu := range menus {
		// 如果当前菜单不是服务监控菜单，则保留
		if menu.Slug != "monitor" {
			// 递归过滤子菜单
			if len(menu.Children) > 0 {
				menu.Children = r.filterMonitorMenu(menu.Children)
			}
			filteredMenus = append(filteredMenus, menu)
		}
	}
	return filteredMenus
}
