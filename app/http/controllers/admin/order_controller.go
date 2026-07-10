package admin

import (
	"encoding/json"
	"fmt"
	appfacades "goravel/app/facades"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/apidoc"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/jobs"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

type OrderController struct {}


// OrderResponse swagger response for order summary.
type OrderResponse struct {
	ID        uint    `json:"id" example:"1"`                           // Order ID
	OrderNo   string  `json:"order_no" example:"ORD202604090001"`       // Order number
	UserID    uint    `json:"user_id" example:"1001"`                   // User ID
	Amount    float64 `json:"amount" example:"199.98"`                  // Total amount
	Status    string  `json:"status" example:"pending"`                 // Order status
	Remark    string  `json:"remark" example:"note"`                    // Remark
	CreatedAt string  `json:"created_at" example:"2024-01-01 00:00:00"` // Created at
	UpdatedAt string  `json:"updated_at" example:"2024-01-01 00:00:00"` // Updated at
}

// OrderDetailResponse swagger response for order item details.
type OrderDetailResponse struct {
	ID          uint    `json:"id" example:"1"`                           // Detail ID
	OrderID     uint    `json:"order_id" example:"1"`                     // Order ID
	ProductID   uint    `json:"product_id" example:"101"`                 // Product ID
	ProductName string  `json:"product_name" example:"sample product"`    // Product name
	Price       float64 `json:"price" example:"99.99"`                    // Unit price
	Quantity    int     `json:"quantity" example:"2"`                     // Quantity
	Subtotal    float64 `json:"subtotal" example:"199.98"`                // Subtotal
	CreatedAt   string  `json:"created_at" example:"2024-01-01 00:00:00"` // Created at
	UpdatedAt   string  `json:"updated_at" example:"2024-01-01 00:00:00"` // Updated at
}

// OrderWithDetailsResponse combines order and details.
type OrderWithDetailsResponse struct {
	Order   OrderResponse         `json:"order"`   // Order summary
	Details []OrderDetailResponse `json:"details"` // Detail list
}

// OrderListData list response payload.
type OrderListData struct {
	Data []OrderWithDetailsResponse `json:"data"` // Order list
	apidoc.Pagination
}

type OrderListResponse struct {
	apidoc.Success
	Data OrderListData `json:"data"`
}

type OrderDetailData struct {
	Order   OrderResponse         `json:"order"`   // Order summary
	Details []OrderDetailResponse `json:"details"` // Order details
}

type OrderDetailResponseWrapper struct {
	apidoc.Success
	Data OrderDetailData `json:"data"`
}

type OrderCreateRequest struct {
	UserID    uint               `json:"user_id" example:"1001"`
	Amount    float64            `json:"amount" example:"199.98"`
	Products  []OrderProductItem `json:"products"`
	RequestID string             `json:"request_id" example:"req_20260409_001"`
	Remark    string             `json:"remark" example:"remark"`
}

type OrderUpdateRequest struct {
	Status string `json:"status" example:"paid"`
	Remark string `json:"remark" example:"remark"`
}

type ExportTaskData struct {
	ExportID uint   `json:"export_id" example:"1"`
	Message  string `json:"message" example:"message"`
}

type ExportTaskResponse struct {
	apidoc.Success
	Data ExportTaskData `json:"data"`
}

type ExportStatusData struct {
	ID         uint   `json:"id" example:"1"`
	Status     uint8  `json:"status" example:"1"`
	StatusText string `json:"status_text" example:"status_text"`
	FileURL    string `json:"file_url" example:"/api/admin/exports/1/download"`
	Filename   string `json:"filename" example:"orders_20260409.csv"`
	Size       int64  `json:"size" example:"1024"`
	ErrorMsg   string `json:"error_msg" example:""`
	CreatedAt  string `json:"created_at" example:"2024-01-01 00:00:00"`
	UpdatedAt  string `json:"updated_at" example:"2024-01-01 00:00:00"`
}

type ExportStatusResponse struct {
	apidoc.Success
	Data ExportStatusData `json:"data"`
}

type ImportResultData struct {
	TotalRows    int      `json:"total_rows" example:"10"`
	SuccessCount int      `json:"success_count" example:"8"`
	FailedCount  int      `json:"failed_count" example:"2"`
	Errors       []string `json:"errors"`
	Message      string   `json:"message" example:"message"`
}

type ImportResultResponse struct {
	apidoc.Success
	Data ImportResultData `json:"data"`
}

type OrderProductItem struct {
	ProductID   uint    `json:"product_id" example:"1" binding:"required"`
	ProductName string  `json:"product_name" example:"product_name" binding:"required"`
	Price       float64 `json:"price" example:"99.99" binding:"required"`
	Quantity    int     `json:"quantity" example:"2" binding:"required"`
}

func NewOrderController() *OrderController {
	return &OrderController{}
}

func (r *OrderController) orderService(ctx http.Context) services.OrderService {
	return services.NewOrderService(ctx)
}


// buildFilters builds filters shared by list/export endpoints.
// It accepts both query params (GET) and body params (POST).
func (r *OrderController) buildFilters(ctx http.Context) (services.OrderFilters, http.Response) {
	// Prefer request body, fallback to query params for compatibility.
	userID := cast.ToUint(ctx.Request().Input("user_id", ctx.Request().Query("user_id", "0")))
	orderNo := ctx.Request().Input("order_no", ctx.Request().Query("order_no", ""))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	minAmount := cast.ToFloat64(ctx.Request().Input("min_amount", ctx.Request().Query("min_amount", "0")))
	maxAmount := cast.ToFloat64(ctx.Request().Input("max_amount", ctx.Request().Query("max_amount", "0")))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))

	// Parse time params and normalize to UTC time strings.
	startTimeStr := getTimeInputOrQueryUTC(ctx, "start_time")
	endTimeStr := getTimeInputOrQueryUTC(ctx, "end_time")

	startTime, endTime, err := r.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		return services.OrderFilters{}, response.Error(ctx, http.StatusBadRequest, err.Error())
	}

	// Validate time range (default limit: 3 months).
	if resp := validateTimeRangeResponse(ctx, startTime, endTime, 3); resp != nil {
		return services.OrderFilters{}, resp
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

// parseTimeRange parses time range; default start is now - 7 days.
func (r *OrderController) parseTimeRange(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	if startTimeStr == "" {
		// Default to recent 7 days in UTC.
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	} else {
		// Parse UTC datetime string.
		startTime, err = utils.ParseDateTime(startTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_start_time")
		}
	}

	if endTimeStr == "" {
		// Empty end time means no upper bound.
		endTime = time.Time{}
	} else {
		// Parse UTC datetime string.
		endTime, err = utils.ParseDateTime(endTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_end_time")
		}
	}

	return startTime, endTime, nil
}

// formatOrderStatus returns localized status text.
func (r *OrderController) formatOrderStatus(ctx http.Context, status string) string {
	switch status {
	case "pending":
		return trans.Get(ctx, "order_status_pending")
	case "paid":
		return trans.Get(ctx, "order_status_paid")
	case "cancelled":
		return trans.Get(ctx, "order_status_cancelled")
	default:
		return status
	}
}

// formatTime converts several time types to string.
func (r *OrderController) formatTime(t any) string {
	if t == nil {
		return ""
	}

	switch v := t.(type) {
	case time.Time:
		return utils.FormatDateTime(v)
	case *time.Time:
		return utils.FormatDateTimePtr(v)
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
		// Fallback for other types.
		if str := fmt.Sprintf("%v", t); str != "" && str != "<nil>" {
			return str
		}
		return ""
	}
}

// convertOrderToJson converts model.Order to JSON payload.
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

// Index returns paginated order list.
// @Summary      Get order list
// @Description  Returns paginated orders with filters; time range is limited.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        page       query    int     false "Page number" default(1)
// @Param        page_size  query    int     false "Page size" default(10)
// @Param        user_id    query    int     false "User ID"
// @Param        order_no   query    string  false "Order number (fuzzy)"
// @Param        status     query    string  false "Order status"
// @Param        min_amount query    float64 false "Minimum amount"
// @Param        max_amount query    float64 false "Maximum amount"
// @Param        start_time query    string  false "Start time (2006-01-02 15:04:05)"
// @Param        end_time   query    string  false "End time (2006-01-02 15:04:05)"
// @Param        order_by   query    string  false "Sort field:direction"
// @Success      200        {object} OrderListResponse
// @Failure      400        {object} apidoc.Error "Bad request"
// @Failure      500        {object} apidoc.Error "Server error"
// @Router       /api/admin/orders [get]
// @Security     BearerAuth
func (r *OrderController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	// Build filters shared with export.
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// Query order list with details.
	ordersWithDetails, total, err := r.orderService(ctx).GetOrdersWithDetails(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"filters": filters,
		})
	}

	// Build response data.
	orderList := make([]http.Json, len(ordersWithDetails))
	for i, orderWithDetails := range ordersWithDetails {
		orderJson := r.convertOrderToJson(orderWithDetails.Order)
		// Append order details.
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

func (r *OrderController) Show(ctx http.Context) http.Response {
	orderNo := ctx.Request().Query("order_no", "")

	if orderNo != "" {
		order, details, err := r.orderService(ctx).GetOrderByOrderNo(orderNo)
		if err == nil {
			return r.buildOrderDetailResponse(ctx, order, details)
		}
		if routeID := ctx.Request().Route("id"); routeID != "" && orderNo == routeID {
			if orderID := cast.ToUint(routeID); orderID > 0 {
				order, details, err := r.orderService(ctx).GetOrderByID(orderID, time.Time{})
				if err == nil {
					return r.buildOrderDetailResponse(ctx, order, details)
				}
			}
		}
		return response.Error(ctx, http.StatusNotFound, "order_not_found")
	}

	return response.Error(ctx, http.StatusBadRequest, "order_no_or_id_required")
}

func (r *OrderController) buildOrderDetailResponse(ctx http.Context, order *models.Order, details []models.OrderDetail) http.Response {
	orderJson := r.convertOrderToJson(*order)

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

	return response.Success(ctx, http.Json{
		"order":   orderJson,
		"details": detailList,
	})
}

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

	order, details, err := r.orderService(ctx).CreateOrder(req.UserID, req.Amount, req.Products, req.RequestID, req.Remark)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "create_failed")
	}

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

	delete(orderJson, "created_at")
	delete(orderJson, "updated_at")

	return response.Success(ctx, http.Json{
		"order":   orderJson,
		"details": detailList,
	})
}

