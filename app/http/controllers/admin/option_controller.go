package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
)

type OptionController struct {
	treeService services.TreeService
}

func NewOptionController() *OptionController {
	return &OptionController{
		treeService: services.NewTreeServiceImpl(),
	}
}

func (r *OptionController) Index(ctx http.Context) http.Response {
	optionType := ctx.Request().Query("type", "")

	switch optionType {
	case "role":
		return r.getRoleOptions(ctx)
	case "department":
		return r.getDepartmentOptions(ctx)
	case "menu":
		return r.getMenuOptions(ctx)
	case "status":
		return r.getStatusOptions(ctx)
	case "method":
		return r.getMethodOptions(ctx)
	case "yes_no":
		return r.getYesNoOptions(ctx)
	default:
		return response.Error(ctx, http.StatusBadRequest, "invalid_option_type")
	}
}

func (r *OptionController) getRoleOptions(ctx http.Context) http.Response {
	var roles []models.Role
	if err := facades.Orm().Query().Where("status", 1).Order("id asc").Get(&roles); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var options []map[string]any
	for _, role := range roles {
		options = append(options, map[string]any{
			"label": role.Name,
			"value": cast.ToString(role.ID),
		})
	}

	return response.Success(ctx, "get_success", http.Json{
		"options": options,
	})
}

func (r *OptionController) getDepartmentOptions(ctx http.Context) http.Response {
	var departments []models.Department
	if err := facades.Orm().Query().Where("status", 1).Order("sort asc, id asc").Get(&departments); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	tree := r.buildDepartmentTree(departments, 0)

	return response.Success(ctx, "get_success", http.Json{
		"options": tree,
		"list":    departments,
	})
}

func (r *OptionController) buildDepartmentTree(departments []models.Department, parentID uint) []map[string]any {
	var tree []map[string]any
	for _, dept := range departments {
		if dept.ParentID == parentID {
			node := map[string]any{
				"id":   dept.ID,
				"name": dept.Name,
			}
			children := r.buildDepartmentTree(departments, dept.ID)
			if len(children) > 0 {
				node["children"] = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}

func (r *OptionController) getMenuOptions(ctx http.Context) http.Response {
	menus, err := r.treeService.BuildMenuTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	tree := r.buildMenuTree(menus)

	return response.Success(ctx, "get_success", http.Json{
		"options": tree,
	})
}

func (r *OptionController) buildMenuTree(menus []models.Menu) []map[string]any {
	var tree []map[string]any
	for _, menu := range menus {
		// 使用菜单标题和路径构建显示标签
		label := menu.Title
		if menu.Path != "" {
			label = label + " (" + menu.Path + ")"
		}
		
		node := map[string]any{
			"id":    menu.ID,
			"name":  menu.Title,
			"label": label,
			"value": menu.ID,
		}
		
		if len(menu.Children) > 0 {
			node["children"] = r.buildMenuTree(menu.Children)
		}
		
		tree = append(tree, node)
	}
	return tree
}

func (r *OptionController) getStatusOptions(ctx http.Context) http.Response {
	options := []map[string]any{
		{"label": trans.Get(ctx, "common.enabled"), "value": "1"},
		{"label": trans.Get(ctx, "common.disabled"), "value": "0"},
	}

	return response.Success(ctx, "get_success", http.Json{
		"options": options,
	})
}

func (r *OptionController) getMethodOptions(ctx http.Context) http.Response {
	options := []map[string]any{
		{"label": "GET", "value": "GET"},
		{"label": "POST", "value": "POST"},
		{"label": "PUT", "value": "PUT"},
		{"label": "DELETE", "value": "DELETE"},
		{"label": "PATCH", "value": "PATCH"},
	}

	return response.Success(ctx, "get_success", http.Json{
		"options": options,
	})
}

func (r *OptionController) getYesNoOptions(ctx http.Context) http.Response {
	options := []map[string]any{
		{"label": trans.Get(ctx, "common.yes"), "value": "1"},
		{"label": trans.Get(ctx, "common.no"), "value": "0"},
	}

	return response.Success(ctx, "get_success", http.Json{
		"options": options,
	})
}
