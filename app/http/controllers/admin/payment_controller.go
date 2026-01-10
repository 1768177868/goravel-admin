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

	// 与列表保持一致：未传 start_time 时默认最近 7 天；未传 end_time 时默认当前时间
	// 这样导出数据集与列表查询数据集一致，并避免扫到未建表的历史月份
	if startTime.IsZero() {
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	}
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}

	// 校验时间范围不超过 3 个月（与列表/导出一致）
	if valid, err := utils.ValidateTimeRange(startTime, endTime); !valid {
		// ValidateTimeRange 返回的是可翻译错误键，这里直接返回 key 交给前端处理
		return services.PaymentFilters{}, response.Error(ctx, http.StatusBadRequest, err.Error())
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
// @Description  根据支付单号获取支付记录详细信息（分表后ID可能重复，使用支付单号查询）
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        payment_no path     string  true  "支付单号"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      404        {object} map[string]any "支付记录不存在"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payments/{payment_no} [get]
// @Security     BearerAuth
func (r *PaymentController) Show(ctx http.Context) http.Response {
	paymentNo := ctx.Request().Route("id") // 路由参数名保持兼容
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

// Export 导出支付记录
// @Summary      导出支付记录
// @Description  异步导出支付记录为CSV文件
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        payment_no        query    string  false "支付单号"
// @Param        order_no          query    string  false "订单号"
// @Param        payment_method_id query    int     false "支付方式ID"
// @Param        user_id           query    int     false "用户ID"
// @Param        status            query    string  false "支付状态"
// @Param        start_time        query    string  false "开始时间"
// @Param        end_time          query    string  false "结束时间"
// @Success      200        {object} map[string]any
// @Failure      400        {object} map[string]any "参数错误"
// @Failure      429        {object} map[string]any "导出任务正在进行中"
// @Failure      500        {object} map[string]any "服务器错误"
// @Router       /api/admin/payments/export [post]
// @Security     BearerAuth
func (r *PaymentController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 防重复点击
	lockKey := fmt.Sprintf("export:payments:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "export_in_progress")
	}

	// 构建筛选条件
	filters, resp := r.buildFilters(ctx)
	if resp != nil {
		return resp
	}

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
		Type:    models.ExportTypePayments,
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "",
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 序列化筛选条件
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

	exportArgsStruct := jobs.ExportPaymentsArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "payments",
		Language: lang,
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		facades.Log().Errorf("序列化导出参数失败: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("提交支付记录导出任务到队列: export_id=%d", exportRecord.ID)

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportPayments{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		facades.Log().Errorf("提交导出任务失败: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("支付记录导出任务已成功提交到队列: export_id=%d", exportRecord.ID)

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "export_task_submitted"),
	})
}

// GetExportStatus 查询导出状态
// @Summary      查询支付记录导出状态
// @Description  根据导出记录ID查询导出任务的状态
// @Tags         支付管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "导出记录ID"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any  "参数错误"
// @Failure      404  {object}  map[string]any  "导出记录不存在"
// @Failure      500  {object}  map[string]any  "服务器错误"
// @Router       /api/admin/payments/export/status/{id} [get]
// @Security     BearerAuth
func (r *PaymentController) GetExportStatus(ctx http.Context) http.Response {
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "export_id_required")
	}

	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportID).First(&exportRecord); err != nil {
		return response.Error(ctx, http.StatusNotFound, "export_not_found")
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

// getCurrentLanguage 获取当前语言（使用通用工具函数）
func (r *PaymentController) getCurrentLanguage(ctx http.Context) string {
	return utils.GetCurrentLanguage(ctx)
}

// getExportStatusText 获取导出状态文本
func (r *PaymentController) getExportStatusText(ctx http.Context, status uint8) string {
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
