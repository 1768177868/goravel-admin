package services

import (
	"context"
	appfacades "goravel/app/facades"
	"time"

	"github.com/goravel/framework/contracts/database/orm"

	apperrors "goravel/app/errors"
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

type NotificationServiceImpl struct {
	ctx context.Context
}

func NewNotificationServiceImpl(ctx context.Context) NotificationService {
	return &NotificationServiceImpl{ctx: ctx}
}

func (s *NotificationServiceImpl) Create(title, content, notifType string, senderID *uint, receiverID *uint) (*models.Notification, error) {
	if receiverID == nil {
		// 批量创建通知给所有管理员，使用事务确保原子性
		var admins []models.Admin
		if err := appfacades.OrmQuery(s.ctx).Find(&admins); err != nil {
			return nil, err
		}

		if len(admins) == 0 {
			return nil, apperrors.ErrRecordNotFound.WithMessage("no admins found")
		}

		var first *models.Notification
		var notifications []*models.Notification
		var createdIDs []uint

		// 使用循环创建通知，如果失败则手动回滚已创建的通知
		// 注意：这不是真正的事务，但在框架可能不支持事务的情况下提供基本的回滚机制
		for _, admin := range admins {
			rid := admin.ID
			notification := &models.Notification{
				Title:      title,
				Content:    content,
				Type:       notifType,
				SenderID:   senderID,
				ReceiverID: &rid,
			}
			if err := appfacades.OrmQuery(s.ctx).Create(notification); err != nil {
				// 如果创建失败，尝试删除已创建的通知（手动回滚）
				for _, id := range createdIDs {
					_, _ = appfacades.OrmQuery(s.ctx).Where("id", id).Delete(&models.Notification{})
				}
				return nil, err
			}
			if first == nil {
				first = notification
			}
			notifications = append(notifications, notification)
			createdIDs = append(createdIDs, notification.ID)
		}

		// 事务成功后，在事务外进行 WebSocket 广播（避免阻塞事务）
		for _, notification := range notifications {
			wsnotifications.Hub().Broadcast(notification)
		}

		return first, nil
	}

	// 单个通知创建
	notification := &models.Notification{
		Title:      title,
		Content:    content,
		Type:       notifType,
		SenderID:   senderID,
		ReceiverID: receiverID,
	}
	if err := appfacades.OrmQuery(s.ctx).Create(notification); err != nil {
		return nil, err
	}

	wsnotifications.Hub().Broadcast(notification)

	return notification, nil
}

// buildNotificationQuery 构建通知查询条件（消除代码重复）
// 对于私信类型，需要同时查询发送和接收的消息
// 对于其他类型，只查询接收的消息
func (s *NotificationServiceImpl) buildNotificationQuery(adminID uint, notifType, isRead string) orm.Query {
	query := appfacades.OrmQuery(s.ctx).Model(&models.Notification{})

	if notifType == "message" {
		// 私信：查询发送或接收的消息
		query = query.Where("(receiver_id = ? OR sender_id = ?) AND type = ?", adminID, adminID, "message")
	} else if notifType != "" {
		// 指定了其他类型：只查询接收的消息
		query = query.Where("receiver_id = ? AND type = ?", adminID, notifType)
	} else {
		// 没有指定类型：查询接收的所有消息 + 发送的私信
		query = query.Where("receiver_id = ? OR (sender_id = ? AND type = ?)", adminID, adminID, "message")
	}

	// 如果指定了已读/未读状态，添加状态筛选
	if isRead == "true" {
		query = query.Where("is_read = ?", true)
	} else if isRead == "false" {
		query = query.Where("is_read = ?", false)
	}

	return query
}

func (s *NotificationServiceImpl) List(adminID uint, page int, pageSize int, notifType string, isRead string) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 分页查询
	var total int64
	query := s.buildNotificationQuery(adminID, notifType, isRead).With("Sender").With("Receiver").Order("created_at desc")
	if err := query.Paginate(page, pageSize, &notifications, &total); err != nil {
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
	if err := appfacades.OrmQuery(s.ctx).Model(&models.Notification{}).With("Sender").With("Receiver").
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
	if err := appfacades.OrmQuery(s.ctx).Where("id = ?", notificationID).
		Where("receiver_id = ?", adminID).
		First(&notification); err != nil {
		return apperrors.ErrRecordNotFound.WithMessage("notification not found")
	}

	if notification.IsRead {
		return nil
	}

	now := time.Now()

	_, err := appfacades.OrmQuery(s.ctx).
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
	_, err := appfacades.OrmQuery(s.ctx).
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
	query := appfacades.OrmQuery(s.ctx).Model(&models.Notification{}).
		Where("receiver_id = ?", adminID).
		Where("is_read = ?", false)

	return query.Count()
}
