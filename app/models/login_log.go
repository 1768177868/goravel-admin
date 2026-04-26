package models

import (
	"github.com/goravel/framework/database/orm"
)

type LoginLog struct {
	orm.Model
	AdminID   uint   `gorm:"index;comment:管理员ID" json:"admin_id"`
	Admin     Admin  `gorm:"foreignKey:AdminID" json:"admin"`
	Username  string `gorm:"size:50;comment:用户名" json:"username"`
	IP        string `gorm:"size:50;comment:IP地址" json:"ip"`
	UserAgent string `gorm:"size:500;comment:用户代理" json:"user_agent"`
	Location  string `gorm:"size:100;comment:登录地点" json:"location"`
	Status    uint8  `gorm:"comment:状态 1:成功 0:失败" json:"status"`
	Message   string `gorm:"size:255;comment:登录信息" json:"message"`
	Request   string `gorm:"type:text;comment:请求数据" json:"request"`
}
