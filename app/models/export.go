package models

import "github.com/goravel/framework/database/orm"

// Export 导出记录表
// 用于记录所有导出文件的信息，便于在后台进行统一管理
type Export struct {
	orm.Model
	AdminID   uint   `gorm:"index;comment:管理员ID"`
	Admin     Admin  `gorm:"foreignKey:AdminID"`
	Disk      string `gorm:"size:50;comment:存储驱动"`
	Path      string `gorm:"size:255;comment:文件路径"`
	Filename  string `gorm:"size:255;comment:文件名"`
	Extension string `gorm:"size:20;comment:文件后缀"`
	Size      int64  `gorm:"comment:文件大小(字节)"`
	Status    uint8  `gorm:"default:1;comment:状态 1:成功 0:失败"`
}


