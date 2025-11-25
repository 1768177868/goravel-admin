package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type NotificationController struct {
	service services.NotificationService
}

func NewNotificationController() *NotificationController {
	return &NotificationController{
		service: services.NewNotificationServiceImpl(),
	}
}

func (r *NotificationController) Index(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	page := cast.ToInt(ctx.Request().Input("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Input("page_size", "20"))
	notifications, total, err := r.service.List(admin.ID, page, pageSize)
	if err != nil {
		facades.Log().Errorf("list notifications error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}
	count, err := r.service.UnreadCount(admin.ID)
	if err != nil {
		facades.Log().Errorf("unread count error: %v", err)
	}

	return response.Success(ctx, "get_success", http.Json{
		"notifications": notifications,
		"unread_count":  count,
		"pagination": http.Json{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (r *NotificationController) UnreadCount(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	count, err := r.service.UnreadCount(admin.ID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, "get_success", http.Json{
		"count": count,
	})
}

func (r *NotificationController) MarkRead(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	id := cast.ToUint(ctx.Request().Route("id"))
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "params_required")
	}

	if err := r.service.MarkRead(admin.ID, id); err != nil {
		if err.Error() == "notification_not_found" {
			return response.Error(ctx, http.StatusNotFound, "notification_not_found")
		}
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"id": id,
	})
}

func (r *NotificationController) MarkAllRead(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	if err := r.service.MarkAllRead(admin.ID); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success")
}

func (r *NotificationController) Store(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	title := ctx.Request().Input("title")
	content := ctx.Request().Input("content")
	notificationType := ctx.Request().Input("type", "announcement")
	if title == "" || content == "" {
		return response.Error(ctx, http.StatusBadRequest, "params_required")
	}

	var receiverID *uint
	receiverVal := ctx.Request().Input("receiver_id")
	if receiverVal != "" {
		id := cast.ToUint(receiverVal)
		if id > 0 {
			receiverID = &id
		}
	}

	senderID := admin.ID
	notification, err := r.service.Create(title, content, notificationType, &senderID, receiverID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	if notification == nil {
		return response.Success(ctx, "create_success")
	}

	return response.Success(ctx, "create_success", http.Json{
		"notification": notification,
	})
}

func (r *NotificationController) currentAdmin(ctx http.Context) *models.Admin {
	if adminValue := ctx.Value("admin"); adminValue != nil {
		if admin, ok := adminValue.(models.Admin); ok {
			return &admin
		}
		if adminPtr, ok := adminValue.(*models.Admin); ok {
			return adminPtr
		}
	}
	return nil
}
