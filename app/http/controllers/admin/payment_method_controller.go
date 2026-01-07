package admin

import (
	"encoding/json"

	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type PaymentMethodController struct {
	paymentService services.PaymentService
}

func NewPaymentMethodController() *PaymentMethodController {
	return &PaymentMethodController{
		paymentService: services.NewPaymentService(),
	}
}

// Index 支付方式列表
// @Summary      获取支付方式列表
// @Description  分页获取支付方式列表，支持多条件筛选
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        page       query    int     false "页码" default(1)
// @Param        page_size  query    int     false "每页数量" default(10)
// @Param        name       query    string  false "支付方式名称（模糊搜索）"
// @Param        code       query    string  false "支付方式代码"
// @Param        type       query    string  false "支付类型"
// @Param        is_active  query    string  false "是否启用：1-启用，0-禁用"
// @Param        order_by   query    string  false "排序（格式：字段:asc/desc，如：created_at:desc）"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payment-methods [get]
// @Security     BearerAuth
func (r *PaymentMethodController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := services.PaymentMethodFilters{
		Name:        ctx.Request().Query("name", ""),
		Code:        ctx.Request().Query("code", ""),
		Type:        ctx.Request().Query("type", ""),
		IsActive:    ctx.Request().Query("is_active", ""),
		Description: ctx.Request().Query("description", ""),
		OrderBy:     ctx.Request().Query("order_by", ""),
	}

	paymentMethods, total, err := r.paymentService.GetPaymentMethods(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "payment_method", err, map[string]any{
			"filters": filters,
		})
	}

	// 转换响应数据（不返回敏感配置信息）
	paymentMethodList := make([]http.Json, len(paymentMethods))
	for i, pm := range paymentMethods {
		paymentMethodList[i] = http.Json{
			"id":          pm.ID,
			"name":        pm.Name,
			"code":        pm.Code,
			"type":        pm.Type,
			"is_active":   pm.IsActive,
			"sort":        pm.Sort,
			"description": pm.Description,
			"created_at":  pm.CreatedAt,
			"updated_at":  pm.UpdatedAt,
		}
	}

	return response.Success(ctx, http.Json{
		"data":      paymentMethodList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 支付方式详情
// @Summary      获取支付方式详情
// @Description  根据ID获取支付方式详细信息
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        id         path     int     true  "支付方式ID"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      404        {object} map[string]any "支付方式不存在"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payment-methods/{id} [get]
// @Security     BearerAuth
func (r *PaymentMethodController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	paymentMethod, err := r.paymentService.GetPaymentMethodByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrPaymentMethodNotFound.Code)
	}

	// 解析配置 JSON
	var config map[string]any
	if paymentMethod.Config != "" {
		if err := json.Unmarshal([]byte(paymentMethod.Config), &config); err != nil {
			config = make(map[string]any)
		}
	} else {
		config = make(map[string]any)
	}

	return response.Success(ctx, http.Json{
		"id":          paymentMethod.ID,
		"name":        paymentMethod.Name,
		"code":        paymentMethod.Code,
		"type":        paymentMethod.Type,
		"config":      config,
		"is_active":   paymentMethod.IsActive,
		"sort":        paymentMethod.Sort,
		"description": paymentMethod.Description,
		"created_at":  paymentMethod.CreatedAt,
		"updated_at":  paymentMethod.UpdatedAt,
	})
}

// Store 创建支付方式
// @Summary      创建支付方式
// @Description  创建新的支付方式
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        name        body     string  true  "支付方式名称"
// @Param        code        body     string  true  "支付方式代码"
// @Param        type        body     string  true  "支付类型"
// @Param        config      body     object  true  "支付配置(JSON对象)"
// @Param        is_active   body     bool    false "是否启用"
// @Param        sort        body     int     false "排序"
// @Param        description body     string  false "描述"
// @Success      200      {object} map[string]any
// @Failure      400      {object} map[string]any "参数错误"
// @Failure      500      {object} map[string]any "服务器错误"
// @Router       /api/admin/payment-methods [post]
// @Security     BearerAuth
func (r *PaymentMethodController) Store(ctx http.Context) http.Response {
	var req adminrequests.PaymentMethodCreate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	paymentMethod, err := r.paymentService.CreatePaymentMethod(
		req.Name,
		req.Code,
		req.Type,
		req.Config,
		req.IsActive,
		req.Sort,
		req.Description,
	)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.ErrorWithLog(ctx, "payment_method", err, map[string]any{
			"name": req.Name,
			"code": req.Code,
		})
	}

	return response.Success(ctx, http.Json{
		"id":          paymentMethod.ID,
		"name":        paymentMethod.Name,
		"code":        paymentMethod.Code,
		"type":        paymentMethod.Type,
		"is_active":   paymentMethod.IsActive,
		"sort":        paymentMethod.Sort,
		"description": paymentMethod.Description,
		"created_at":  paymentMethod.CreatedAt,
		"updated_at":  paymentMethod.UpdatedAt,
	})
}

// Update 更新支付方式
// @Summary      更新支付方式
// @Description  更新支付方式信息
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        id          path     int     true  "支付方式ID"
// @Param        name        body     string  true  "支付方式名称"
// @Param        config      body     object  false "支付配置(JSON对象)"
// @Param        is_active   body     bool    false "是否启用"
// @Param        sort        body     int     false "排序"
// @Param        description body     string  false "描述"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payment-methods/{id} [put]
// @Security     BearerAuth
func (r *PaymentMethodController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.PaymentMethodUpdate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	err = r.paymentService.UpdatePaymentMethod(
		id,
		req.Name,
		req.Config,
		req.IsActive,
		req.Sort,
		req.Description,
	)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.ErrorWithLog(ctx, "payment_method", err, map[string]any{
			"id": id,
		})
	}

	return response.Success(ctx)
}

// Destroy 删除支付方式
// @Summary      删除支付方式
// @Description  删除支付方式
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        id         path     int     true  "支付方式ID"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payment-methods/{id} [delete]
// @Security     BearerAuth
func (r *PaymentMethodController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

	err := r.paymentService.DeletePaymentMethod(id)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.ErrorWithLog(ctx, "payment_method", err, map[string]any{
			"id": id,
		})
	}

	return response.Success(ctx)
}
