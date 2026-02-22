package services

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type TreeService interface {
	// BuildMenuTree 构建菜单树形结构
	BuildMenuTree(parentID uint) ([]models.Menu, error)
	// BuildDepartmentTree 构建部门树形结构
	BuildDepartmentTree(parentID uint) ([]models.Department, error)
	// HasMenuChildren 检查菜单是否有子节点
	HasMenuChildren(menuID uint) (bool, error)
	// HasDepartmentChildren 检查部门是否有子节点
	HasDepartmentChildren(departmentID uint) (bool, error)
	// GetMenuChildrenIDs 获取菜单及其所有子菜单的ID列表
	GetMenuChildrenIDs(menuID uint) ([]uint, error)
	// GetMenuIDsWithAncestors 给定菜单ID列表，返回这些ID及其所有祖先菜单ID（用于在只勾选权限时也能显示父级目录）
	GetMenuIDsWithAncestors(ids []uint) ([]uint, error)
}

type TreeServiceImpl struct {
}

func NewTreeServiceImpl() *TreeServiceImpl {
	return &TreeServiceImpl{}
}

// BuildMenuTree 构建菜单树形结构
func (s *TreeServiceImpl) BuildMenuTree(parentID uint) ([]models.Menu, error) {
	var menus []models.Menu
	if err := facades.Orm().Query().Where("parent_id", parentID).Order("sort asc, id asc").Get(&menus); err != nil {
		return nil, err
	}

	// 递归加载子菜单
	for i := range menus {
		children, err := s.BuildMenuTree(menus[i].ID)
		if err != nil {
			return nil, err
		}
		menus[i].Children = children
	}

	return menus, nil
}

// BuildDepartmentTree 构建部门树形结构
func (s *TreeServiceImpl) BuildDepartmentTree(parentID uint) ([]models.Department, error) {
	var departments []models.Department
	if err := facades.Orm().Query().Where("parent_id", parentID).Order("sort asc, id asc").Get(&departments); err != nil {
		return nil, err
	}

	// 递归加载子部门
	for i := range departments {
		children, err := s.BuildDepartmentTree(departments[i].ID)
		if err != nil {
			return nil, err
		}
		departments[i].Children = children
	}

	return departments, nil
}

// HasMenuChildren 检查菜单是否有子节点
func (s *TreeServiceImpl) HasMenuChildren(menuID uint) (bool, error) {
	count, err := facades.Orm().Query().Model(&models.Menu{}).Where("parent_id", menuID).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasDepartmentChildren 检查部门是否有子节点
func (s *TreeServiceImpl) HasDepartmentChildren(departmentID uint) (bool, error) {
	count, err := facades.Orm().Query().Model(&models.Department{}).Where("parent_id", departmentID).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMenuChildrenIDs 获取菜单及其所有子菜单的ID列表（递归）
func (s *TreeServiceImpl) GetMenuChildrenIDs(menuID uint) ([]uint, error) {
	var menuIDs []uint
	menuIDs = append(menuIDs, menuID)

	// 递归获取所有子菜单ID
	var getChildren func(parentID uint) error
	getChildren = func(parentID uint) error {
		var children []models.Menu
		if err := facades.Orm().Query().Model(&models.Menu{}).Where("parent_id", parentID).Select("id").Get(&children); err != nil {
			return err
		}

		for _, child := range children {
			menuIDs = append(menuIDs, child.ID)
			// 递归获取子菜单的子菜单
			if err := getChildren(child.ID); err != nil {
				return err
			}
		}

		return nil
	}

	if err := getChildren(menuID); err != nil {
		return nil, err
	}

	return menuIDs, nil
}

// GetMenuIDsWithAncestors 给定菜单ID列表，返回这些ID及其所有祖先菜单ID（用于在只勾选权限时也能显示父级目录）
func (s *TreeServiceImpl) GetMenuIDsWithAncestors(ids []uint) ([]uint, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	result := make(map[uint]bool)
	for _, id := range ids {
		result[id] = true
	}
	current := ids
	for len(current) > 0 {
		var rows []struct {
			ID       uint
			ParentID uint
		}
		if err := facades.Orm().Query().Model(&models.Menu{}).Where("id IN ?", current).Select("id", "parent_id").Get(&rows); err != nil {
			return nil, err
		}
		var next []uint
		for _, row := range rows {
			if row.ParentID != 0 && !result[row.ParentID] {
				result[row.ParentID] = true
				next = append(next, row.ParentID)
			}
		}
		current = next
	}
	out := make([]uint, 0, len(result))
	for id := range result {
		out = append(out, id)
	}
	return out, nil
}
