package admin

import (
	"strings"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils"
)

type BlacklistController struct {
}

func NewBlacklistController() *BlacklistController {
	return &BlacklistController{}
}

// findBlacklistByID 根据ID查找黑名单，如果不存在则返回错误响应
func (r *BlacklistController) findBlacklistByID(ctx http.Context, id uint) (*models.Blacklist, http.Response) {
	return response.FindByID[models.Blacklist](ctx, id, &response.FindByIDOptions{
		NotFoundMessageKey: "blacklist_not_found",
	})
}

// buildQuery 构建黑名单查询
func (r *BlacklistController) buildQuery(ctx http.Context) orm.Query {
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

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 id desc
	query = helpers.ApplySort(query, orderBy, "id:desc")

	return query
}

// Index 黑名单列表
func (r *BlacklistController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var blacklists []models.Blacklist
	return response.PaginateQuery(ctx, query, &blacklists, nil)
}

// Show 黑名单详情
func (r *BlacklistController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	blacklist, resp := r.findBlacklistByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"blacklist": *blacklist,
	})
}

// Store 创建黑名单
func (r *BlacklistController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var blacklistCreate adminrequests.BlacklistCreate
	errors, err := ctx.Request().ValidateRequest(&blacklistCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 验证IP格式（使用自定义验证函数）
	if errMsg := utils.ValidateBlacklistIP(blacklistCreate.IP); errMsg != "" {
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
		"ip":         blacklistCreate.IP,
		"remark":     blacklistCreate.Remark,
		"status":     blacklistCreate.Status,
		"created_at": now,
		"updated_at": now,
	}

	if err := facades.Orm().Query().Table("blacklists").Create(blacklistData); err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"ip": blacklistCreate.IP,
		})
	}

	var blacklist models.Blacklist
	if err := facades.Orm().Query().Where("ip", blacklistCreate.IP).First(&blacklist); err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"ip": blacklistCreate.IP,
		})
	}

	return response.Success(ctx, http.Json{
		"blacklist": blacklist,
	})
}

// Update 更新黑名单
func (r *BlacklistController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	blacklist, resp := r.findBlacklistByID(ctx, id)
	if resp != nil {
		return resp
	}

	// 使用请求验证
	var blacklistUpdate adminrequests.BlacklistUpdate
	errors, err := ctx.Request().ValidateRequest(&blacklistUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用 All() 方法检查字段是否存在
	allInputs := ctx.Request().All()

	if _, exists := allInputs["ip"]; exists {
		// 验证IP格式（使用自定义验证函数）
		if errMsg := utils.ValidateBlacklistIP(blacklistUpdate.IP); errMsg != "" {
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
		blacklist.IP = blacklistUpdate.IP
	}
	if _, exists := allInputs["remark"]; exists {
		blacklist.Remark = blacklistUpdate.Remark
	}
	if _, exists := allInputs["status"]; exists {
		blacklist.Status = blacklistUpdate.Status
	}

	if err := facades.Orm().Query().Save(blacklist); err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"blacklist_id": blacklist.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"blacklist": *blacklist,
	})
}

// Destroy 删除黑名单
func (r *BlacklistController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	blacklist, resp := r.findBlacklistByID(ctx, id)
	if resp != nil {
		return resp
	}

	if _, err := facades.Orm().Query().Delete(blacklist); err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"blacklist_id": blacklist.ID,
		})
	}

	return response.Success(ctx)
}
