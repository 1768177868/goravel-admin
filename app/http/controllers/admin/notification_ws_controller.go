package admin

import (
	"net/http"

	apphttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
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
	// 记录 WebSocket 连接尝试
	logger.InfofHTTP(ctx, "WebSocket connection attempt from %s, path: %s, upgrade: %s, connection: %s",
		ctx.Request().Ip(),
		ctx.Request().Path(),
		ctx.Request().Header("Upgrade", ""),
		ctx.Request().Header("Connection", ""))

	token := ctx.Request().Query("token")
	if token == "" {
		logger.WarnfHTTP(ctx, "WebSocket connection rejected: token required")
		_ = ctx.Response().Json(http.StatusUnauthorized, apphttp.Json{
			"code":    http.StatusUnauthorized,
			"message": "token_required",
		}).Abort()
		return nil
	}

	token = str.Of(token).ChopStart("Bearer ").Trim().String()
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

	// logger.InfofHTTP(ctx, "WebSocket connection established for admin ID: %d", admin.ID)
	wsnotifications.Hub().RegisterConnection(conn, admin.ID)

	return nil
}