func (r *OrderController) Update(ctx http.Context) http.Response {
	// 浣跨敤璁㈠崟鍙锋煡璇紙鍙洿鎺ュ畾浣嶅垎琛級
	orderNo := ctx.Request().Query("order_no", "")
	if orderNo == "" {
		return response.Error(ctx, http.StatusBadRequest, "order_no_required")
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Remark string `json:"remark"`
	}

	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_params")
	}

	if err := r.orderService(ctx).UpdateOrderByOrderNo(orderNo, req.Status, req.Remark); err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"order_no": orderNo,
			"status":   req.Status,
			"remark":   req.Remark,
		})
	}

	return response.Success(ctx)
}

// Destroy 鍒犻櫎璁㈠崟
// @Summary      鍒犻櫎璁㈠崟
// @Description  鍒犻櫎璁㈠崟鍙婂叾璇︽儏銆備娇鐢ㄨ鍗曞彿鏌ヨ锛堝彲鐩存帴瀹氫綅鍒嗚〃锛?
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        id         path     string  true "璁㈠崟鍙?
// @Success      200        {object} apidoc.Success
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders/{id} [delete]
// @Security     BearerAuth
func (r *OrderController) Destroy(ctx http.Context) http.Response {
	// 浣跨敤璁㈠崟鍙锋煡璇紙鍙洿鎺ュ畾浣嶅垎琛級
	orderNo := ctx.Request().Query("order_no", "")

	if err := r.orderService(ctx).DeleteOrderByOrderNo(orderNo); err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"order_no": orderNo,
		})
	}

	return response.Success(ctx)
}

