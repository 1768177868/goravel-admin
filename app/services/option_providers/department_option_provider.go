package option_providers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type DepartmentOptionProvider struct{}

func NewDepartmentOptionProvider() *DepartmentOptionProvider {
	return &DepartmentOptionProvider{}
}

func (p *DepartmentOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	var departments []models.Department
	if err := facades.Orm().Query().Where("status", 1).Order("sort asc, id asc").Get(&departments); err != nil {
		return nil, err
	}

	tree := p.buildDepartmentTree(departments, 0)

	return map[string]any{
		"options": tree,
		"list":    departments,
	}, nil
}

func (p *DepartmentOptionProvider) buildDepartmentTree(departments []models.Department, parentID uint) []map[string]any {
	var tree []map[string]any
	for _, dept := range departments {
		if dept.ParentID == parentID {
			node := map[string]any{
				"id":   dept.ID,
				"name": dept.Name,
			}
			children := p.buildDepartmentTree(departments, dept.ID)
			if len(children) > 0 {
				node["children"] = children
			}
			tree = append(tree, node)
		}
	}
	return tree
}

