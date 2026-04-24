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
	"goravel/app/http/apidoc"
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

// OrderResponse 璁㈠崟涓讳俊鎭?
type OrderResponse struct {
	ID        uint    `json:"id" example:"1"`                           // 璁㈠崟ID
	OrderNo   string  `json:"order_no" example:"ORD202604090001"`       // 璁㈠崟鍙?
	UserID    uint    `json:"user_id" example:"1001"`                   // 鐢ㄦ埛ID
	Amount    float64 `json:"amount" example:"199.98"`                  // 璁㈠崟閲戦
	Status    string  `json:"status" example:"pending"`                 // 璁㈠崟鐘舵€?
	Remark    string  `json:"remark" example:"澶囨敞淇℃伅"`                  // 澶囨敞
	CreatedAt string  `json:"created_at" example:"2024-01-01 00:00:00"` // 鍒涘缓鏃堕棿
	UpdatedAt string  `json:"updated_at" example:"2024-01-01 00:00:00"` // 鏇存柊鏃堕棿
}

// OrderDetailResponse 璁㈠崟璇︽儏椤?
type OrderDetailResponse struct {
	ID          uint    `json:"id" example:"1"`                           // 璇︽儏ID
	OrderID     uint    `json:"order_id" example:"1"`                     // 璁㈠崟ID
	ProductID   uint    `json:"product_id" example:"101"`                 // 鍟嗗搧ID
	ProductName string  `json:"product_name" example:"鍟嗗搧鍚嶇О"`            // 鍟嗗搧鍚嶇О
	Price       float64 `json:"price" example:"99.99"`                    // 鍗曚环
	Quantity    int     `json:"quantity" example:"2"`                     // 鏁伴噺
	Subtotal    float64 `json:"subtotal" example:"199.98"`                // 灏忚
	CreatedAt   string  `json:"created_at" example:"2024-01-01 00:00:00"` // 鍒涘缓鏃堕棿
	UpdatedAt   string  `json:"updated_at" example:"2024-01-01 00:00:00"` // 鏇存柊鏃堕棿
}

// OrderWithDetailsResponse 璁㈠崟鍙婅鎯?
type OrderWithDetailsResponse struct {
	Order   OrderResponse         `json:"order"`   // 璁㈠崟涓讳俊鎭?
	Details []OrderDetailResponse `json:"details"` // 璁㈠崟璇︽儏鍒楄〃
}

// OrderListData 璁㈠崟鍒楄〃鍝嶅簲 data
type OrderListData struct {
	Data []OrderWithDetailsResponse `json:"data"` // 璁㈠崟鍒楄〃
	apidoc.Pagination
}

type OrderListResponse struct {
	apidoc.Success
	Data OrderListData `json:"data"`
}

type OrderDetailData struct {
	Order   OrderResponse         `json:"order"`   // 璁㈠崟涓讳俊鎭?
	Details []OrderDetailResponse `json:"details"` // 璁㈠崟璇︽儏
}

type OrderDetailResponseWrapper struct {
	apidoc.Success
	Data OrderDetailData `json:"data"`
}

type OrderCreateRequest struct {
	UserID    uint               `json:"user_id" example:"1001"`                // 鐢ㄦ埛ID锛堝繀濉級
	Amount    float64            `json:"amount" example:"199.98"`               // 璁㈠崟閲戦锛堝繀濉級
	Products  []OrderProductItem `json:"products"`                              // 鍟嗗搧鍒楄〃锛堝繀濉級
	RequestID string             `json:"request_id" example:"req_20260409_001"` // 骞傜瓑璇锋眰ID锛堝彲閫夛級
	Remark    string             `json:"remark" example:"澶囨敞淇℃伅"`               // 澶囨敞锛堝彲閫夛級
}

type OrderUpdateRequest struct {
	Status string `json:"status" example:"paid"`    // 璁㈠崟鐘舵€侊紙蹇呭～锛?
	Remark string `json:"remark" example:"宸插畬鎴愭敮浠?` // 澶囨敞锛堝彲閫夛級
}

type ExportTaskData struct {
	ExportID uint   `json:"export_id" example:"1"`        // 瀵煎嚭璁板綍ID
	Message  string `json:"message" example:"瀵煎嚭浠诲姟宸叉彁浜?` // 鎻愮ず淇℃伅
}

type ExportTaskResponse struct {
	apidoc.Success
	Data ExportTaskData `json:"data"`
}