// Export creates async export task for order list.
// @Summary      Export order list
// @Description  Create export task by current filters and return export_id.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        user_id    query    int     false "User ID"
// @Param        order_no   query    string  false "Order number"
// @Param        status     query    string  false "Order status"
// @Param        min_amount query    float64 false "Minimum amount"
// @Param        max_amount query    float64 false "Maximum amount"
// @Param        start_time query    string  false "Start time"
// @Param        end_time   query    string  false "End time"
// @Param        order_by   query    string  false "Sort field:direction"
// @Success      200        {object} ExportTaskResponse "Task queued with export_id"
// @Failure      400        {object} apidoc.Error "Bad request"
// @Failure      401        {object} apidoc.Error "Unauthorized"
// @Failure      403        {object} apidoc.Error "Forbidden"
// @Failure      500        {object} apidoc.Error "Server error"
// @Router       /api/admin/orders/export [post]
// @Security     BearerAuth
func (r *OrderController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// Prevent duplicate exports in short period.
	lockKey := fmt.Sprintf("export:orders:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	// Try lock.
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "already_queued")
	}

	// Build filters.
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// Create export record in processing status.
	// Resolve storage disk config.
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = utils.GetConfigValue("storage", "export_disk", "")
	}
	if disk == "" {
		disk = "local"
	}

	exportRecord := models.Export{
		AdminID: adminID,
		Type:    models.ExportTypeOrders,
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "", // Updated when job is finished.
	}
	if err := appfacades.OrmQuery(ctx).Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// Build filter map for job payload.
	filtersMap := map[string]any{
		"user_id":    filters.UserID,
		"order_no":   filters.OrderNo,
		"status":     filters.Status,
		"min_amount": filters.MinAmount,
		"max_amount": filters.MaxAmount,
		"order_by":   filters.OrderBy,
	}
	if !filters.StartTime.IsZero() {
		filtersMap["start_time"] = utils.FormatDateTime(filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		filtersMap["end_time"] = utils.FormatDateTime(filters.EndTime)
	}

	// Resolve language and timezone.
	lang := r.getCurrentLanguage(ctx)
	timezone := helpers.GetCurrentTimezone(ctx)

	// Build async export job args.
	exportArgsStruct := jobs.ExportOrdersArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "orders",
		Language: lang,
		Timezone: timezone,
	}

	// Marshal args as JSON string.
	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		facades.Log().Errorf("Failed to marshal export args: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	// Log dispatch info.
	facades.Log().Infof("Dispatch export task: export_id=%d, queue_driver=%s, args_json=%s",
		exportRecord.ID, facades.Config().GetString("queue.default"), string(exportArgsJSON))

	// Wrap payload as queue args.
	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	// Use long-running queue for export tasks.
	if err := facades.Queue().Job(&jobs.ExportOrders{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		// Release lock immediately when dispatch fails.
		lock.Release()
		facades.Log().Errorf("Failed to dispatch export task: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("Export task dispatched successfully: export_id=%d", exportRecord.ID)

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "queued"),
	})
}

