package admin

import (
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/apidoc"
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

type BlacklistResponse struct {
	ID        uint   `json:"id" example:"1"`                           // 黑名单ID
	IP        string `json:"ip" example:"192.168.1.1"`                 // IP地址/IP段
	Remark    string `json:"remark" example:"测试IP"`                    // 备注说明
	Status    uint8  `json:"status" enums:"0,1" example:"1"`           // 状态（1-启用，0-禁用）
	CreatedAt string `json:"created_at" example:"2024-01-01 00:00:00"` // 创建时间
	UpdatedAt string `json:"updated_at" example:"2024-01-01 00:00:00"` // 更新时间
}

type BlacklistListData struct {
	List []BlacklistResponse `json:"list"`
	apidoc.Pagination
}

type BlacklistListResponse struct {
	apidoc.Success
	Data BlacklistListData `json:"data"`
}

type BlacklistDetailData struct {
	Blacklist BlacklistResponse `json:"blacklist"`
}

type BlacklistDetailResponse struct {
	apidoc.Success
	Data BlacklistDetailData `json:"data"`
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
	startTime := getTimeQueryUTC(ctx, "start_time")
	endTime := getTimeQueryUTC(ctx, "end_time")
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
// @Summary      获取黑名单列表
// @Description  分页获取黑名单，支持按IP、状态、时间区间筛选
// @Tags         风控管理
// @Accept       json
// @Produce      json
// @Param        page        query     int     false  "页码（从1开始）" default(1)
// @Param        page_size   query     int     false  "每页数量（建议 10-100）" default(10)
// @Param        ip          query     string  false  "IP地址或IP段（模糊匹配）"
// @Param        status      query     string  false  "状态（1-启用，0-禁用）" Enums(0,1)
// @Param        start_time  query     string  false  "开始时间（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        end_time    query     string  false  "结束时间（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        order_by    query     string  false  "排序（格式：字段:asc/desc，例如：created_at:desc）"
// @Success      200         {object}  BlacklistListResponse
// @Failure      500         {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/blacklists [get]
// @Security     BearerAuth
func (r *BlacklistController) Index(ctx http.Context) http.Response {
	page, pageSize := helpers.PaginationFromQuery(ctx, helpers.PaginationLimits{})

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
// @Summary      获取黑名单详情
// @Description  根据ID获取黑名单详情
// @Tags         风控管理
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "黑名单ID"
// @Success      200   {object}  BlacklistDetailResponse
// @Failure      404   {object}  apidoc.Error "黑名单不存在"
// @Router       /api/admin/blacklists/{id} [get]
// @Security     BearerAuth
func (r *BlacklistController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	blacklist, resp := r.findBlacklistByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"blacklist": *blacklist,
	})
}

// Store 创建黑名单
// @Summary      创建黑名单
// @Description  创建黑名单记录，支持单个IP、CIDR、IP范围等格式
// @Description  字段说明：ip-IP地址/IP段（必填）；remark-备注；status-状态（1启用/0禁用）
// @Tags         风控管理
// @Accept       json
// @Produce      json
// @Param        request  body      adminrequests.BlacklistCreate  true  "创建参数"
// @Success      200      {object}  BlacklistDetailResponse
// @Failure      400      {object}  apidoc.Error "参数错误或IP格式错误"
// @Failure      500      {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/blacklists [post]
// @Security     BearerAuth
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

	blacklist, err := r.blacklistService.Create(
		blacklistCreate.IP,
		blacklistCreate.Remark,
		blacklistCreate.Status,
	)
	if err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"ip": blacklistCreate.IP,
		})
	}

	return response.Success(ctx, http.Json{
		"blacklist": blacklist,
	})
}

// Update 更新黑名单
// @Summary      更新黑名单
// @Description  根据ID更新黑名单信息
// @Description  字段说明：ip-IP地址/IP段；remark-备注；status-状态（1启用/0禁用）（均可选）
// @Tags         风控管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                           true  "黑名单ID"
// @Param        request  body      adminrequests.BlacklistUpdate true  "更新参数"
// @Success      200      {object}  BlacklistDetailResponse
// @Failure      400      {object}  apidoc.Error "参数错误或IP格式错误"
// @Failure      404      {object}  apidoc.Error "黑名单不存在"
// @Failure      500      {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/blacklists/{id} [put]
// @Security     BearerAuth
func (r *BlacklistController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
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

	if err := r.blacklistService.Update(blacklist); err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"blacklist_id": blacklist.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"blacklist": *blacklist,
	})
}

// Destroy 删除黑名单
// @Summary      删除黑名单
// @Description  根据ID删除黑名单记录
// @Tags         风控管理
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "黑名单ID"
// @Success      200   {object}  apidoc.Success
// @Failure      404   {object}  apidoc.Error "黑名单不存在"
// @Failure      500   {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/blacklists/{id} [delete]
// @Security     BearerAuth
func (r *BlacklistController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	blacklist, resp := r.findBlacklistByID(ctx, id)
	if resp != nil {
		return resp
	}

	if err := r.blacklistService.Delete(blacklist); err != nil {
		return response.ErrorWithLog(ctx, "blacklist", err, map[string]any{
			"blacklist_id": blacklist.ID,
		})
	}

	return response.Success(ctx)
}