type ExportStatusData struct {
	ID         uint   `json:"id" example:"1"`                                   // 瀵煎嚭璁板綍ID
	Status     uint8  `json:"status" example:"1"`                               // 鐘舵€佺爜
	StatusText string `json:"status_text" example:"澶勭悊涓?`                       // 鐘舵€佹枃鏈?
	FileURL    string `json:"file_url" example:"/api/admin/exports/1/download"` // 涓嬭浇鍦板潃
	Filename   string `json:"filename" example:"orders_20260409.csv"`           // 鏂囦欢鍚?
	Size       int64  `json:"size" example:"1024"`                              // 鏂囦欢澶у皬锛堝瓧鑺傦級
	ErrorMsg   string `json:"error_msg" example:""`                             // 閿欒淇℃伅
	CreatedAt  string `json:"created_at" example:"2024-01-01 00:00:00"`         // 鍒涘缓鏃堕棿
	UpdatedAt  string `json:"updated_at" example:"2024-01-01 00:00:00"`         // 鏇存柊鏃堕棿
}

type ExportStatusResponse struct {
	apidoc.Success
	Data ExportStatusData `json:"data"`
}

type ImportResultData struct {
	TotalRows    int      `json:"total_rows" example:"10"`   // 鎬昏鏁?
	SuccessCount int      `json:"success_count" example:"8"` // 鎴愬姛鏁伴噺
	FailedCount  int      `json:"failed_count" example:"2"`  // 澶辫触鏁伴噺
	Errors       []string `json:"errors"`                    // 澶辫触鍘熷洜鍒楄〃
	Message      string   `json:"message" example:"瀵煎叆鎴愬姛"`  // 鎻愮ず淇℃伅
}

type ImportResultResponse struct {
	apidoc.Success
	Data ImportResultData `json:"data"`
}

// OrderProductItem 璁㈠崟鍟嗗搧椤癸紙鐢ㄤ簬 Swagger 鏂囨。锛?
type OrderProductItem struct {
	ProductID   uint    `json:"product_id" example:"1" binding:"required"`        // 鍟嗗搧ID
	ProductName string  `json:"product_name" example:"鍟嗗搧鍚嶇О" binding:"required"` // 鍟嗗搧鍚嶇О
	Price       float64 `json:"price" example:"99.99" binding:"required"`         // 鍗曚环
	Quantity    int     `json:"quantity" example:"2" binding:"required"`          // 鏁伴噺
}

func NewOrderController() *OrderController {
	return &OrderController{
		orderService: services.NewOrderService(),
	}
}

// buildFilters 鏋勫缓绛涢€夋潯浠讹紙鍒楄〃鍜屽鍑哄叡鐢級
// 鍚屾椂鏀寔鏌ヨ鍙傛暟锛圙ET锛夊拰璇锋眰浣撳弬鏁帮紙POST锛?
func (r *OrderController) buildFilters(ctx http.Context) (services.OrderFilters, http.Response) {
	// 浼樺厛浠庤姹備綋璇诲彇锛屽鏋滄病鏈夊垯浠庢煡璇㈠弬鏁拌鍙栵紙鍏煎 GET 鍜?POST锛?
	userID := cast.ToUint(ctx.Request().Input("user_id", ctx.Request().Query("user_id", "0")))
	orderNo := ctx.Request().Input("order_no", ctx.Request().Query("order_no", ""))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	minAmount := cast.ToFloat64(ctx.Request().Input("min_amount", ctx.Request().Query("min_amount", "0")))
	maxAmount := cast.ToFloat64(ctx.Request().Input("max_amount", ctx.Request().Query("max_amount", "0")))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))

	// 瑙ｆ瀽鏃堕棿鍙傛暟锛坬uery 浼樺厛锛宨nput 鍏滃簳锛夛紝骞剁粺涓€杞崲涓?UTC 鏃堕棿瀛楃涓?
	startTimeStr := getTimeInputOrQueryUTC(ctx, "start_time")
	endTimeStr := getTimeInputOrQueryUTC(ctx, "end_time")

	startTime, endTime, err := r.parseTimeRange(startTimeStr, endTimeStr)
	if err != nil {
		return services.OrderFilters{}, response.Error(ctx, http.StatusBadRequest, err.Error())
	}

	// 楠岃瘉鏃堕棿鑼冨洿锛堣鍗曟煡璇㈤檺鍒朵负3涓湀锛屽彲閫氳繃閰嶇疆淇敼锛?
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

