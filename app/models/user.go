package models

import (
	"database/sql/driver"
	"errors"

	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/json"
)

type User struct {
	orm.Model
	Username string    `gorm:"size:50;uniqueIndex;comment:用户名" json:"username"`
	Password string    `gorm:"size:255;comment:密码" json:"-"` // 使用 json:"-" 隐藏密码字段
	Name     string    `gorm:"size:100;comment:姓名" json:"name"`
	Avatar   string    `gorm:"size:255;comment:头像" json:"avatar"`
	Alias    string    `gorm:"size:100;comment:别名" json:"alias"`
	Mail     string    `gorm:"size:100;comment:邮箱" json:"mail"`
	Status   uint8     `gorm:"default:1;comment:状态 1:启用 0:禁用" json:"status"`
	Tags     []UserTag `gorm:"serializer:json" json:"tags"`
	orm.SoftDeletes
}

type UserTag struct {
	Key string `json:"key"`
	Val int    `json:"value"`
}

func (r *UserTag) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, r)
}

func (r *UserTag) Value() (driver.Value, error) {
	return json.Marshal(r)
}
