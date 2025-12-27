package admin

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

type OrderController struct {
	orderService services.OrderService
}

// OrderProductItem 订单商品项（用于 Swagger 文档）
type OrderProductItem struct {
	ProductID   uint    `json:"product_id" example:"1" binding:"required"`      // 商品ID
	ProductName string  `json:"product_name" example:"商品名称" binding:"required"` // 商品名称
	Price       float64 `json:"price" example:"99.99" binding:"required"`       // 单价
	Quantity    int     `json:"quantity" example:"2" binding:"required"`        // 数量
}

func NewOrderController() *OrderController {
	return &OrderController{
		orderService: services.NewOrderService(),
	}
}

// buildFilters 构建筛选条件（列表和导出共用）
// 同时支持查询参数（GET）和请求体参数（POST）
func (r *OrderController) buildFilters(ctx http.Context) (services.OrderFilters, http.Response) {
	// 优先从请求体读取，如果没有则从查询参数读取（兼容 GET 和 POST）
	userID := cast.ToUint(ctx.Request().Input("user_id", ctx.Request().Query("user_id", "0")))
	orderNo := ctx.Request().Input("order_no", ctx.Request().Query("order_no", ""))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	minAmount := cast.ToFloat64(ctx.Request().Input("min_amount", ctx.Request().Query("min_amount", "0")))
	maxAmount := cast.ToFloat64(ctx.Request().Input("max_amount", ctx.Request().Query("max_amount", "0")))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))

	// 解析时间参数（使用 GetTimeQueryParam 处理时区转换）
	// GetTimeQueryParam 会自动从查询参数读取并转换为 UTC 时间字符串
	// 如果查询参数不存在，尝试从请求体读取
	startTimeStr := ctx.Request().Query("start_time", "")
	if startTimeStr == "" {
		startTimeStr = ctx.Request().Input("start_time", "")
	}

	endTimeStr := ctx.Request().Query("end_time", "")
	if endTimeStr == "" {
		endTimeStr = ctx.Request().Input("end_time", "")
	}

	startTime, endTime, err := r.parseTimeRange(ctx, startTimeStr, endTimeStr)
	if err != nil {
		return services.OrderFilters{}, response.Error(ctx, http.StatusBadRequest, err.Error())
	}

	// 验证时间范围（订单查询限制为3个月，可通过配置修改）
	valid, err := utils.ValidateTimeRange(startTime, endTime, 3)
	if !valid {
		return services.OrderFilters{}, response.Error(ctx, http.StatusBadRequest, err.Error())
	}

	return services.OrderFilters{
		UserID:    userID,
		OrderNo:   orderNo,
		Status:    status,
		MinAmount: minAmount,
		MaxAmount: maxAmount,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}, nil
}

// parseTimeRange 解析时间范围（默认最近1周）
func (r *OrderController) parseTimeRange(ctx http.Context, startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	if startTimeStr == "" {
		// 默认查询最近1周（UTC 时间）
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	} else {
		// 使用 ConvertTimeToUTC 处理时区转换（将本地时区转换为 UTC）
		utcTimeStr := helpers.ConvertTimeToUTC(ctx, startTimeStr)
		if utcTimeStr == "" {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_start_time")
		}
		// 解析 UTC 时间字符串
		startTime, err = time.Parse("2006-01-02 15:04:05", utcTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_start_time")
		}
	}

	if endTimeStr == "" {
		// 默认结束时间为当前时间（UTC）
		endTime = time.Now().UTC()
	} else {
		// 使用 ConvertTimeToUTC 处理时区转换（将本地时区转换为 UTC）
		utcTimeStr := helpers.ConvertTimeToUTC(ctx, endTimeStr)
		if utcTimeStr == "" {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_end_time")
		}
		// 解析 UTC 时间字符串
		endTime, err = time.Parse("2006-01-02 15:04:05", utcTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_end_time")
		}
	}

	return startTime, endTime, nil
}

// parseOrderTime 解析订单时间参数（保留以兼容旧接口，但不再使用）
func (r *OrderController) parseOrderTime(ctx http.Context) (time.Time, http.Response) {
	// 保留此方法以兼容旧接口，实际不再使用
	return time.Time{}, nil
}

// formatOrderStatus 格式化订单状态文本
func (r *OrderController) formatOrderStatus(ctx http.Context, status string) string {
	switch status {
	case "pending":
		return trans.Get(ctx, "export_order_status_pending")
	case "paid":
		return trans.Get(ctx, "export_order_status_paid")
	case "cancelled":
		return trans.Get(ctx, "export_order_status_cancelled")
	default:
		return status
	}
}