// parseTimeRange 瑙ｆ瀽鏃堕棿鑼冨洿锛堥粯璁ゆ渶杩?鍛級
func (r *OrderController) parseTimeRange(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	if startTimeStr == "" {
		// 榛樿鏌ヨ鏈€杩?鍛紙UTC 鏃堕棿锛?
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	} else {
		// 瑙ｆ瀽 UTC 鏃堕棿瀛楃涓?
		startTime, err = utils.ParseDateTime(startTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_start_time")
		}
	}

	if endTimeStr == "" {
		// 涓嶄紶缁撴潫鏃堕棿鍒欎笉闄愬埗锛岃繑鍥為浂鍊硷紙WHERE 鏉′欢涓細璺宠繃锛?
		endTime = time.Time{}
	} else {
		// 瑙ｆ瀽 UTC 鏃堕棿瀛楃涓?
		endTime, err = utils.ParseDateTime(endTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid_end_time")
		}
	}

	return startTime, endTime, nil
}

// formatOrderStatus 鏍煎紡鍖栬鍗曠姸鎬佹枃鏈?
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

// formatTime 鏍煎紡鍖栨椂闂翠负瀛楃涓诧紙鏀寔 time.Time 鍜?carbon.DateTime锛?
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
		// 灏濊瘯杞崲涓哄瓧绗︿覆锛堝叾浠栫被鍨嬶級
		if str := fmt.Sprintf("%v", t); str != "" && str != "<nil>" {
			return str
		}
		return ""
	}
}

// convertOrderToJson 杞崲璁㈠崟涓哄搷搴旀牸寮?
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

