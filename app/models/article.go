package models

import (
	"github.com/goravel/framework/database/orm"
)

type Article struct {
	orm.Model

	Title string `gorm:"column:title" json:"title" comment:"标题"`

	Content string `gorm:"column:content" json:"content" comment:"文章内容"`

	Status uint8 `gorm:"column:status" json:"status" comment:"0:未发布 1:发布"`

	AdminId uint `gorm:"column:admin_id" json:"admin_id" comment:"管理id"`

	Admin *Admin `gorm:"foreignKey:AdminId" json:"admin"`

	orm.SoftDeletes
}

func (Article) TableName() string {
	return "articles"
}
