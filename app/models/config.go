package models

import (
	"github.com/goravel/framework/database/orm"
)

type Config struct {
	orm.Model
	Group string `gorm:"index;not null;size:50;comment:配置分组 website:网站配置 email:邮箱配置"`
	Key   string `gorm:"index;not null;size:100;comment:配置键"`
	Value string `gorm:"type:text;comment:配置值"`
	Label string `gorm:"size:100;comment:配置标签"`
	Type  string `gorm:"size:20;default:'input';comment:配置类型 input:text:textarea:select:switch"`
	Sort  int    `gorm:"default:0;comment:排序"`
	Remark string `gorm:"size:500;comment:备注"`
}

