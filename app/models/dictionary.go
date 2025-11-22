package models

import (
	"github.com/goravel/framework/database/orm"
)

type Dictionary struct {
	orm.Model
	Type        string `gorm:"index;not null;size:50;comment:字典类型"`
	Label       string `gorm:"not null;size:50;comment:字典标签"`
	Value       string `gorm:"not null;size:100;comment:字典值"`
	Description string `gorm:"size:255;comment:字典描述"`
	Status      uint8  `gorm:"default:1;comment:状态 1:启用 0:禁用"`
	Sort        int    `gorm:"default:0;comment:排序"`
	Remark      string `gorm:"size:500;comment:备注"`
}
