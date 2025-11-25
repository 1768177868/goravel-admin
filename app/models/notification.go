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
	ReceiverID *uint      `json:"receiver_id"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	orm.SoftDeletes
}
