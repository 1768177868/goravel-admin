package services

import (
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/models"
)

type MenuService interface {
	// GetByID 根据ID获取菜单
	GetByID(id uint) (*models.Menu, error)
	// Create 创建菜单
	Create(parentID uint, title, slug, icon, path, component, permission string, menuType uint8, status uint8, sort int, isHidden uint8, linkType, openType uint8) (*models.Menu, error)
	// Update 更新菜单
	Update(menu *models.Menu) error
	// Delete 删除菜单
	Delete(menu *models.Menu) error
}

type MenuServiceImpl struct {
}

func NewMenuService() MenuService {
	return &MenuServiceImpl{}
}

// GetByID 根据ID获取菜单
func (s *MenuServiceImpl) GetByID(id uint) (*models.Menu, error) {
	var menu models.Menu
	if err := facades.Orm().Query().Where("id", id).FirstOrFail(&menu); err != nil {
		return nil, apperrors.ErrMenuNotFound.WithError(err)
	}
	return &menu, nil
}

// Create 创建菜单
func (s *MenuServiceImpl) Create(parentID uint, title, slug, icon, path, component, permission string, menuType uint8, status uint8, sort int, isHidden uint8, linkType, openType uint8) (*models.Menu, error) {
	menu := &models.Menu{}
	createData := map[string]any{
		"parent_id":  parentID,
		"title":      title,
		"slug":       slug,
		"icon":       icon,
		"path":       path,
		"component":  component,
		"permission": permission,
		"type":       menuType,
		"status":     status,
		"sort":       sort,
		"is_hidden":  isHidden,
		"link_type":  linkType,
		"open_type":  openType,
	}

	if err := facades.Orm().Query().Model(menu).Create(createData); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return menu, nil
}

// Update 更新菜单
func (s *MenuServiceImpl) Update(menu *models.Menu) error {
	if err := facades.Orm().Query().Save(menu); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除菜单
func (s *MenuServiceImpl) Delete(menu *models.Menu) error {
	if _, err := facades.Orm().Query().Delete(menu); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