// Index 璁㈠崟鍒楄〃
// @Summary      鑾峰彇璁㈠崟鍒楄〃
// @Description  鍒嗛〉鑾峰彇璁㈠崟鍒楄〃锛屾敮鎸佸鏉′欢绛涢€夛紝鏌ヨ鏃堕棿鑼冨洿涓嶈兘瓒呰繃3涓湀
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        page       query    int     false "椤电爜" default(1)
// @Param        page_size  query    int     false "姣忛〉鏁伴噺" default(10)
// @Param        user_id    query    int     false "鐢ㄦ埛ID"
// @Param        order_no   query    string  false "璁㈠崟鍙凤紙妯＄硦鎼滅储锛?
// @Param        status     query    string  false "璁㈠崟鐘舵€侊紙pending/paid/cancelled锛?
// @Param        min_amount query    float64 false "鏈€灏忛噾棰?
// @Param        max_amount query    float64 false "鏈€澶ч噾棰?
// @Param        start_time query    string  false "寮€濮嬫椂闂达紙鏍煎紡锛?006-01-02 15:04:05锛?
// @Param        end_time   query    string  false "缁撴潫鏃堕棿锛堟牸寮忥細2006-01-02 15:04:05锛?
// @Param        order_by   query    string  false "鎺掑簭锛堟牸寮忥細瀛楁:asc/desc锛屽锛歝reated_at:desc锛?
// @Success      200        {object} OrderListResponse
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders [get]
// @Security     BearerAuth
func (r *OrderController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	// 鏋勫缓绛涢€夋潯浠讹紙鍒楄〃鍜屽鍑哄叡鐢級
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// 鏌ヨ璁㈠崟锛堝寘鍚鎯咃級
	ordersWithDetails, total, err := r.orderService.GetOrdersWithDetails(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"filters": filters,
		})
	}

	// 杞崲鍝嶅簲鏁版嵁
	orderList := make([]http.Json, len(ordersWithDetails))
	for i, orderWithDetails := range ordersWithDetails {
		orderJson := r.convertOrderToJson(orderWithDetails.Order)
		// 娣诲姞璁㈠崟璇︽儏
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

// Show 璁㈠崟璇︽儏
// @Summary      鑾峰彇璁㈠崟璇︽儏
// @Description  鏍规嵁璁㈠崟鍙锋垨璁㈠崟ID鑾峰彇璁㈠崟璇︾粏淇℃伅锛岃繑鍥炶鍗曚富琛ㄦ暟鎹拰璁㈠崟璇︽儏琛ㄦ暟鎹紙鏀寔鍒嗚〃鏌ヨ锛夈€備紭鍏堜娇鐢ㄨ鍗曞彿鏌ヨ锛堟洿楂樻晥锛夛紝濡傛灉娌℃湁璁㈠崟鍙峰垯浣跨敤璁㈠崟ID鏌ヨ
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        id         path     string  false "璁㈠崟ID锛堝鏋滄彁渚涗簡璁㈠崟鍙凤紝姝ゅ弬鏁板彲閫夛級"
// @Param        order_no   query    string  false "璁㈠崟鍙凤紙浼樺厛浣跨敤锛屽彲鐩存帴瀹氫綅鍒嗚〃锛?
// @Success      200        {object} OrderDetailResponseWrapper "杩斿洖鏁版嵁鍖呭惈 order锛堣鍗曚富琛級鍜?details锛堣鍗曡鎯呰〃鏁扮粍锛?
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      404        {object} apidoc.Error "璁㈠崟涓嶅瓨鍦?
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders/{id} [get]
// @Security     BearerAuth
func (r *OrderController) Show(ctx http.Context) http.Response {
	// 浼樺厛浠庢煡璇㈠弬鏁拌幏鍙栬鍗曞彿
	orderNo := ctx.Request().Query("order_no", "")

	if orderNo != "" {
		order, details, err := r.orderService.GetOrderByOrderNo(orderNo)
		if err == nil {
			return r.buildOrderDetailResponse(ctx, order, details)
		}
		// 濡傛灉璁㈠崟鍙锋煡璇㈠け璐ワ紝涓旇矾鐢卞弬鏁版槸鏁板瓧ID锛屽皾璇曚娇鐢↖D鏌ヨ
		if routeID := ctx.Request().Route("id"); routeID != "" && orderNo == routeID {
			if orderID := cast.ToUint(routeID); orderID > 0 {
				// 浣跨敤璁㈠崟ID鏌ヨ锛堥渶瑕侀亶鍘嗗垎琛級
				order, details, err := r.orderService.GetOrderByID(orderID, time.Time{})
				if err == nil {
					return r.buildOrderDetailResponse(ctx, order, details)
				}
			}
		}
		return response.Error(ctx, http.StatusNotFound, "order_not_found")
	}

	return response.Error(ctx, http.StatusBadRequest, "order_no_or_id_required")
}

// buildOrderDetailResponse 鏋勫缓璁㈠崟璇︽儏鍝嶅簲锛堟彁鍙栧叕鍏遍€昏緫锛?
func (r *OrderController) buildOrderDetailResponse(ctx http.Context, order *models.Order, details []models.OrderDetail) http.Response {

	// 杞崲璁㈠崟涓昏〃鏁版嵁锛堜娇鐢ㄧ粺涓€鐨勬柟娉曪級
	orderJson := r.convertOrderToJson(*order)

	// 杞崲璁㈠崟璇︽儏鏁版嵁
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

	// 杩斿洖涓昏〃鍜岃鎯呰〃鏁版嵁
	return response.Success(ctx, http.Json{
		"order":   orderJson,
		"details": detailList,
	})
}

// Store 鍒涘缓璁㈠崟
// @Summary      鍒涘缓璁㈠崟
// @Description  鍒涘缓鏂拌鍗曪紝鑷姩闃叉閲嶅鎻愪氦
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        request body     OrderCreateRequest true "鍒涘缓鍙傛暟"
// @Success      200      {object} OrderDetailResponseWrapper
// @Failure      400      {object} apidoc.Error "鍙傛暟閿欒鎴栭噸澶嶆彁浜?
// @Failure      500      {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
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

	// 鍒涘缓璁㈠崟
	order, details, err := r.orderService.CreateOrder(req.UserID, req.Amount, req.Products, req.RequestID, req.Remark)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "create_failed")
	}

	// 杞崲璁㈠崟璇︽儏
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
	// 绉婚櫎涓嶉渶瑕佺殑瀛楁锛堝垱寤鸿鍗曟椂涓嶉渶瑕佽繑鍥?created_at 鍜?updated_at锛?
	delete(orderJson, "created_at")
	delete(orderJson, "updated_at")

	return response.Success(ctx, http.Json{
		"order":   orderJson,
		"details": detailList,
	})
}

// Update 鏇存柊璁㈠崟
// @Summary      鏇存柊璁㈠崟
// @Description  鏇存柊璁㈠崟淇℃伅锛堜富瑕佹槸鐘舵€侊級銆備娇鐢ㄨ鍗曞彿鏌ヨ锛堝彲鐩存帴瀹氫綅鍒嗚〃锛?
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        id         path     string  true "璁㈠崟鍙?
// @Param        request    body     OrderUpdateRequest true  "鏇存柊鍙傛暟"
// @Success      200        {object} apidoc.Success
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders/{id} [put]
// @Security     BearerAuth
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

	if err := r.orderService.UpdateOrderByOrderNo(orderNo, req.Status, req.Remark); err != nil {
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

	if err := r.orderService.DeleteOrderByOrderNo(orderNo); err != nil {
		return response.ErrorWithLog(ctx, "order", err, map[string]any{
			"order_no": orderNo,
		})
	}

	return response.Success(ctx)
}

