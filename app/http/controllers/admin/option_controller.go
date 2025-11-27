package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
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

// Index 获取下拉选项数据（统一接口）
// 支持多种类型：role, department, status, method 等
func (r *OptionController) Index(ctx http.Context) http.Response {
	optionType := ctx.Request().Query("type", "")
	
	switch optionType {
	case "role":
		return r.getRoleOptions(ctx)
	case "department":
		return r.getDepartmentOptions(ctx)
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

// getRoleOptions 获取角色选项
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

// getDepartmentOptions 获取部门选项（树形结构）
func (r *OptionController) getDepartmentOptions(ctx http.Context) http.Response {
	// 获取所有部门
	var departments []models.Department
	if err := facades.Orm().Query().Where("status", 1).Order("sort asc, id asc").Get(&departments); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	// 转换为树形结构
	tree := r.buildDepartmentTree(departments, 0)

	return response.Success(ctx, "get_success", http.Json{
		"options": tree, // 树形结构，用于 tree-select
		"list":    departments, // 扁平列表，用于 select
	})
}

// buildDepartmentTree 构建部门树形结构
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

// getStatusOptions 获取状态选项
func (r *OptionController) getStatusOptions(ctx http.Context) http.Response {
	options := []map[string]any{
		{"label": "启用", "value": "1"},
		{"label": "禁用", "value": "0"},
	}

	return response.Success(ctx, "get_success", http.Json{
		"options": options,
	})
}

// getMethodOptions 获取 HTTP 方法选项
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

// getYesNoOptions 获取是/否选项
func (r *OptionController) getYesNoOptions(ctx http.Context) http.Response {
	options := []map[string]any{
		{"label": "是", "value": "1"},
		{"label": "否", "value": "0"},
	}

	return response.Success(ctx, "get_success", http.Json{
		"options": options,
	})
}

