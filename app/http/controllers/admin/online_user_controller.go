package admin

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type OnlineUserController struct {
}

func NewOnlineUserController() *OnlineUserController {
	return &OnlineUserController{}
}

// Index 在线用户列表
// 只显示在线的用户（last_used_at在最近5分钟内的）
func (r *OnlineUserController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	username := ctx.Request().Query("username", "")
	ip := ctx.Request().Query("ip", "")
	browser := ctx.Request().Query("browser", "")
	os := ctx.Request().Query("os", "")

	// 只查询最近15分钟内有活动的token（在线用户）
	// 默认只显示admin类型的token
	onlineThreshold := time.Now().Add(-15 * time.Minute)
	query := facades.Orm().Query().Model(&models.PersonalAccessToken{}).
		Where("tokenable_type", "admin").
		Where("last_used_at IS NOT NULL").
		Where("last_used_at >= ?", onlineThreshold)

	// 搜索条件
	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if browser != "" {
		query = query.Where("browser LIKE ?", "%"+browser+"%")
	}
	if os != "" {
		query = query.Where("os LIKE ?", "%"+os+"%")
	}

	var tokens []models.PersonalAccessToken
	if err := query.Order("last_used_at desc").Get(&tokens); err != nil {
		errorlog.RecordHTTP(ctx, "online_user", "Failed to query online user list", map[string]any{
			"error": err.Error(),
		}, "Query online user list error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}


	// 查询admin信息并组装数据，同时过滤username
	var onlineUsers []http.Json
	for _, token := range tokens {
		var admin models.Admin
		if err := facades.Orm().Query().Where("id", token.TokenableID).First(&admin); err != nil {
			continue
		}

		// 如果指定了username搜索条件，进行过滤
		if username != "" && !strings.Contains(strings.ToLower(admin.Username), strings.ToLower(username)) {
			continue
		}

		onlineUser := http.Json{
			"id":          token.ID,
			"admin_id":    admin.ID,
			"username":    admin.Username,
			"nickname":    admin.Nickname,
			"avatar":      admin.Avatar,
			"browser":     token.Browser,
			"ip":          token.IP,
			"os":          token.OS,
			"session_id":  token.SessionID,
			"last_active": token.LastUsedAt,
			"created_at":  token.CreatedAt,
		}
		onlineUsers = append(onlineUsers, onlineUser)
	}

	// 手动分页
	total := len(onlineUsers)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		onlineUsers = []http.Json{}
	} else if end > total {
		onlineUsers = onlineUsers[start:]
	} else {
		onlineUsers = onlineUsers[start:end]
	}

	return response.Paginate(ctx, "get_success", onlineUsers, int64(total), page, pageSize)
}

// KickOut 踢下线（删除token）
func (r *OnlineUserController) KickOut(ctx http.Context) http.Response {
	tokenID := cast.ToUint(ctx.Request().Route("id"))
	if tokenID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "token_id_required")
	}

	// 查询token是否存在
	var token models.PersonalAccessToken
	if err := facades.Orm().Query().Where("id", tokenID).First(&token); err != nil {
		errorlog.RecordHTTP(ctx, "online_user", "Token not found for kick out", map[string]any{
			"error":    err.Error(),
			"token_id": tokenID,
		}, "Token not found for kick out error: %v", err)
		return response.Error(ctx, http.StatusNotFound, "token_not_found")
	}

	// 删除token
	if _, err := facades.Orm().Query().Delete(&token); err != nil {
		errorlog.RecordHTTP(ctx, "online_user", "Failed to kick out user", map[string]any{
			"error":    err.Error(),
			"token_id": tokenID,
		}, "Kick out user error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "kick_out_failed")
	}

	return response.Success(ctx, "kick_out_success")
}

// BatchKickOut 批量踢下线
func (r *OnlineUserController) BatchKickOut(ctx http.Context) http.Response {
	tokenIDs := ctx.Request().Input("token_ids")
	if tokenIDs == "" {
		return response.Error(ctx, http.StatusBadRequest, "token_ids_required")
	}

	// 解析token IDs（假设是逗号分隔的字符串）
	var ids []uint
	idStrs := strings.Split(tokenIDs, ",")
	for _, idStr := range idStrs {
		idStr = strings.TrimSpace(idStr)
		if id := cast.ToUint(idStr); id > 0 {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "invalid_token_ids")
	}

	// 批量删除token
	idsAny := make([]interface{}, len(ids))
	for i, id := range ids {
		idsAny[i] = id
	}
	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.PersonalAccessToken{}); err != nil {
		errorlog.RecordHTTP(ctx, "online_user", "Failed to batch kick out users", map[string]any{
			"error":     err.Error(),
			"token_ids": ids,
		}, "Batch kick out users error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "batch_kick_out_failed")
	}

	return response.Success(ctx, "batch_kick_out_success", http.Json{
		"count": len(ids),
	})
}
