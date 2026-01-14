package models

import (
	"github.com/goravel/framework/database/orm"
)

type Article struct {
	orm.Model

	Name string `gorm:"name" json:"name" comment:"名称""`

	Status string `gorm:"status" json:"status" comment:"状态""`

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

	}
}

func (r *Article) Deserialize(data map[string]any) {

	if val, ok := data["name"]; ok {
		r.Name = val.(string)
	}

	if val, ok := data["status"]; ok {
		r.Status = val.(string)
	}

}
