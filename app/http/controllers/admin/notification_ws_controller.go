package admin

import (
	"net/http"
	"net/url"
	"strings"

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
	// 记录 WebSocket 连接尝试（仅 Debug 模式）
	logger.DebugfHTTP(ctx, "WebSocket connection attempt from %s, path: %s, upgrade: %s, connection: %s",
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
	if err := facades.Orm().Query().Where("id", accessToken.TokenableID).FirstOrFail(&admin); err != nil {
		_ = ctx.Response().Json(http.StatusUnauthorized, apphttp.Json{
			"code":    http.StatusUnauthorized,
			"message": "user_not_found",
		}).Abort()
		return nil
	}
	_ = r.tokenService.UpdateLastUsedAt(token)

	upgrader := websocket.Upgrader{
		CheckOrigin: r.isOriginAllowed,
		ReadBufferSize:  1024, // 读缓冲区大小
		WriteBufferSize: 1024, // 写缓冲区大小
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

func (r *NotificationWsController) isOriginAllowed(req *http.Request) bool {
	origin := strings.TrimSpace(req.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Hostname() == "" {
		return false
	}

	originHost := strings.ToLower(parsed.Hostname())
	allowedAdminDomains := getConfigStringSlice("domains.admin")
	if len(allowedAdminDomains) > 0 && !matchDomain(originHost, allowedAdminDomains) {
		return false
	}

	allowedOrigins := getConfigStringSlice("cors.allowed_origins")
	if len(allowedOrigins) == 0 {
		return true
	}

	normalizedOrigin := strings.TrimRight(strings.ToLower(origin), "/")
	for _, allowed := range allowedOrigins {
		normalizedAllowed := strings.TrimSpace(strings.ToLower(allowed))
		if normalizedAllowed == "" {
			continue
		}
		if normalizedAllowed == "*" {
			return true
		}
		if strings.TrimRight(normalizedAllowed, "/") == normalizedOrigin {
			return true
		}
	}

	return false
}

func matchDomain(host string, patterns []string) bool {
	for _, pattern := range patterns {
		p := strings.TrimSpace(strings.ToLower(pattern))
		if p == "" {
			continue
		}
		if p == host {
			return true
		}
		if strings.HasPrefix(p, "*.") {
			suffix := strings.TrimPrefix(p, "*.")
			if strings.HasSuffix(host, "."+suffix) {
				return true
			}
		}
	}
	return false
}

func getConfigStringSlice(key string) []string {
	value := facades.Config().Get(key)
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return []string{}
	}
}
