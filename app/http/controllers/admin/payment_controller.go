package admin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
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

type PaymentController struct {
	paymentService services.PaymentService
}

type PaymentMethodSimple struct {
	ID   uint   `json:"id" example:"1"`        // 鏀粯鏂瑰紡ID
	Name string `json:"name" example:"寰俊鏀粯"` // 鏀粯鏂瑰紡鍚嶇О
	Code string `json:"code" example:"wechat"` // 鏀粯鏂瑰紡浠ｇ爜
	Type string `json:"type" example:"wechat"` // 鏀粯绫诲瀷
}

type PaymentResponse struct {
	ID              uint                `json:"id" example:"1"`                           // 鏀粯璁板綍ID
	PaymentNo       string              `json:"payment_no" example:"PAY202604090001"`     // 鏀粯鍗曞彿
	OrderNo         string              `json:"order_no" example:"ORD202604090001"`       // 璁㈠崟鍙?
	PaymentMethodID uint                `json:"payment_method_id" example:"1"`            // 鏀粯鏂瑰紡ID
	UserID          uint                `json:"user_id" example:"1001"`                   // 鐢ㄦ埛ID
	Amount          float64             `json:"amount" example:"99.99"`                   // 鏀粯閲戦
	Status          string              `json:"status" example:"paid"`                    // 鏀粯鐘舵€?
	ThirdPartyNo    string              `json:"third_party_no" example:"WX202604090001"`  // 绗笁鏂瑰崟鍙?
	PayTime         string              `json:"pay_time" example:"2024-01-01 00:00:00"`   // 鏀粯鏃堕棿
	FailReason      string              `json:"fail_reason" example:""`                   // 澶辫触鍘熷洜
	Remark          string              `json:"remark" example:"澶囨敞淇℃伅"`                  // 澶囨敞
	CreatedAt       string              `json:"created_at" example:"2024-01-01 00:00:00"` // 鍒涘缓鏃堕棿
	UpdatedAt       string              `json:"updated_at" example:"2024-01-01 00:00:00"` // 鏇存柊鏃堕棿
	PaymentMethod   PaymentMethodSimple `json:"payment_method,omitempty"`                 // 鏀粯鏂瑰紡淇℃伅
}

type PaymentListData struct {
	Data []PaymentResponse `json:"data"` // 鏀粯璁板綍鍒楄〃
	apidoc.Pagination
}

type PaymentListResponse struct {
	apidoc.Success
	Data PaymentListData `json:"data"`
}

type PaymentDetailResponse struct {
	apidoc.Success
	Data PaymentResponse `json:"data"`
}

func NewPaymentController() *PaymentController {
	return &PaymentController{
		paymentService: services.NewPaymentService(),
	}
}