// formatTime 格式化时间为字符串（支持 time.Time 和 carbon.DateTime）
func (r *OrderController) formatTime(t any) string {
	if t == nil {
		return ""
	}

	switch v := t.(type) {
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02 15:04:05")
	case *time.Time:
		if v == nil || v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02 15:04:05")
	case carbon.DateTime:
		if v.IsZero() {
			return ""
		}
		return v.ToDateTimeString()
	case *carbon.DateTime:
		if v == nil || v.IsZero() {
			return ""
		}
		return v.ToDateTimeString()
	default:
		// 尝试转换为字符串（其他类型）
		if str := fmt.Sprintf("%v", t); str != "" && str != "<nil>" {
			return str
		}
		return ""
	}
}

// convertOrderToJson 转换订单为响应格式
func (r *OrderController) convertOrderToJson(order models.Order) http.Json {
	return http.Json{
		"id":         order.ID,
		"order_no":   order.OrderNo,
		"user_id":    order.UserID,
		"amount":     order.Amount,
		"status":     order.Status,
		"remark":     order.Remark,
		"created_at": order.CreatedAt,
		"updated_at": order.UpdatedAt,
	}
}

// Index 订单列表
// @Summary      获取订单列表
// @Description  分页获取订单列表，支持多条件筛选，查询时间范围不能超过3个月
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        page       query    int     false "页码" default(1)
// @Param        page_size  query    int     false "每页数量" default(10)
// @Param        user_id    query    int     false "用户ID"
// @Param        order_no   query    string  false "订单号（模糊搜索）"
// @Param        status     query    string  false "订单状态（pending/paid/cancelled）"
// @Param        min_amount query    float64 false "最小金额"
// @Param        max_amount query    float64 false "最大金额"
// @Param        start_time query    string  false "开始时间（格式：2006-01-02 15:04:05）"
// @Param        end_time   query    string  false "结束时间（格式：2006-01-02 15:04:05）"
// @Param        order_by   query    string  false "排序（格式：字段:asc/desc，如：created_at:desc）"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/orders [get]
// @Security     BearerAuth
func (r *OrderController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	// 构建筛选条件（列表和导出共用）
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// 查询订单（包含详情）
	ordersWithDetails, total, err := r.orderService.GetOrdersWithDetails(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"filters": filters,
		})
	}

	// 转换响应数据
	orderList := make([]http.Json, len(ordersWithDetails))
	for i, orderWithDetails := range ordersWithDetails {
		orderJson := r.convertOrderToJson(orderWithDetails.Order)
		// 添加订单详情
		detailsList := make([]http.Json, len(orderWithDetails.Details))
		for j, detail := range orderWithDetails.Details {
			detailsList[j] = http.Json{
				"id":           detail.ID,
				"order_id":     detail.OrderID,
				"product_id":   detail.ProductID,
				"product_name": detail.ProductName,
				"price":        detail.Price,
				"quantity":     detail.Quantity,
				"subtotal":     detail.Subtotal,
				"created_at":   detail.CreatedAt,
				"updated_at":   detail.UpdatedAt,
			}
		}
		orderJson["details"] = detailsList
		orderList[i] = orderJson
	}

	return response.Success(ctx, http.Json{
		"data":      orderList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 订单详情
// @Summary      获取订单详情
// @Description  根据ID获取订单详细信息，返回订单主表数据和订单详情表数据（支持分表查询）
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        id         path     int     true  "订单ID"
// @Success      200        {object} map[string]any "返回数据包含 order（订单主表）和 details（订单详情表数组）"
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      404        {object} map[string]any "订单不存在"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/orders/{id} [get]
// @Security     BearerAuth
func (r *OrderController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	// GetOrderByID 会自动根据订单ID查找对应的分表，不再需要 orderTime 参数
	order, details, err := r.orderService.GetOrderByID(id, time.Time{})
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, "order_not_found")
	}

	// 转换订单主表数据（使用统一的方法）
	orderJson := r.convertOrderToJson(*order)

	// 转换订单详情数据
	detailList := make([]http.Json, len(details))
	for i, detail := range details {
		detailList[i] = http.Json{
			"id":           detail.ID,
			"order_id":     detail.OrderID,
			"product_id":   detail.ProductID,
			"product_name": detail.ProductName,
			"price":        detail.Price,
			"quantity":     detail.Quantity,
			"subtotal":     detail.Subtotal,
			"created_at":   detail.CreatedAt,
			"updated_at":   detail.UpdatedAt,
		}
	}

	// 返回主表和详情表数据
	return response.Success(ctx, http.Json{
		"order":   orderJson,
		"details": detailList,
	})
}

