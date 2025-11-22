package models

import (
	"github.com/goravel/framework/database/orm"
)

type LoginLog struct {
	orm.Model
	AdminID   uint   `gorm:"index;comment:管理员ID"`
	Admin     Admin  `gorm:"foreignKey:AdminID"`
	Username  string `gorm:"size:50;comment:用户名"`
	IP        string `gorm:"size:50;comment:IP地址"`
	UserAgent string `gorm:"size:500;comment:用户代理"`
	Location  string `gorm:"size:100;comment:登录地点"`
	Status    uint8  `gorm:"default:1;comment:状态 1:成功 0:失败"`
	Message   string `gorm:"size:255;comment:登录信息"`
}