// Export 瀵煎嚭璁㈠崟鍒楄〃
// @Summary      瀵煎嚭璁㈠崟鍒楄〃
// @Description  鏍规嵁绛涢€夋潯浠跺鍑鸿鍗曞垪琛ㄤ负CSV鏂囦欢锛屾敮鎸佷笌鍒楄〃鏌ヨ鐩稿悓鐨勭瓫閫夋潯浠讹紝鏌ヨ鏃堕棿鑼冨洿涓嶈兘瓒呰繃3涓湀
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        user_id    query    int     false "鐢ㄦ埛ID"
// @Param        order_no   query    string  false "璁㈠崟鍙凤紙妯＄硦鎼滅储锛?
// @Param        status     query    string  false "璁㈠崟鐘舵€侊紙pending/paid/cancelled锛?
// @Param        min_amount query    float64 false "鏈€灏忛噾棰?
// @Param        max_amount query    float64 false "鏈€澶ч噾棰?
// @Param        start_time query    string  false "寮€濮嬫椂闂达紙鏍煎紡锛?006-01-02 15:04:05锛?
// @Param        end_time   query    string  false "缁撴潫鏃堕棿锛堟牸寮忥細2006-01-02 15:04:05锛?
// @Param        order_by   query    string  false "鎺掑簭锛堟牸寮忥細瀛楁:asc/desc锛屽锛歝reated_at:desc锛?
// @Success      200        {object} ExportTaskResponse "瀵煎嚭浠诲姟宸叉彁浜わ紝杩斿洖瀵煎嚭璁板綍ID"
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      401        {object} apidoc.Error "鏈櫥褰?
// @Failure      403        {object} apidoc.Error "鏃犳潈闄?
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders/export [post]
// @Security     BearerAuth
func (r *OrderController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 闃查噸澶嶇偣鍑伙細浣跨敤妗嗘灦鑷甫鐨勫師瀛愰攣锛堥攣浼氬湪10绉掑悗鑷姩杩囨湡锛岄槻姝㈢煭鏃堕棿鍐呴噸澶嶈姹傦級
	lockKey := fmt.Sprintf("export:orders:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	// 灏濊瘯鑾峰彇閿侊紝濡傛灉鑾峰彇澶辫触鍒欒繑鍥為敊璇?
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "already_queued")
	}

	// 鏋勫缓绛涢€夋潯浠?
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	// 鍒涘缓瀵煎嚭璁板綍锛堢姸鎬佷负澶勭悊涓級
	// 鑾峰彇瀛樺偍椹卞姩閰嶇疆
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
		Path:    "", // 澶勭悊瀹屾垚鍚庢洿鏂?
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 灏嗙瓫閫夋潯浠跺簭鍒楀寲涓?JSON
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

	// 鑾峰彇褰撳墠璇█锛堜粠璇锋眰澶存垨鏌ヨ鍙傛暟锛屼笌 middleware 閫昏緫涓€鑷达級
	lang := r.getCurrentLanguage(ctx)
	timezone := helpers.GetCurrentTimezone(ctx)

	// 寮傛鎵ц瀵煎嚭浠诲姟锛堜娇鐢?Job锛?
	// 灏嗗弬鏁板簭鍒楀寲涓?JSON 瀛楃涓蹭紶閫掞紝閬垮厤妗嗘灦瀵瑰鏉傜被鍨嬬殑搴忓垪鍖栭棶棰?
	exportArgsStruct := jobs.ExportOrdersArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "orders",
		Language: lang,
		Timezone: timezone,
	}

	// 搴忓垪鍖栦负 JSON 瀛楃涓?
	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		facades.Log().Errorf("搴忓垪鍖栧鍑哄弬鏁板け璐? export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 璁板綍浠诲姟鎻愪氦鏃ュ織
	facades.Log().Infof("鎻愪氦瀵煎嚭浠诲姟鍒伴槦鍒? export_id=%d, queue_driver=%s, args_json=%s",
		exportRecord.ID, facades.Config().GetString("queue.default"), string(exportArgsJSON))

	// 浣跨敤 queue.Arg 鍖呰 JSON 瀛楃涓插弬鏁?
	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	// 浼犻€?JSON 瀛楃涓蹭綔涓哄弬鏁帮紝浣跨敤 long-running 闃熷垪锛岄伩鍏嶉暱鏃堕棿杩愯鐨勫鍑轰换鍔″奖鍝嶅叾浠栭槦鍒椾换鍔?
	// 鎵€鏈夎€楁椂浠诲姟锛堝鍑恒€佹姤琛ㄧ敓鎴愩€佹壒閲忓鐞嗙瓑锛夐兘搴旇浣跨敤 long-running 闃熷垪
	if err := facades.Queue().Job(&jobs.ExportOrders{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		// 濡傛灉浠诲姟鎻愪氦澶辫触锛岀珛鍗抽噴鏀鹃攣锛岃鐢ㄦ埛鍙互绔嬪嵆閲嶈瘯
		lock.Release()
		facades.Log().Errorf("鎻愪氦瀵煎嚭浠诲姟澶辫触: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("瀵煎嚭浠诲姟宸叉垚鍔熸彁浜ゅ埌闃熷垪: export_id=%d", exportRecord.ID)

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "queued"),
	})
}

