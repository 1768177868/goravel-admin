package admin

import (
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/constants"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type OnlineUserController struct {
}

func NewOnlineUserController() *OnlineUserController {
	return &OnlineUserController{}
}

// Index 获取在线用户列表
// 只显示最近15分钟内有活动的用户（根据 OnlineUserThreshold 常量判断）
func (r *OnlineUserController) Index(ctx http.Context) http.Response {
	// 验证并规范化分页参数
	page, pageSize := helpers.ValidatePagination(
		helpers.GetIntQuery(ctx, "page", 1),
		helpers.GetIntQuery(ctx, "page_size", 10),
	)

	username := ctx.Request().Query("username", "")
	ip := ctx.Request().Query("ip", "")
	browser := ctx.Request().Query("browser", "")
	os := ctx.Request().Query("os", "")
	orderBy := ctx.Request().Query("order_by", "")

	// 只查询最近15分钟内有活动的token（在线用户）
	// 默认只显示admin类型的token
	onlineThreshold := time.Now().Add(-constants.OnlineUserThreshold)
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

	// 应用排序，默认排序为 last_used_at desc
	query = helpers.ApplySort(query, orderBy, "last_used_at:desc")

	var tokens []models.PersonalAccessToken
	if err := query.Get(&tokens); err != nil {
		errorlog.RecordHTTP(ctx, "online_user", "Failed to query online user list", map[string]any{
			"error": err.Error(),
		}, "Query online user list error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	// 批量查询所有 admin 信息，避免 N+1 查询
	var adminIDs []uint
	adminIDMap := make(map[uint]bool) // 用于去重
	for _, token := range tokens {
		if !adminIDMap[token.TokenableID] {
			adminIDs = append(adminIDs, token.TokenableID)
			adminIDMap[token.TokenableID] = true
		}
	}

	// 批量查询 admin
	adminMap := make(map[uint]models.Admin)
	if len(adminIDs) > 0 {
		var admins []models.Admin
		if err := facades.Orm().Query().Where("id IN ?", adminIDs).Find(&admins); err != nil {
			errorlog.RecordHTTP(ctx, "online_user", "Failed to query admin list", map[string]any{
				"error":     err.Error(),
				"admin_ids": adminIDs,
			}, "Query admin list error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "query_failed")
		}

		// 构建 admin map
		for _, admin := range admins {
			adminMap[admin.ID] = admin
		}
	}

	// 组装数据，同时过滤 username
	var onlineUsers []http.Json
	for _, token := range tokens {
		admin, ok := adminMap[token.TokenableID]
		if !ok {
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

	// 使用工具函数进行分页
	paginatedUsers, total := helpers.PaginateSlice(onlineUsers, page, pageSize)

	return response.Paginate(ctx, "get_success", paginatedUsers, total, page, pageSize)
}

// KickOut 踢下线（删除token）
func (r *OnlineUserController) KickOut(ctx http.Context) http.Response {
	tokenID := helpers.GetUintRoute(ctx, "id")
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

	// 使用工具函数解析 token IDs
	ids := helpers.ParseIDsFromString(tokenIDs)
	if len(ids) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "invalid_token_ids")
	}

	// 批量删除token
	idsAny := helpers.ConvertUintSliceToAny(ids)
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
