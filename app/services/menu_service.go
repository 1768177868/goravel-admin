package services

import (
	"context"
	appfacades "goravel/app/facades"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/utils"
)

type MenuService interface {
	// GetByID 根据ID获取菜单
	GetByID(id uint) (*models.Menu, error)
	// Create 创建菜单
	Create(parentID uint, title, slug, icon, path, component, permission string, menuType uint8, status uint8, sort int, isHidden uint8, linkType, openType, noCache uint8) (*models.Menu, error)
	// ValidateSlugUnique 校验菜单 slug 唯一性（硬删后可复用）。
	ValidateSlugUnique(slug string, excludeID uint) error
	// Update 更新菜单
	Update(menu *models.Menu) error
	// Delete 删除菜单
	Delete(menu *models.Menu) error
}

type MenuServiceImpl struct {
	ctx context.Context
}

func NewMenuService(ctx context.Context) MenuService {
	return &MenuServiceImpl{ctx: ctx}
}

// GetByID 根据ID获取菜单
func (s *MenuServiceImpl) GetByID(id uint) (*models.Menu, error) {
	var menu models.Menu
	if err := appfacades.OrmQuery(s.ctx).Where("id", id).FirstOrFail(&menu); err != nil {
		return nil, apperrors.ErrMenuNotFound.WithError(err)
	}
	return &menu, nil
}

// ValidateSlugUnique 校验菜单 slug 唯一性（硬删后可复用）。
func (s *MenuServiceImpl) ValidateSlugUnique(slug string, excludeID uint) error {
	if slug == "" {
		return nil
	}
	exists, err := utils.ExistsColumnValue(s.ctx, "menus", &models.Menu{}, utils.UniqueReuseAllow, "slug", slug, excludeID)
	if err != nil {
		return apperrors.ErrCreateFailed.WithError(err)
	}
	if exists {
		return apperrors.ErrMenuSlugExists
	}
	return nil
}

// Create 创建菜单
func (s *MenuServiceImpl) Create(parentID uint, title, slug, icon, path, component, permission string, menuType uint8, status uint8, sort int, isHidden uint8, linkType, openType, noCache uint8) (*models.Menu, error) {
	if err := s.ValidateSlugUnique(slug, 0); err != nil {
		return nil, err
	}

	menu := &models.Menu{
		ParentID:   parentID,
		Title:      title,
		Slug:       slug,
		Icon:       icon,
		Path:       path,
		Component:  component,
		Permission: permission,
		Type:       menuType,
		Status:     status,
		Sort:       sort,
		IsHidden:   isHidden,
		LinkType:   linkType,
		OpenType:   openType,
		NoCache:    noCache,
	}

	if err := appfacades.OrmQuery(s.ctx).Create(menu); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return menu, nil
}

// Update 更新菜单
func (s *MenuServiceImpl) Update(menu *models.Menu) error {
	if err := s.ValidateSlugUnique(menu.Slug, menu.ID); err != nil {
		return err
	}
	if err := appfacades.OrmQuery(s.ctx).Save(menu); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除菜单
func (s *MenuServiceImpl) Delete(menu *models.Menu) error {
	if _, err := appfacades.OrmQuery(s.ctx).Delete(menu); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