// buildFilters 鏋勫缓绛涢€夋潯浠讹紙鍒楄〃鍜屽鍑哄叡鐢級
func (r *PaymentController) buildFilters(ctx http.Context) (services.PaymentFilters, http.Response) {
	paymentNo := ctx.Request().Input("payment_no", ctx.Request().Query("payment_no", ""))
	orderNo := ctx.Request().Input("order_no", ctx.Request().Query("order_no", ""))
	paymentMethodID := cast.ToUint(ctx.Request().Input("payment_method_id", ctx.Request().Query("payment_method_id", "0")))
	userID := cast.ToUint(ctx.Request().Input("user_id", ctx.Request().Query("user_id", "0")))
	status := ctx.Request().Input("status", ctx.Request().Query("status", ""))
	orderBy := ctx.Request().Input("order_by", ctx.Request().Query("order_by", ""))

	var startTime, endTime time.Time
	if parsedStartTime, resp := parseOptionalTimeFromInputOrQuery(ctx, "start_time", "invalid_start_time"); resp != nil {
		return services.PaymentFilters{}, resp
	} else {
		startTime = parsedStartTime
	}
	if parsedEndTime, resp := parseOptionalTimeFromInputOrQuery(ctx, "end_time", "invalid_end_time"); resp != nil {
		return services.PaymentFilters{}, resp
	} else {
		endTime = parsedEndTime
	}

	// 涓庡垪琛ㄤ繚鎸佷竴鑷达細鏈紶 start_time 鏃堕粯璁ゆ渶杩?7 澶╋紱鏈紶 end_time 鏃堕粯璁ゅ綋鍓嶆椂闂?
	// 杩欐牱瀵煎嚭鏁版嵁闆嗕笌鍒楄〃鏌ヨ鏁版嵁闆嗕竴鑷达紝骞堕伩鍏嶆壂鍒版湭寤鸿〃鐨勫巻鍙叉湀浠?
	if startTime.IsZero() {
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	}
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}

	// 鏍￠獙鏃堕棿鑼冨洿涓嶈秴杩?3 涓湀锛堜笌鍒楄〃/瀵煎嚭涓€鑷达級
	if resp := validateTimeRangeResponse(ctx, startTime, endTime); resp != nil {
		return services.PaymentFilters{}, resp
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

// Index 鏀粯璁板綍鍒楄〃
// @Summary      鑾峰彇鏀粯璁板綍鍒楄〃
// @Description  鍒嗛〉鑾峰彇鏀粯璁板綍鍒楄〃锛屾敮鎸佸鏉′欢绛涢€?
// @Tags         鏀粯绠＄悊
// @Accept       json
// @Produce      json
// @Param        page             query    int     false "椤电爜" default(1)
// @Param        page_size        query    int     false "姣忛〉鏁伴噺" default(10)
// @Param        payment_no        query    string  false "鏀粯鍗曞彿锛堟ā绯婃悳绱級"
// @Param        order_no          query    string  false "璁㈠崟鍙凤紙妯＄硦鎼滅储锛?
// @Param        payment_method_id query    int     false "鏀粯鏂瑰紡ID"
// @Param        user_id          query    int     false "鐢ㄦ埛ID"
// @Param        status           query    string  false "鏀粯鐘舵€侊紙pending/paid/failed/cancelled锛?
// @Param        start_time       query    string  false "寮€濮嬫椂闂达紙鏍煎紡锛?006-01-02 15:04:05锛?
// @Param        end_time         query    string  false "缁撴潫鏃堕棿锛堟牸寮忥細2006-01-02 15:04:05锛?
// @Param        order_by         query    string  false "鎺掑簭锛堟牸寮忥細瀛楁:asc/desc锛屽锛歝reated_at:desc锛?
// @Success      200        {object} PaymentListResponse
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
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

	// 杞崲鍝嶅簲鏁版嵁
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

		// 娣诲姞鏀粯鏂瑰紡淇℃伅
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

// Show 鏀粯璁板綍璇︽儏
// @Summary      鑾峰彇鏀粯璁板綍璇︽儏
// @Description  鏍规嵁鏀粯鍗曞彿鑾峰彇鏀粯璁板綍璇︾粏淇℃伅锛堝垎琛ㄥ悗ID鍙兘閲嶅锛屼娇鐢ㄦ敮浠樺崟鍙锋煡璇級
// @Tags         鏀粯绠＄悊
// @Accept       json
// @Produce      json
// @Param        id         path     string  true  "鏀粯鍗曞彿"
// @Success      200        {object} PaymentDetailResponse
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      404        {object} apidoc.Error "鏀粯璁板綍涓嶅瓨鍦?
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/payments/{id} [get]
// @Security     BearerAuth
func (r *PaymentController) Show(ctx http.Context) http.Response {
	paymentNo := ctx.Request().Route("id") // 璺敱鍙傛暟鍚嶄繚鎸佸吋瀹?
	if paymentNo == "" {
		return response.Error(ctx, http.StatusBadRequest, "payment_no_required")
	}
	payment, err := r.paymentService.GetPaymentByPaymentNo(paymentNo)
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

	// 娣诲姞鏀粯鏂瑰紡淇℃伅
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

// formatPayTime 鏍煎紡鍖栨敮浠樻椂闂翠负瀛楃涓?
func (r *PaymentController) formatPayTime(t *time.Time) string {
	return utils.FormatDateTimePtr(t)
}

// Export 瀵煎嚭鏀粯璁板綍
// @Summary      瀵煎嚭鏀粯璁板綍
// @Description  寮傛瀵煎嚭鏀粯璁板綍涓篊SV鏂囦欢
// @Tags         鏀粯绠＄悊
// @Accept       json
// @Produce      json
// @Param        payment_no        query    string  false "鏀粯鍗曞彿"
// @Param        order_no          query    string  false "璁㈠崟鍙?
// @Param        payment_method_id query    int     false "鏀粯鏂瑰紡ID"
// @Param        user_id           query    int     false "鐢ㄦ埛ID"
// @Param        status            query    string  false "鏀粯鐘舵€?
// @Param        start_time        query    string  false "寮€濮嬫椂闂?
// @Param        end_time          query    string  false "缁撴潫鏃堕棿"
// @Success      200        {object} ExportTaskResponse
// @Failure      400        {object} apidoc.Error "鍙傛暟閿欒"
// @Failure      429        {object} apidoc.Error "瀵煎嚭浠诲姟姝ｅ湪杩涜涓?
// @Failure      500        {object} apidoc.Error "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/payments/export [post]
// @Security     BearerAuth
func (r *PaymentController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 闃查噸澶嶇偣鍑?
	lockKey := fmt.Sprintf("export:payments:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "already_queued")
	}

	// 鏋勫缓绛涢€夋潯浠?
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

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
		Type:    models.ExportTypePayments,
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "",
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 搴忓垪鍖栫瓫閫夋潯浠?
	filtersMap := map[string]any{
		"payment_no":        filters.PaymentNo,
		"order_no":          filters.OrderNo,
		"payment_method_id": filters.PaymentMethodID,
		"user_id":           filters.UserID,
		"status":            filters.Status,
		"order_by":          filters.OrderBy,
	}
	if !filters.StartTime.IsZero() {
		filtersMap["start_time"] = utils.FormatDateTime(filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		filtersMap["end_time"] = utils.FormatDateTime(filters.EndTime)
	}

	lang := r.getCurrentLanguage(ctx)
	timezone := helpers.GetCurrentTimezone(ctx)

	exportArgsStruct := jobs.ExportPaymentsArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "payments",
		Language: lang,
		Timezone: timezone,
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		facades.Log().Errorf("搴忓垪鍖栧鍑哄弬鏁板け璐? export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("鎻愪氦鏀粯璁板綍瀵煎嚭浠诲姟鍒伴槦鍒? export_id=%d", exportRecord.ID)

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportPayments{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		facades.Log().Errorf("鎻愪氦瀵煎嚭浠诲姟澶辫触: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("鏀粯璁板綍瀵煎嚭浠诲姟宸叉垚鍔熸彁浜ゅ埌闃熷垪: export_id=%d", exportRecord.ID)

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "queued"),
	})
}

// GetExportStatus 鏌ヨ瀵煎嚭鐘舵€?
// @Summary      鏌ヨ鏀粯璁板綍瀵煎嚭鐘舵€?
// @Description  鏍规嵁瀵煎嚭璁板綍ID鏌ヨ瀵煎嚭浠诲姟鐨勭姸鎬?
// @Tags         鏀粯绠＄悊
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "瀵煎嚭璁板綍ID"
// @Success      200  {object}  ExportStatusResponse
// @Failure      400  {object}  apidoc.Error  "鍙傛暟閿欒"
// @Failure      404  {object}  apidoc.Error  "瀵煎嚭璁板綍涓嶅瓨鍦?
// @Failure      500  {object}  apidoc.Error  "鏈嶅姟鍣ㄩ敊璇?
// @Router       /api/admin/payments/export/status/{id} [get]
// @Security     BearerAuth
func (r *PaymentController) GetExportStatus(ctx http.Context) http.Response {
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportID).FirstOrFail(&exportRecord); err != nil {
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	result := http.Json{
		"id":          exportRecord.ID,
		"status":      exportRecord.Status,
		"status_text": r.getExportStatusText(ctx, exportRecord.Status),
		"path":        exportRecord.Path,
		"filename":    exportRecord.Filename,
		"size":        exportRecord.Size,
		"error_msg":   exportRecord.ErrorMsg,
		"created_at":  exportRecord.CreatedAt,
		"updated_at":  exportRecord.UpdatedAt,
	}

	if exportRecord.Status == models.ExportStatusSuccess && exportRecord.Path != "" {
		result["download_url"] = fmt.Sprintf("/api/admin/exports/%d/download", exportRecord.ID)
	}

	return response.Success(ctx, result)
}

// getCurrentLanguage 鑾峰彇褰撳墠璇█锛堜娇鐢ㄩ€氱敤宸ュ叿鍑芥暟锛?
func (r *PaymentController) getCurrentLanguage(ctx http.Context) string {
	return utils.GetCurrentLanguage(ctx)
}

// getExportStatusText 鑾峰彇瀵煎嚭鐘舵€佹枃鏈?
func (r *PaymentController) getExportStatusText(ctx http.Context, status uint8) string {
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
