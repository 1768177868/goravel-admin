package admin

import (
	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

type MenuController struct {}


func NewMenuController() *MenuController {
	return &MenuController{}
}

func (r *MenuController) treeService(ctx http.Context) services.TreeService {
	return services.NewTreeServiceImpl(ctx)
}

func (r *MenuController) menuService(ctx http.Context) services.MenuService {
	return services.NewMenuService(ctx)
}


// findMenuByID 根据ID查找菜单，如果不存在则返回错误响应
func (r *MenuController) findMenuByID(ctx http.Context, id uint) (*models.Menu, http.Response) {
	return response.FindByID[models.Menu](ctx, id, &response.FindByIDOptions{
		NotFoundMessageKey: apperrors.ErrMenuNotFound.Code,
	})
}

// buildMenuTreeResponse 构建菜单树并转为前端格式（供 Index 与 Tree 复用）
func (r *MenuController) buildMenuTreeResponse(ctx http.Context, menus []models.Menu) http.Response {
	treeData := utils.ConvertMenuTree(menus)
	return response.Success(ctx, http.Json{
		"menus": treeData,
		"list":  treeData, // 兼容前端可能使用的 list 字段
	})
}

// Tree 菜单树（仅登录即可访问，不校验菜单权限；用于角色/权限表单等）
func (r *MenuController) Tree(ctx http.Context) http.Response {
	menus, err := r.treeService(ctx).BuildMenuTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	menus = r.applyMenuTreeFilters(ctx, menus)
	return r.buildMenuTreeResponse(ctx, menus)
}

// applyMenuTreeFilters 应用监控/开发工具等过滤
func (r *MenuController) applyMenuTreeFilters(ctx http.Context, menus []models.Menu) []models.Menu {
	adminID := r.extractAdminID(ctx)
	developerIDsStr := facades.Config().GetString("admin.developer_ids", "2")
	isDeveloper := r.isDeveloperAdmin(adminID, developerIDsStr)
	appEnv := str.Of(facades.Config().GetString("app.env", "production")).Lower().Trim().String()
	isLocalOrDevelopment := appEnv == "local" || appEnv == "development"

	monitorHidden := facades.Config().GetString("admin.monitor_hidden", "")
	if monitorHidden != "" && monitorHidden != "0" {
		if !isDeveloper {
			menus = r.filterMonitorMenu(menus)
		}
	}
	if !isLocalOrDevelopment {
		menus = r.filterDevMenu(menus)
	}
	return menus
}

func (r *MenuController) extractAdminID(ctx http.Context) uint {
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return 0
	}

	if admin, ok := adminValue.(models.Admin); ok {
		return admin.ID
	}
	if adminPtr, ok := adminValue.(*models.Admin); ok && adminPtr != nil {
		return adminPtr.ID
	}

	return 0
}

// Index 菜单列表（树形结构，需要菜单权限）
func (r *MenuController) Index(ctx http.Context) http.Response {
	menus, err := r.treeService(ctx).BuildMenuTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	menus = r.applyMenuTreeFilters(ctx, menus)
	return r.buildMenuTreeResponse(ctx, menus)
}

// Show 菜单详情
func (r *MenuController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
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
	exists, err := appfacades.OrmQuery(ctx).Model(&models.Menu{}).Where("slug", menuCreate.Slug).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrMenuSlugExists.Code)
	}

	menu, err := r.menuService(ctx).Create(
		menuCreate.ParentID,
		menuCreate.Title,
		menuCreate.Slug,
		menuCreate.Icon,
		menuCreate.Path,
		menuCreate.Component,
		menuCreate.Permission,
		menuCreate.Type,
		menuCreate.Status,
		menuCreate.Sort,
		menuCreate.IsHidden,
		menuCreate.LinkType,
		menuCreate.OpenType,
		menuCreate.NoCache,
	)
	if err != nil {
		return response.ErrorWithLog(ctx, "menu", err, map[string]any{
			"title": menuCreate.Title,
			"slug":  menuCreate.Slug,
		})
	}

	return response.Success(ctx, http.Json{
		"menu": *menu,
	})
}

func (r *MenuController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
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
		exists, err := appfacades.OrmQuery(ctx).Model(&models.Menu{}).Where("slug", menuUpdate.Slug).Where("id != ?", id).Exists()
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
		}
		if exists {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrMenuSlugExists.Code)
		}
		menu.Slug = menuUpdate.Slug
	}
	// 处理 parent_id，需要根据是否存在来更新，如果存在但为空或0，则设为0
	if val, exists := allInputs["parent_id"]; exists {
		if val == nil {
			menu.ParentID = 0
		} else {
			menu.ParentID = menuUpdate.ParentID
		}
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
	if _, exists := allInputs["no_cache"]; exists {
		menu.NoCache = menuUpdate.NoCache
	}

	if err := r.menuService(ctx).Update(menu); err != nil {
		return response.ErrorWithLog(ctx, "menu", err, map[string]any{
			"menu_id": menu.ID,
		})
	}
	return response.Success(ctx, http.Json{
		"menu": *menu,
	})
}

func (r *MenuController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	menu, resp := r.findMenuByID(ctx, id)
	if resp != nil {
		return resp
	}

	hasChildren, err := r.treeService(ctx).HasMenuChildren(id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	if hasChildren {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrMenuHasChildren.Code)
	}
	if err := r.menuService(ctx).Delete(menu); err != nil {
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
		// 监控中心及其子菜单整体隐藏
		if menu.Slug != "monitor" && menu.Slug != "monitor-center" {
			// 递归过滤子菜单
			if len(menu.Children) > 0 {
				menu.Children = r.filterMonitorMenu(menu.Children)
			}
			filteredMenus = append(filteredMenus, menu)
		}
	}
	return filteredMenus
}

// filterDevMenu 递归过滤掉开发工具菜单
func (r *MenuController) filterDevMenu(menus []models.Menu) []models.Menu {
	var filteredMenus []models.Menu
	for _, menu := range menus {
		// 如果当前菜单不是开发工具菜单，则保留
		if menu.Slug != "dev" {
			// 递归过滤子菜单
			if len(menu.Children) > 0 {
				menu.Children = r.filterDevMenu(menu.Children)
			}
			filteredMenus = append(filteredMenus, menu)
		}
	}
	return filteredMenus
}
