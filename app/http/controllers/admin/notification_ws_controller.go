package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	apphttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
	wsnotifications "goravel/app/websocket/notifications"
)

type NotificationWsController struct {
	tokenService services.TokenService
}

const (
	wsTicketCachePrefix = "ws:ticket:"
	wsTicketTTL         = 60 * time.Second
)

func NewNotificationWsController() *NotificationWsController {
	return &NotificationWsController{
		tokenService: services.NewTokenServiceImpl(),
	}
}

func (r *NotificationWsController) Ticket(ctx apphttp.Context) apphttp.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	token := str.Of(ctx.Request().Header("Authorization", "")).ChopStart("Bearer ").Trim().String()
	if token == "" {
		return response.Error(ctx, http.StatusUnauthorized, "token_required")
	}

	ticket := strings.ToLower(ulid.Make().String())
	cacheKey := wsTicketCachePrefix + ticket
	cacheValue := fmt.Sprintf("%d|%s", admin.ID, token)
	if err := facades.Cache().Put(cacheKey, cacheValue, wsTicketTTL); err != nil {
		return response.ErrorWithLog(ctx, "notification", err, map[string]any{
			"admin_id": admin.ID,
		})
	}

	return response.Success(ctx, apphttp.Json{
		"ticket":     ticket,
		"expires_in": int(wsTicketTTL / time.Second),
	})
}

func (r *NotificationWsController) Server(ctx apphttp.Context) apphttp.Response {
	// 记录 WebSocket 连接尝试（仅 Debug 模式）
	logger.DebugfHTTP(ctx, "WebSocket connection attempt from %s, path: %s, upgrade: %s, connection: %s",
		ctx.Request().Ip(),
		ctx.Request().Path(),
		ctx.Request().Header("Upgrade", ""),
		ctx.Request().Header("Connection", ""))

	token := r.extractToken(ctx)
	if token == "" {
		logger.WarnfHTTP(ctx, "WebSocket connection rejected: token required")
		response.Error(ctx, http.StatusUnauthorized, "token_required")
		ctx.Request().Abort()
		return nil
	}

	token = str.Of(token).ChopStart("Bearer ").Trim().String()
	accessToken, err := r.tokenService.FindToken(token)
	if err != nil || accessToken == nil || accessToken.TokenableType != "admin" {
		response.Error(ctx, http.StatusUnauthorized, "invalid_token")
		ctx.Request().Abort()
		return nil
	}

	var admin models.Admin
	if err := facades.Orm().Query().Where("id", accessToken.TokenableID).FirstOrFail(&admin); err != nil {
		response.Error(ctx, http.StatusUnauthorized, "user_not_found")
		ctx.Request().Abort()
		return nil
	}
	_ = r.tokenService.UpdateLastUsedAt(token)

	upgrader := websocket.Upgrader{
		CheckOrigin:     r.isOriginAllowed,
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

func (r *NotificationWsController) extractToken(ctx apphttp.Context) string {
	ticket := str.Of(ctx.Request().Query("ticket")).Trim().String()
	if ticket != "" {
		token, ok := r.consumeTicket(ticket)
		if ok {
			return token
		}
		logger.WarnfHTTP(ctx, "WebSocket ticket invalid or expired")
	}

	// Keep header parsing only for non-browser clients.
	authorization := str.Of(ctx.Request().Header("Authorization", "")).Trim().String()
	if authorization != "" {
		return authorization
	}

	return ""
}

func (r *NotificationWsController) consumeTicket(ticket string) (string, bool) {
	cacheKey := wsTicketCachePrefix + ticket
	value := facades.Cache().GetString(cacheKey, "")
	if value == "" {
		return "", false
	}
	_ = facades.Cache().Forget(cacheKey)

	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return "", false
	}
	if _, err := strconv.ParseUint(parts[0], 10, 32); err != nil {
		return "", false
	}
	return parts[1], parts[1] != ""
}

func (r *NotificationWsController) currentAdmin(ctx apphttp.Context) *models.Admin {
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
