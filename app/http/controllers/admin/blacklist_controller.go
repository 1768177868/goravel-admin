package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

type BlacklistController struct {
	blacklistService services.BlacklistService
}

func NewBlacklistController() *BlacklistController {
	return &BlacklistController{
		blacklistService: services.NewBlacklistService(),
	}
}

// findBlacklistByID 根据ID查找黑名单，如果不存在则返回错误响应
func (r *BlacklistController) findBlacklistByID(ctx http.Context, id uint) (*models.Blacklist, http.Response) {
	blacklist, err := r.blacklistService.GetByID(id)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}
	return blacklist, nil
}

// buildFilters 构建查询过滤器
func (r *BlacklistController) buildFilters(ctx http.Context) services.BlacklistFilters {
	ip := ctx.Request().Query("ip", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.BlacklistFilters{
		IP:        ip,
		Status:    status,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Index 黑名单列表
func (r *BlacklistController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "20"))

	filters := r.buildFilters(ctx)

	blacklists, total, err := r.blacklistService.GetList(filters, page, pageSize)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      blacklists,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
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
	if err := utils.ValidateBlacklistIP(blacklistCreate.IP); err != nil {
		// 使用业务错误类型，直接提取错误码
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		// 如果不是业务错误，返回通用错误
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidIPFormat.Code)
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
		if err := utils.ValidateBlacklistIP(blacklistUpdate.IP); err != nil {
			// 使用业务错误类型，直接提取错误码
			if businessErr, ok := apperrors.GetBusinessError(err); ok {
				return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
			}
			// 如果不是业务错误，返回通用错误
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidIPFormat.Code)
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
