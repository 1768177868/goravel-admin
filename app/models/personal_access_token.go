package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type PersonalAccessToken struct {
	orm.Model
	TokenableType string     `gorm:"column:tokenable_type;type:varchar(255);comment:模型类型" json:"tokenable_type"`
	TokenableID   uint       `gorm:"column:tokenable_id;type:bigint;comment:模型ID" json:"tokenable_id"`
	Name          string     `gorm:"column:name;type:varchar(255);default:'';comment:token名称" json:"name"`
	Token         string     `gorm:"column:token;type:varchar(64);uniqueIndex;comment:token值（hash）" json:"token"`
	Abilities     string     `gorm:"column:abilities;type:text;nullable;comment:权限列表（JSON）" json:"abilities"`
	LastUsedAt    *time.Time `gorm:"column:last_used_at;type:timestamp;nullable;comment:最后使用时间" json:"last_used_at"`
	ExpiresAt     *time.Time `gorm:"column:expires_at;type:timestamp;nullable;comment:过期时间，NULL表示永不过期" json:"expires_at"`
}

func (PersonalAccessToken) TableName() string {
	return "personal_access_tokens"
}

