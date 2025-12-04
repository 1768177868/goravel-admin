package admin

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

type BlacklistController struct {
}

func NewBlacklistController() *BlacklistController {
	return &BlacklistController{}
}

// Index 黑名单列表
func (r *BlacklistController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	query := facades.Orm().Query().Model(&models.Blacklist{})

	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var blacklists []models.Blacklist
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id desc").Get(&blacklists); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", blacklists, total, page, pageSize)
}

// Show 黑名单详情
func (r *BlacklistController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var blacklist models.Blacklist
	if err := facades.Orm().Query().Where("id", id).First(&blacklist); err != nil {
		return response.Error(ctx, http.StatusNotFound, "blacklist_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"blacklist": blacklist,
	})
}

// Store 创建黑名单
func (r *BlacklistController) Store(ctx http.Context) http.Response {
	ip := ctx.Request().Input("ip")
	remark := ctx.Request().Input("remark")
	status := cast.ToUint8(ctx.Request().Input("status", "1"))

	if ip == "" {
		return response.Error(ctx, http.StatusBadRequest, "blacklist_ip_required")
	}

	// 验证IP格式
	if errMsg := utils.ValidateBlacklistIP(ip); errMsg != "" {
		// 根据错误消息类型返回对应的错误码
		if strings.Contains(errMsg, "不能为空") {
			return response.Error(ctx, http.StatusBadRequest, "ip_address_required")
		} else if strings.Contains(errMsg, "CIDR格式错误") {
			return response.Error(ctx, http.StatusBadRequest, "invalid_cidr_format")
		} else if strings.Contains(errMsg, "IP范围格式错误") {
			return response.Error(ctx, http.StatusBadRequest, "invalid_ip_range_format")
		} else if strings.Contains(errMsg, "起始IP格式错误") || strings.Contains(errMsg, "结束IP格式错误") {
			return response.Error(ctx, http.StatusBadRequest, "invalid_ip_format")
		} else if strings.Contains(errMsg, "必须大于等于") {
			return response.Error(ctx, http.StatusBadRequest, "invalid_ip_range_order")
		} else {
			return response.Error(ctx, http.StatusBadRequest, "invalid_ip_format")
		}
	}

	now := carbon.Now()
	blacklistData := map[string]any{
		"ip":         ip,
		"remark":     remark,
		"status":     status,
		"created_at": now,
		"updated_at": now,
	}

	if err := facades.Orm().Query().Table("blacklists").Create(blacklistData); err != nil {
		errorlog.RecordHTTP(ctx, "blacklist", "Failed to create blacklist", map[string]any{
			"error": err.Error(),
			"ip":    ip,
		}, "Create blacklist error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	var blacklist models.Blacklist
	if err := facades.Orm().Query().Where("ip", ip).First(&blacklist); err != nil {
		errorlog.RecordHTTP(ctx, "blacklist", "Failed to query created blacklist", map[string]any{
			"error": err.Error(),
			"ip":    ip,
		}, "Query created blacklist error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return response.Success(ctx, "create_success", http.Json{
		"blacklist": blacklist,
	})
}

// Update 更新黑名单
func (r *BlacklistController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var blacklist models.Blacklist
	if err := facades.Orm().Query().Where("id", id).First(&blacklist); err != nil {
		return response.Error(ctx, http.StatusNotFound, "blacklist_not_found")
	}

	ip := ctx.Request().Input("ip")
	remark := ctx.Request().Input("remark")
	status := ctx.Request().Input("status", "")

	if ip != "" {
		// 验证IP格式
		if errMsg := utils.ValidateBlacklistIP(ip); errMsg != "" {
			// 根据错误消息类型返回对应的错误码
			if strings.Contains(errMsg, "不能为空") {
				return response.Error(ctx, http.StatusBadRequest, "ip_address_required")
			} else if strings.Contains(errMsg, "CIDR格式错误") {
				return response.Error(ctx, http.StatusBadRequest, "invalid_cidr_format")
			} else if strings.Contains(errMsg, "IP范围格式错误") {
				return response.Error(ctx, http.StatusBadRequest, "invalid_ip_range_format")
			} else if strings.Contains(errMsg, "起始IP格式错误") || strings.Contains(errMsg, "结束IP格式错误") {
				return response.Error(ctx, http.StatusBadRequest, "invalid_ip_format")
			} else if strings.Contains(errMsg, "必须大于等于") {
				return response.Error(ctx, http.StatusBadRequest, "invalid_ip_range_order")
			} else {
				return response.Error(ctx, http.StatusBadRequest, "invalid_ip_format")
			}
		}
		blacklist.IP = ip
	}
	if remark != "" {
		blacklist.Remark = remark
	}
	if status != "" {
		blacklist.Status = cast.ToUint8(status)
	}

	if err := facades.Orm().Query().Save(&blacklist); err != nil {
		errorlog.RecordHTTP(ctx, "blacklist", "Failed to update blacklist", map[string]any{
			"error":        err.Error(),
			"blacklist_id": blacklist.ID,
		}, "Update blacklist error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"blacklist": blacklist,
	})
}

// Destroy 删除黑名单
func (r *BlacklistController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var blacklist models.Blacklist
	if err := facades.Orm().Query().Where("id", id).First(&blacklist); err != nil {
		return response.Error(ctx, http.StatusNotFound, "blacklist_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&blacklist); err != nil {
		errorlog.RecordHTTP(ctx, "blacklist", "Failed to delete blacklist", map[string]any{
			"error":        err.Error(),
			"blacklist_id": blacklist.ID,
		}, "Delete blacklist error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
