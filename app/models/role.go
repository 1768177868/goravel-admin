package models

import (
	"github.com/goravel/framework/database/orm"
)

type Role struct {
	orm.Model
	Name        string       `gorm:"uniqueIndex;not null;size:50;comment:角色名称"`
	Slug        string       `gorm:"uniqueIndex;not null;size:50;comment:角色标识"`
	Description string       `gorm:"size:255;comment:角色描述"`
	Status      uint8        `gorm:"default:1;comment:状态 1:启用 0:禁用"`
	Sort        int          `gorm:"default:0;comment:排序"`
	Admins      []Admin      `gorm:"many2many:admin_role;comment:管理员"`
	Permissions []Permission `gorm:"many2many:role_permission;comment:权限"`
	Menus       []Menu       `gorm:"many2many:role_menu;comment:菜单"`
}
