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
	List(adminID uint, page int, pageSize int, notifType string, isRead string) ([]models.Notification, int64, error)
	ListRecent(adminID uint, limit int) ([]models.Notification, error)
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

func (s *NotificationServiceImpl) List(adminID uint, page int, pageSize int, notifType string, isRead string) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询条件
	// 对于私信类型，需要同时查询发送和接收的消息
	// 对于其他类型，只查询接收的消息
	countQuery := facades.Orm().Query().Model(&models.Notification{})
	if notifType == "message" {
		// 私信：查询发送或接收的消息
		countQuery = countQuery.Where("(receiver_id = ? OR sender_id = ?) AND type = ?", adminID, adminID, "message")
	} else if notifType != "" {
		// 指定了其他类型：只查询接收的消息
		countQuery = countQuery.Where("receiver_id = ? AND type = ?", adminID, notifType)
	} else {
		// 没有指定类型：查询接收的所有消息 + 发送的私信
		countQuery = countQuery.Where("receiver_id = ? OR (sender_id = ? AND type = ?)", adminID, adminID, "message")
	}

	// 如果指定了已读/未读状态，添加状态筛选
	if isRead == "true" {
		countQuery = countQuery.Where("is_read = ?", true)
	} else if isRead == "false" {
		countQuery = countQuery.Where("is_read = ?", false)
	}

	total, err := countQuery.Count()
	if err != nil {
		return nil, 0, err
	}

	// 构建列表查询
	listQuery := facades.Orm().Query().Model(&models.Notification{}).With("Sender").With("Receiver")
	if notifType == "message" {
		// 私信：查询发送或接收的消息
		listQuery = listQuery.Where("(receiver_id = ? OR sender_id = ?) AND type = ?", adminID, adminID, "message")
	} else if notifType != "" {
		// 指定了其他类型：只查询接收的消息
		listQuery = listQuery.Where("receiver_id = ? AND type = ?", adminID, notifType)
	} else {
		// 没有指定类型：查询接收的所有消息 + 发送的私信
		listQuery = listQuery.Where("receiver_id = ? OR (sender_id = ? AND type = ?)", adminID, adminID, "message")
	}

	// 如果指定了已读/未读状态，添加状态筛选
	if isRead == "true" {
		listQuery = listQuery.Where("is_read = ?", true)
	} else if isRead == "false" {
		listQuery = listQuery.Where("is_read = ?", false)
	}

	if err := listQuery.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (s *NotificationServiceImpl) ListRecent(adminID uint, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	// 查询最近的通知，包括接收的消息和发送的私信
	if err := facades.Orm().Query().Model(&models.Notification{}).With("Sender").With("Receiver").
		Where("(receiver_id = ? OR (sender_id = ? AND type = ?))", adminID, adminID, "message").
		Order("created_at desc").
		Limit(limit).
		Find(&notifications); err != nil {
		return nil, err
	}
	return notifications, nil
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
