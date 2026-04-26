package models

import (
	"github.com/goravel/framework/database/orm"
)

type SystemLog struct {
	orm.Model
	Level     string `gorm:"size:20;comment:日志级别" json:"level"`
	Module    string `gorm:"size:50;comment:模块" json:"module"`
	TraceID   string `gorm:"size:120;comment:链路ID" json:"trace_id"`
	Message   string `gorm:"type:text;comment:日志消息" json:"message"`
	Context   string `gorm:"type:text;comment:上下文信息" json:"context"`
	IP        string `gorm:"size:50;comment:IP地址" json:"ip"`
	UserAgent string `gorm:"size:500;comment:用户代理" json:"user_agent"`
}
