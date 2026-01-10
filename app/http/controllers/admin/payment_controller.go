package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/services"
	"goravel/app/utils"
)

type PaymentController struct {
	paymentService services.PaymentService
}

func NewPaymentController() *PaymentController {
	return &PaymentController{
		paymentService: services.NewPaymentService(),
	}
}

// buildFilters 构建筛选条件（列表和导出共用）
func (r *PaymentController) buildFilters(ctx http.Context) (services.PaymentFilters, http.Response) {
	paymentNo := ctx.Request().Input("payment_no", ctx.Request().Query("payment_no", ""))
	orderNo := ctx.Request().Input("order_no", ctx.Request().Query("order_no", ""))
	paymentMethodID := cast.ToUint(ctx.Request().Input("payment_method_id", ctx.Request().Query("payment_method_id", "0")))
	userID := cast.ToUint(ctx.Request().Input("user_id", ctx.Request().Query("user_id", "0")))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))

	// 解析时间参数
	startTimeStr := ctx.Request().Query("start_time", "")
	if startTimeStr == "" {
		startTimeStr = ctx.Request().Input("start_time", "")
	}

	endTimeStr := ctx.Request().Query("end_time", "")
	if endTimeStr == "" {
		endTimeStr = ctx.Request().Input("end_time", "")
	}

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		utcTimeStr := helpers.ConvertTimeToUTC(ctx, startTimeStr)
		if utcTimeStr == "" {
			return services.PaymentFilters{}, response.Error(ctx, http.StatusBadRequest, "invalid_start_time")
		}
		startTime, err = utils.ParseDateTime(utcTimeStr)
		if err != nil {
			return services.PaymentFilters{}, response.Error(ctx, http.StatusBadRequest, "invalid_start_time")
		}
	}

	if endTimeStr != "" {
		utcTimeStr := helpers.ConvertTimeToUTC(ctx, endTimeStr)
		if utcTimeStr == "" {
			return services.PaymentFilters{}, response.Error(ctx, http.StatusBadRequest, "invalid_end_time")
		}
		endTime, err = utils.ParseDateTime(utcTimeStr)
		if err != nil {
			return services.PaymentFilters{}, response.Error(ctx, http.StatusBadRequest, "invalid_end_time")
		}
	}

	return services.PaymentFilters{
		PaymentNo:       paymentNo,
		OrderNo:         orderNo,
		PaymentMethodID: paymentMethodID,
		UserID:          userID,
		Status:          status,
		StartTime:       startTime,
		EndTime:         endTime,
		OrderBy:         orderBy,
	}, nil
}

// Index 支付记录列表
// @Summary      获取支付记录列表
// @Description  分页获取支付记录列表，支持多条件筛选
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        page             query    int     false "页码" default(1)
// @Param        page_size        query    int     false "每页数量" default(10)
// @Param        payment_no        query    string  false "支付单号（模糊搜索）"
// @Param        order_no          query    string  false "订单号（模糊搜索）"
// @Param        payment_method_id query    int     false "支付方式ID"
// @Param        user_id          query    int     false "用户ID"
// @Param        status           query    string  false "支付状态（pending/paid/failed/cancelled）"
// @Param        start_time       query    string  false "开始时间（格式：2006-01-02 15:04:05）"
// @Param        end_time         query    string  false "结束时间（格式：2006-01-02 15:04:05）"
// @Param        order_by         query    string  false "排序（格式：字段:asc/desc，如：created_at:desc）"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payments [get]
// @Security     BearerAuth
func (r *PaymentController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	payments, total, err := r.paymentService.GetPayments(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "payment", err, map[string]any{
			"filters": filters,
		})
	}

	// 转换响应数据
	paymentList := make([]http.Json, len(payments))
	for i, payment := range payments {
		paymentJson := http.Json{
			"id":                payment.ID,
			"payment_no":        payment.PaymentNo,
			"order_no":          payment.OrderNo,
			"payment_method_id": payment.PaymentMethodID,
			"user_id":           payment.UserID,
			"amount":            payment.Amount,
			"status":            payment.Status,
			"third_party_no":    payment.ThirdPartyNo,
			"pay_time":          r.formatPayTime(payment.PayTime),
			"fail_reason":       payment.FailReason,
			"remark":            payment.Remark,
			"created_at":        payment.CreatedAt,
			"updated_at":        payment.UpdatedAt,
		}

		// 添加支付方式信息
		if payment.PaymentMethod.ID > 0 {
			paymentJson["payment_method"] = http.Json{
				"id":   payment.PaymentMethod.ID,
				"name": payment.PaymentMethod.Name,
				"code": payment.PaymentMethod.Code,
				"type": payment.PaymentMethod.Type,
			}
		}

		paymentList[i] = paymentJson
	}

	return response.Success(ctx, http.Json{
		"data":      paymentList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 支付记录详情
// @Summary      获取支付记录详情
// @Description  根据ID获取支付记录详细信息
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        id         path     int     true  "支付记录ID"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      404        {object} map[string]any "支付记录不存在"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payments/{id} [get]
// @Security     BearerAuth
func (r *PaymentController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	payment, err := r.paymentService.GetPaymentByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrPaymentNotFound.Code)
	}

	paymentJson := http.Json{
		"id":                payment.ID,
		"payment_no":        payment.PaymentNo,
		"order_no":          payment.OrderNo,
		"payment_method_id": payment.PaymentMethodID,
		"user_id":           payment.UserID,
		"amount":            payment.Amount,
		"status":            payment.Status,
		"third_party_no":    payment.ThirdPartyNo,
		"pay_time":          r.formatPayTime(payment.PayTime),
		"fail_reason":       payment.FailReason,
		"remark":            payment.Remark,
		"created_at":        payment.CreatedAt,
		"updated_at":        payment.UpdatedAt,
	}

	// 添加支付方式信息
	if payment.PaymentMethod.ID > 0 {
		paymentJson["payment_method"] = http.Json{
			"id":   payment.PaymentMethod.ID,
			"name": payment.PaymentMethod.Name,
			"code": payment.PaymentMethod.Code,
			"type": payment.PaymentMethod.Type,
		}
	}

	return response.Success(ctx, paymentJson)
}

// formatPayTime 格式化支付时间为字符串
func (r *PaymentController) formatPayTime(t *time.Time) string {
	return utils.FormatDateTimePtr(t)
}
