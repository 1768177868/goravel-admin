package option_providers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/models"
	"goravel/app/services"
)

type MenuOptionProvider struct {
	treeService services.TreeService
}

func NewMenuOptionProvider(treeService services.TreeService) *MenuOptionProvider {
	return &MenuOptionProvider{
		treeService: treeService,
	}
}

func (p *MenuOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	menus, err := p.treeService.BuildMenuTree(0)
	if err != nil {
		return nil, err
	}

	tree := p.buildMenuTree(menus)

	return map[string]any{
		"options": tree,
	}, nil
}

func (p *MenuOptionProvider) buildMenuTree(menus []models.Menu) []map[string]any {
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
			node["children"] = p.buildMenuTree(menu.Children)
		}

		tree = append(tree, node)
	}
	return tree
}