// GetExportStatus 鏌ヨ瀵煎嚭鐘舵€?
// @Summary      鏌ヨ瀵煎嚭鐘舵€?
// @Description  鏍规嵁瀵煎嚭璁板綍ID鏌ヨ瀵煎嚭浠诲姟鐨勭姸鎬?
// @Tags         璁㈠崟绠＄悊
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "瀵煎嚭璁板綍ID"
// @Success      200  {object}  ExportStatusResponse
// @Failure      400  {object}  apidoc.Error  "鍙傛暟閿欒"
// @Failure      401  {object}  apidoc.Error  "鏈櫥褰?
// @Failure      403  {object}  apidoc.Error  "鏃犳潈闄?
// @Failure      500  {object}  apidoc.Error  "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders/export/status/{id} [get]
// @Security     BearerAuth
func (r *OrderController) GetExportStatus(ctx http.Context) http.Response {
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	exportRecordService := services.NewExportRecordService()
	exportRecord, err := exportRecordService.GetByID(exportID)
	if err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 妫€鏌ユ潈闄愶細鍙兘鏌ョ湅鑷繁鐨勫鍑鸿褰?
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}
	if exportRecord.AdminID != adminID {
		return response.Error(ctx, http.StatusForbidden, "forbidden")
	}

	// 鐢熸垚鏂囦欢URL
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

// getCurrentLanguage 鑾峰彇褰撳墠璇锋眰鐨勮瑷€锛堜娇鐢ㄩ€氱敤宸ュ叿鍑芥暟锛?
func (r *OrderController) getCurrentLanguage(ctx http.Context) string {
	return utils.GetCurrentLanguage(ctx)
}

// Import 瀵煎叆璁㈠崟
// @Summary      瀵煎叆璁㈠崟
// @Description  浠嶤SV鏂囦欢瀵煎叆璁㈠崟鏁版嵁锛屾敮鎸佹壒閲忓鍏?
// @Tags         璁㈠崟绠＄悊
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "CSV鏂囦欢"
// @Success      200  {object} ImportResultResponse "瀵煎叆鎴愬姛锛岃繑鍥炲鍏ョ粨鏋?
// @Failure      400  {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      401  {object} apidoc.Error "鏈櫥褰?
// @Failure      403  {object} apidoc.Error "鏃犳潈闄?
// @Failure      500  {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/orders/import [post]
// @Security     BearerAuth
func (r *OrderController) Import(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 鑾峰彇涓婁紶鐨勬枃浠?
	file, err := ctx.Request().File("file")
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "file_required")
	}

	// 楠岃瘉鏂囦欢绫诲瀷锛堝彧鍏佽CSV锛?
	filename := file.GetClientOriginalName()
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidFileType.Code)
	}

	// 璇诲彇鏂囦欢鍐呭
	storage := facades.Storage().Disk("local")
	savedPath, err := storage.PutFile("", file)
	if err != nil {
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
		})
	}

	// 璇诲彇鏂囦欢鍐呭
	csvContent, err := storage.Get(savedPath)
	if err != nil {
		_ = storage.Delete(savedPath)
		return response.ErrorWithLog(ctx, "import", err, map[string]any{
			"filename": filename,
		})
	}

	// 娓呯悊涓存椂鏂囦欢
	defer func() {
		_ = storage.Delete(savedPath)
	}()

	// 瀵煎叆璁㈠崟
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
