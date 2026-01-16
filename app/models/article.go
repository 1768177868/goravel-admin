package models

import (
	"github.com/goravel/framework/database/orm"
)

type Article struct {
	orm.Model
	orm.SoftDeletes

	Name string `gorm:"name" json:"name" comment:"名称"`

	Status string `gorm:"status" json:"status" comment:"状态"`

	AdminId string `gorm:"admin_id" json:"admin_id"`

	Author *Admin `gorm:"foreignKey:AdminId" json:"author"`
}

func (Article) TableName() string {
	return "articles"
}

func (r *Article) Serialize() map[string]any {
	return map[string]any{
		"id":         r.ID,
		"created_at": r.CreatedAt,
		"updated_at": r.UpdatedAt,

		"name": r.Name,

		"status": r.Status,

		"admin_id": r.AdminId,
	}
}

func (r *Article) Deserialize(data map[string]any) {

	if val, ok := data["name"]; ok {
		r.Name = val.(string)
	}

	if val, ok := data["status"]; ok {
		r.Status = val.(string)
	}

	if val, ok := data["admin_id"]; ok {
		r.AdminId = val.(string)
	}

}
