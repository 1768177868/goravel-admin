package models

import (
	"github.com/goravel/framework/database/orm"
)

type OperationLog struct {
	orm.Model
	AdminID   uint   `gorm:"index;comment:管理员ID"`
	Admin     Admin  `gorm:"foreignKey:AdminID"`
	Method    string `gorm:"size:10;comment:请求方法"`
	Path      string `gorm:"size:255;comment:请求路径"`
	Title     string `gorm:"size:255;comment:操作标题"`
	IP        string `gorm:"size:50;comment:IP地址"`
	UserAgent string `gorm:"size:500;comment:用户代理"`
	Request   string `gorm:"type:text;comment:请求参数"`
	Response  string `gorm:"type:text;comment:响应数据"`
	Status    uint8  `gorm:"default:1;comment:状态 1:成功 0:失败"`
	ErrorMsg  string `gorm:"type:text;comment:错误信息"`
	Duration  int    `gorm:"comment:耗时(毫秒)"`
}
