package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
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
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)
	notifType := ctx.Request().Query("type", "")
	isRead := ctx.Request().Query("is_read", "")
	notifications, total, err := r.service.List(admin.ID, page, pageSize, notifType, isRead)
	if err != nil {
		logger.ErrorfHTTP(ctx, "list notifications error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	count, err := r.service.UnreadCount(admin.ID)
	if err != nil {
		logger.ErrorfHTTP(ctx, "unread count error: %v", err)
	}

	return response.Success(ctx, http.Json{
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
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	count, err := r.service.UnreadCount(admin.ID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	return response.Success(ctx, http.Json{
		"count": count,
	})
}

func (r *NotificationController) Recent(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	limit := helpers.GetIntQuery(ctx, "limit", 5)
	notifications, err := r.service.ListRecent(admin.ID, limit)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	count, _ := r.service.UnreadCount(admin.ID)

	return response.Success(ctx, http.Json{
		"notifications": notifications,
		"unread_count":  count,
	})
}

func (r *NotificationController) MarkRead(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsRequired.Code)
	}

	if err := r.service.MarkRead(admin.ID, id); err != nil {
		// 使用业务错误类型，直接提取错误码
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusNotFound, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
	}

	return response.Success(ctx, http.Json{
		"id": id,
	})
}

func (r *NotificationController) MarkAllRead(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	if err := r.service.MarkAllRead(admin.ID); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
	}

	return response.Success(ctx)
}

func (r *NotificationController) Store(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	title := ctx.Request().Input("title")
	content := ctx.Request().Input("content")
	notificationType := ctx.Request().Input("type", "announcement")
	if title == "" || content == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsRequired.Code)
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
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}

	if notification == nil {
		return response.Success(ctx)
	}

	return response.Success(ctx, http.Json{
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
