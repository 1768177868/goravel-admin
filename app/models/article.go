package models

import (
	"github.com/goravel/framework/database/orm"
)

type Article struct {
	orm.Model

	AdminId uint `gorm:"column:admin_id" json:"admin_id" comment:"用户ID"`

	Admin *Admin `gorm:"foreignKey:AdminId" json:"admin"`

	Title string `gorm:"column:title" json:"title" comment:"标题"`

	Content string `gorm:"column:content" json:"content" comment:"内容"`

	Status uint8 `gorm:"column:status" json:"status" comment:"0:未发布 1:发布"`

	orm.SoftDeletes
}

func (Article) TableName() string {
	return "articles"
}
