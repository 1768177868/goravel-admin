package admin

import (
	"encoding/json"
	"fmt"
	appfacades "goravel/app/facades"
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

type PaymentController struct {}


type PaymentMethodSimple struct {
	ID   uint   `json:"id" example:"1"`
	Name string `json:"name" example:"name"`
	Code string `json:"code" example:"wechat"`
	Type string `json:"type" example:"wechat"`
}

type PaymentResponse struct {
	ID              uint                `json:"id" example:"1"`
	PaymentNo       string              `json:"payment_no" example:"PAY202604090001"`
	OrderNo         string              `json:"order_no" example:"ORD202604090001"`
	PaymentMethodID uint                `json:"payment_method_id" example:"1"`
	UserID          uint                `json:"user_id" example:"1001"`
	Amount          float64             `json:"amount" example:"99.99"`
	Status          string              `json:"status" example:"paid"`
	ThirdPartyNo    string              `json:"third_party_no" example:"WX202604090001"`
	PayTime         string              `json:"pay_time" example:"2024-01-01 00:00:00"`
	FailReason      string              `json:"fail_reason" example:""`
	Remark          string              `json:"remark" example:"remark"`
	CreatedAt       string              `json:"created_at" example:"2024-01-01 00:00:00"`
	UpdatedAt       string              `json:"updated_at" example:"2024-01-01 00:00:00"`
	PaymentMethod   PaymentMethodSimple `json:"payment_method,omitempty"`
}

type PaymentListData struct {
	Data []PaymentResponse `json:"data"`
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
	return &PaymentController{}
}

func (r *PaymentController) paymentService(ctx http.Context) services.PaymentService {
	return services.NewPaymentService(ctx)
}


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

	if startTime.IsZero() {
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	}
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}

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

func (r *PaymentController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

	payments, total, err := r.paymentService(ctx).GetPayments(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "payment", err, map[string]any{
			"filters": filters,
		})
	}

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

func (r *PaymentController) Show(ctx http.Context) http.Response {
	paymentNo := ctx.Request().Route("id")
	if paymentNo == "" {
		return response.Error(ctx, http.StatusBadRequest, "payment_no_required")
	}
	payment, err := r.paymentService(ctx).GetPaymentByPaymentNo(paymentNo)
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

func (r *PaymentController) formatPayTime(t *time.Time) string {
	return utils.FormatDateTimePtr(t)
}

func (r *PaymentController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	lockKey := fmt.Sprintf("export:payments:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "already_queued")
	}

	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

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
	if err := appfacades.OrmQuery(ctx).Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

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
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportPayments{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "queued"),
	})
}

func (r *PaymentController) GetExportStatus(ctx http.Context) http.Response {
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var exportRecord models.Export
	if err := appfacades.OrmQuery(ctx).Where("id", exportID).FirstOrFail(&exportRecord); err != nil {
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

func (r *PaymentController) getCurrentLanguage(ctx http.Context) string {
	return utils.GetCurrentLanguage(ctx)
}

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
