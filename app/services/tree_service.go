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
