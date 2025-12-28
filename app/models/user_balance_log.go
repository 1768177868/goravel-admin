package models

import (
	"github.com/goravel/framework/database/orm"
)

// UserBalanceLog 用户余额变动记录（使用 GORM Sharding 分表）
type UserBalanceLog struct {
	orm.Model
	UserID      uint    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	User        *User   `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联 users 表
	Type        string  `gorm:"size:20;not null;comment:变动类型:income收入,expense支出,refund退款" json:"type"`
	Amount      float64 `gorm:"type:decimal(18,8);not null;comment:变动金额" json:"amount"`
	Balance     float64 `gorm:"type:decimal(18,8);not null;comment:变动后余额" json:"balance"`
	Source      string  `gorm:"size:50;comment:来源:order订单,recharge充值,withdraw提现,manual手动" json:"source"`
	SourceID    *uint   `gorm:"index;comment:来源ID(如订单ID)" json:"source_id"`
	Description string  `gorm:"type:text;comment:描述" json:"description"`
	OperatorID  *uint   `gorm:"index;comment:操作员ID" json:"operator_id"`
	Operator    *Admin  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"` // 关联 admins 表
	Status      string  `gorm:"size:20;default:'success';comment:状态:success成功,failed失败" json:"status"`
	Remark      string  `gorm:"type:text;comment:备注" json:"remark"`
}

// TableName 指定表名
func (UserBalanceLog) TableName() string {
	return "user_balance_logs"
}
