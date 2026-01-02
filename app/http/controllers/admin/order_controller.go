package admin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/jobs"
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
		// 如果是 TimeRangeError，使用翻译键和参数进行翻译
		if timeRangeErr, ok := err.(*utils.TimeRangeError); ok {
			message := trans.Get(ctx, timeRangeErr.Key)
			// 如果有参数，替换占位符 {key}
			if timeRangeErr.Params != nil {
				for key, value := range timeRangeErr.Params {
					placeholder := fmt.Sprintf("{%s}", key)
					message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
				}
			}
			return services.OrderFilters{}, response.Error(ctx, http.StatusBadRequest, message)
		}
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
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 防重复点击：使用框架自带的原子锁（锁会在10秒后自动过期，防止短时间内重复请求）
	lockKey := fmt.Sprintf("export:orders:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	// 尝试获取锁，如果获取失败则返回错误
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "export_in_progress")
	}

	// 构建筛选条件
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// 创建导出记录（状态为处理中）
	// 获取存储驱动配置
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = utils.GetConfigValue("storage", "export_disk", "")
	}
	if disk == "" {
		disk = "local"
	}

	exportRecord := models.Export{
		AdminID: adminID,
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "", // 处理完成后更新
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 将筛选条件序列化为 JSON
	filtersMap := map[string]any{
		"user_id":    filters.UserID,
		"order_no":   filters.OrderNo,
		"status":     filters.Status,
		"min_amount": filters.MinAmount,
		"max_amount": filters.MaxAmount,
		"order_by":   filters.OrderBy,
	}
	if !filters.StartTime.IsZero() {
		filtersMap["start_time"] = filters.StartTime.Format("2006-01-02 15:04:05")
	}
	if !filters.EndTime.IsZero() {
		filtersMap["end_time"] = filters.EndTime.Format("2006-01-02 15:04:05")
	}

	// 获取当前语言（从请求头或查询参数，与 middleware 逻辑一致）
	lang := r.getCurrentLanguage(ctx)

	// 异步执行导出任务（使用 Job）
	// 将参数序列化为 JSON 字符串传递，避免框架对复杂类型的序列化问题
	exportArgsStruct := jobs.ExportOrdersArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "orders",
		Language: lang,
	}

	// 序列化为 JSON 字符串
	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		facades.Log().Errorf("序列化导出参数失败: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 记录任务提交日志
	facades.Log().Infof("提交导出任务到队列: export_id=%d, queue_driver=%s, args_json=%s",
		exportRecord.ID, facades.Config().GetString("queue.default"), string(exportArgsJSON))

	// 使用 queue.Arg 包装 JSON 字符串参数
	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	// 传递 JSON 字符串作为参数，使用 long-running 队列，避免长时间运行的导出任务影响其他队列任务
	// 所有耗时任务（导出、报表生成、批量处理等）都应该使用 long-running 队列
	if err := facades.Queue().Job(&jobs.ExportOrders{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		// 如果任务提交失败，立即释放锁，让用户可以立即重试
		lock.Release()
		facades.Log().Errorf("提交导出任务失败: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("导出任务已成功提交到队列: export_id=%d", exportRecord.ID)

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "export_task_submitted"),
	})
}

// GetExportStatus 查询导出状态
// @Summary      查询导出状态
// @Description  根据导出记录ID查询导出任务的状态
// @Tags         订单管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "导出记录ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any  "参数错误"
// @Failure      401  {object}  map[string]any  "未登录"
// @Failure      403  {object}  map[string]any  "无权限"
// @Failure      500  {object}  map[string]any  "服务器错误"
// @Router       /api/admin/orders/export/status/{id} [get]
// @Security     BearerAuth
func (r *OrderController) GetExportStatus(ctx http.Context) http.Response {
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "export_id_required")
	}

	exportRecordService := services.NewExportRecordService()
	exportRecord, err := exportRecordService.GetByID(exportID)
	if err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 检查权限：只能查看自己的导出记录
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if exportRecord.AdminID != adminID {
		return response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	// 生成文件URL
	fileURL := ""
	if exportRecord.Path != "" && exportRecord.Status == models.ExportStatusSuccess {
		exportService := services.NewExportService(ctx)
		if exportRecord.Disk == "local" || exportRecord.Disk == "public" {
			fileURL = fmt.Sprintf("/api/admin/exports/%d/download", exportRecord.ID)
		} else {
			fileURL = exportService.GetExportURL(exportRecord.Path)
		}
	}

	return response.Success(ctx, http.Json{
		"id":          exportRecord.ID,
		"status":      exportRecord.Status,
		"status_text": r.getExportStatusText(ctx, exportRecord.Status),
		"file_url":    fileURL,
		"filename":    exportRecord.Filename,
		"size":        exportRecord.Size,
		"error_msg":   exportRecord.ErrorMsg,
		"created_at":  exportRecord.CreatedAt.ToDateTimeString(),
		"updated_at":  exportRecord.UpdatedAt.ToDateTimeString(),
	})
}

func (r *OrderController) getExportStatusText(ctx http.Context, status uint8) string {
	switch status {
	case models.ExportStatusProcessing:
		return trans.Get(ctx, "export_task_status_processing")
	case models.ExportStatusSuccess:
		return trans.Get(ctx, "export_task_status_success")
	case models.ExportStatusFailed:
		return trans.Get(ctx, "export_task_status_failed")
	default:
		return trans.Get(ctx, "export_task_status_unknown")
	}
}

// getCurrentLanguage 获取当前请求的语言（使用通用工具函数）
func (r *OrderController) getCurrentLanguage(ctx http.Context) string {
	return utils.GetCurrentLanguage(ctx)
}

// Import 导入订单
// @Summary      导入订单
// @Description  从CSV文件导入订单数据，支持批量导入
// @Tags         订单管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "CSV文件"
// @Success      200  {object} map[string]any "导入成功，返回导入结果"
// @Failure      400  {object} map[string]any "参数错误"
// @Failure      401  {object} map[string]any "未登录"
// @Failure      403  {object} map[string]any "无权限"
// @Failure      500  {object} map[string]any "服务器错误"
// @Router       /api/admin/orders/import [post]
// @Security     BearerAuth
func (r *OrderController) Import(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 获取上传的文件
	file, err := ctx.Request().File("file")
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "file_required")
	}

	// 验证文件类型（只允许CSV）
	filename := file.GetClientOriginalName()
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidFileType.Code)
	}

	// 读取文件内容
	storage := facades.Storage().Disk("local")
	savedPath, err := storage.PutFile("", file)
	if err != nil {
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
		})
	}

	// 读取文件内容
	csvContent, err := storage.Get(savedPath)
	if err != nil {
		_ = storage.Delete(savedPath)
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
		})
	}

	// 清理临时文件
	defer func() {
		_ = storage.Delete(savedPath)
	}()

	// 导入订单
	importService := services.NewImportOrderService(ctx)
	result, err := importService.ImportOrders(csvContent)
	if err != nil {
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
			"admin_id": adminID,
		})
	}

	return response.Success(ctx, http.Json{
		"total_rows":    result.TotalRows,
		"success_count": result.SuccessCount,
		"failed_count":  result.FailedCount,
		"errors":        result.Errors,
		"message":       trans.Get(ctx, "import_success"),
	})
}
