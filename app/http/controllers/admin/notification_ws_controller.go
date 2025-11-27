package admin

import (
	"net/http"
	"strings"

	apphttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/gorilla/websocket"

	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
	wsnotifications "goravel/app/websocket/notifications"
)

type NotificationWsController struct {
	tokenService services.TokenService
}

func NewNotificationWsController() *NotificationWsController {
	return &NotificationWsController{
		tokenService: services.NewTokenServiceImpl(),
	}
}

func (r *NotificationWsController) Server(ctx apphttp.Context) apphttp.Response {
	token := ctx.Request().Query("token")
	if token == "" {
		_ = ctx.Response().Json(http.StatusUnauthorized, apphttp.Json{
			"code":    http.StatusUnauthorized,
			"message": "token_required",
		}).Abort()
		return nil
	}

	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	accessToken, err := r.tokenService.FindToken(token)
	if err != nil || accessToken == nil || accessToken.TokenableType != "admin" {
		_ = ctx.Response().Json(http.StatusUnauthorized, apphttp.Json{
			"code":    http.StatusUnauthorized,
			"message": "invalid_token",
		}).Abort()
		return nil
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("id", accessToken.TokenableID).First(&admin); err != nil {
		_ = ctx.Response().Json(http.StatusUnauthorized, apphttp.Json{
			"code":    http.StatusUnauthorized,
			"message": "user_not_found",
		}).Abort()
		return nil
	}
	_ = r.tokenService.UpdateLastUsedAt(token)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(ctx.Response().Writer(), ctx.Request().Origin(), nil)
	if err != nil {
		logger.ErrorfHTTP(ctx, "notification ws upgrade error: %v", err)
		return ctx.Response().String(http.StatusInternalServerError, "upgrade_failed")
	}

	wsnotifications.Hub().RegisterConnection(conn, admin.ID)

	return nil
}
