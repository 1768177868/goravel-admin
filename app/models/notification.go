package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type Notification struct {
	orm.Model
	Title      string     `gorm:"size:150;not null" json:"title"`
	Content    string     `gorm:"type:text" json:"content"`
	Type       string     `gorm:"size:20;default:announcement" json:"type"`
	SenderID   *uint      `json:"sender_id"`
	Sender     *Admin     `gorm:"foreignKey:SenderID" json:"sender"`
	ReceiverID *uint      `json:"receiver_id"`
	Receiver   *Admin     `gorm:"foreignKey:ReceiverID" json:"receiver"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	orm.SoftDeletes
}