// Store 创建订单
// @Summary      创建订单
// @Description  创建新订单，自动防止重复提交
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        user_id  body     int                true  "用户ID"
// @Param        amount   body     float64            true  "订单金额"
// @Param        products body     []OrderProductItem true "商品列表"
// @Param        request_id body   string             false "请求ID（用于防重复提交，不传则自动生成）"
// @Success      200      {object} map[string]any
// @Failure      400      {object} map[string]any "参数错误或重复提交"
// @Failure      500      {object} map[string]any "服务器错误"
// @Router       /api/admin/orders [post]
// @Security     BearerAuth
func (r *OrderController) Store(ctx http.Context) http.Response {
	var req struct {
		UserID    uint                    `json:"user_id" binding:"required"`
		Amount    float64                 `json:"amount" binding:"required"`
		Products  []services.OrderProduct `json:"products" binding:"required"`
		RequestID string                  `json:"request_id"`
		Remark    string                  `json:"remark"`
	}

	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_params")
	}

	if len(req.Products) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "empty_products")
	}

	// 创建订单
	order, details, err := r.orderService.CreateOrder(req.UserID, req.Amount, req.Products, req.RequestID, req.Remark)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "create_failed")
	}

	// 转换订单详情
	detailList := make([]http.Json, len(details))
	for i, detail := range details {
		detailList[i] = http.Json{
			"id":           detail.ID,
			"order_id":     detail.OrderID,
			"product_id":   detail.ProductID,
			"product_name": detail.ProductName,
			"price":        detail.Price,
			"quantity":     detail.Quantity,
			"subtotal":     detail.Subtotal,
		}
	}

	orderJson := r.convertOrderToJson(*order)
	// 移除不需要的字段（创建订单时不需要返回 created_at 和 updated_at）
	delete(orderJson, "created_at")
	delete(orderJson, "updated_at")

	return response.Success(ctx, http.Json{
		"order":   orderJson,
		"details": detailList,
	})
}

// Update 更新订单
// @Summary      更新订单
// @Description  更新订单信息（主要是状态）
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        id         path     int     true  "订单ID"
// @Param        status     body     string  true  "订单状态"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/orders/{id} [put]
// @Security     BearerAuth
func (r *OrderController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	orderTime, resp := r.parseOrderTime(ctx)
	if resp != nil {
		return resp
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Remark string `json:"remark"`
	}

	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_params")
	}

	if err := r.orderService.UpdateOrder(id, orderTime, req.Status, req.Remark); err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"order_id": id,
			"status":   req.Status,
			"remark":   req.Remark,
		})
	}

	return response.Success(ctx)
}

// Destroy 删除订单
// @Summary      删除订单
// @Description  删除订单及其详情
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        id         path     int     true  "订单ID"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/orders/{id} [delete]
// @Security     BearerAuth
func (r *OrderController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	orderTime, resp := r.parseOrderTime(ctx)
	if resp != nil {
		return resp
	}

	if err := r.orderService.DeleteOrder(id, orderTime); err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"order_id": id,
		})
	}

	return response.Success(ctx)
}

// Export 导出订单列表
// @Summary      导出订单列表
// @Description  根据筛选条件导出订单列表为CSV文件，支持与列表查询相同的筛选条件，查询时间范围不能超过3个月
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        user_id    query    int     false "用户ID"
// @Param        order_no   query    string  false "订单号（模糊搜索）"
// @Param        status     query    string  false "订单状态（pending/paid/cancelled）"
// @Param        min_amount query    float64 false "最小金额"
// @Param        max_amount query    float64 false "最大金额"
// @Param        start_time query    string  false "开始时间（格式：2006-01-02 15:04:05）"
// @Param        end_time   query    string  false "结束时间（格式：2006-01-02 15:04:05）"
// @Param        order_by   query    string  false "排序（格式：字段:asc/desc，如：created_at:desc）"
// @Success      200        {object} map[string]any "导出成功，返回文件下载信息"
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      401        {object} map[string]any "未登录"
// @Failure      403        {object} map[string]any "无权限"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/orders/export [post]
// @Security     BearerAuth
func (r *OrderController) Export(ctx http.Context) http.Response {
	// 构建筛选条件（列表和导出共用）
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// 获取所有订单（不分页）
	orders, err := r.orderService.GetAllOrdersForExport(filters)
	if err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"filters": filters,
		})
	}

	// 准备表头（使用翻译键）
	headers := []string{
		"export_header_id",
		"export_header_order_no",
		"export_header_user_id",
		"export_header_amount",
		"export_header_status",
		"export_header_remark",
		"export_header_created_at",
		"export_header_updated_at",
	}

	// 准备数据
	var data [][]string
	for _, order := range orders {
		row := []string{
			cast.ToString(order.ID),
			order.OrderNo,
			cast.ToString(order.UserID),
			fmt.Sprintf("%.2f", order.Amount),
			r.formatOrderStatus(ctx, order.Status),
			order.Remark,
			r.formatTime(order.CreatedAt),
			r.formatTime(order.UpdatedAt),
		}
		data = append(data, row)
	}

	return response.Export(ctx, "export_success", headers, data, "orders")
}
