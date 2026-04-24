package models

import (
	"github.com/goravel/framework/database/orm"
)

type Department struct {
	orm.Model
	ParentID uint         `gorm:"index;default:0;comment:父级ID"`
	Name     string       `gorm:"not null;size:50;comment:部门名称"`
	Code     string       `gorm:"size:50;comment:部门编码"`
	Leader   string       `gorm:"size:50;comment:负责人"`
	Phone    string       `gorm:"size:20;comment:联系电话"`
	Email    string       `gorm:"size:100;comment:邮箱"`
	Status   uint8        `gorm:"default:1;comment:状态 1:启用 0:禁用"`
	Sort     int          `gorm:"default:0;comment:排序"`
	Remark   string       `gorm:"size:500;comment:备注"`
	Children []Department `gorm:"foreignKey:ParentID"`
	Admins   []Admin      `gorm:"foreignKey:DepartmentID"`
}