// GetExportStatus returns export task status.
// @Summary      Get export status
// @Description  Query export status by export record ID.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Export record ID"
// @Success      200  {object}  ExportStatusResponse
// @Failure      400  {object}  apidoc.Error  "Bad request"
// @Failure      401  {object}  apidoc.Error  "Unauthorized"
// @Failure      403  {object}  apidoc.Error  "Forbidden"
// @Failure      500  {object}  apidoc.Error  "Server error"
// @Router       /api/admin/orders/export/status/{id} [get]
// @Security     BearerAuth
func (r *OrderController) GetExportStatus(ctx http.Context) http.Response {
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	exportRecordService := services.NewExportRecordService(ctx)
	exportRecord, err := exportRecordService.GetByID(exportID)
	if err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// Permission check: only owner can access.
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if exportRecord.AdminID != adminID {
		return response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	// Build file URL.
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
		return trans.Get(ctx, "processing")
	case models.ExportStatusSuccess:
		return trans.Get(ctx, "success")
	case models.ExportStatusFailed:
		return trans.Get(ctx, "failed")
	default:
		return trans.Get(ctx, "unknown")
	}
}

func (r *OrderController) getCurrentLanguage(ctx http.Context) string {
	return utils.GetCurrentLanguage(ctx)
}

func (r *OrderController) Import(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	file, err := ctx.Request().File("file")
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "file_required")
	}

	filename := file.GetClientOriginalName()
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidFileType.Code)
	}

	storage := facades.Storage().Disk("local")
	savedPath, err := storage.PutFile("", file)
	if err != nil {
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
		})
	}

	csvContent, err := storage.Get(savedPath)
	if err != nil {
		_ = storage.Delete(savedPath)
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
		})
	}

	defer func() {
		_ = storage.Delete(savedPath)
	}()

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
