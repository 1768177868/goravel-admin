package services

import (
	"errors"
	"time"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
	wsnotifications "goravel/app/websocket/notifications"
)

type NotificationService interface {
	Create(title, content, notifType string, senderID *uint, receiverID *uint) (*models.Notification, error)
	List(adminID uint, page int, pageSize int) ([]models.Notification, int64, error)
	MarkRead(adminID uint, notificationID uint) error
	MarkAllRead(adminID uint) error
	UnreadCount(adminID uint) (int64, error)
}

type NotificationServiceImpl struct{}

func NewNotificationServiceImpl() NotificationService {
	return &NotificationServiceImpl{}
}

func (s *NotificationServiceImpl) Create(title, content, notifType string, senderID *uint, receiverID *uint) (*models.Notification, error) {
	if receiverID == nil {
		var admins []models.Admin
		if err := facades.Orm().Query().Find(&admins); err != nil {
			return nil, err
		}
		var first *models.Notification
		for _, admin := range admins {
			rid := admin.ID
			notification := &models.Notification{
				Title:      title,
				Content:    content,
				Type:       notifType,
				SenderID:   senderID,
				ReceiverID: &rid,
			}
			if err := facades.Orm().Query().Create(notification); err != nil {
				return nil, err
			}
			if first == nil {
				first = notification
			}
			wsnotifications.Hub().Broadcast(notification)
		}
		return first, nil
	}

	notification := &models.Notification{
		Title:      title,
		Content:    content,
		Type:       notifType,
		SenderID:   senderID,
		ReceiverID: receiverID,
	}
	if err := facades.Orm().Query().Create(notification); err != nil {
		return nil, err
	}

	wsnotifications.Hub().Broadcast(notification)

	return notification, nil
}

func (s *NotificationServiceImpl) List(adminID uint, page int, pageSize int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	countQuery := facades.Orm().Query().Model(&models.Notification{}).
		Where("receiver_id = ?", adminID)
	total, err := countQuery.Count()
	if err != nil {
		return nil, 0, err
	}

	listQuery := facades.Orm().Query().Model(&models.Notification{}).
		Where("receiver_id = ?", adminID)

	if err := listQuery.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (s *NotificationServiceImpl) MarkRead(adminID uint, notificationID uint) error {
	var notification models.Notification
	if err := facades.Orm().Query().Where("id = ?", notificationID).
		Where("receiver_id = ?", adminID).
		First(&notification); err != nil {
		return errors.New("notification_not_found")
	}

	if notification.IsRead {
		return nil
	}

	now := time.Now()

	_, err := facades.Orm().Query().
		Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Update(map[string]any{
			"is_read": true,
			"read_at": now,
		})

	return err
}

func (s *NotificationServiceImpl) MarkAllRead(adminID uint) error {
	now := time.Now()
	_, err := facades.Orm().Query().
		Table("notifications").
		Where("receiver_id = ?", adminID).
		Where("is_read = ?", false).
		Update(map[string]any{
			"is_read": true,
			"read_at": now,
		})
	return err
}

func (s *NotificationServiceImpl) UnreadCount(adminID uint) (int64, error) {
	query := facades.Orm().Query().Model(&models.Notification{}).
		Where("receiver_id = ?", adminID).
		Where("is_read = ?", false)

	return query.Count()
}
