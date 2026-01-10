package jobs

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

// ExportPaymentsArgs 导出支付记录任务的参数
type ExportPaymentsArgs struct {
	ExportID uint           `json:"export_id"` // 导出记录ID
	AdminID  uint           `json:"admin_id"`  // 管理员ID
	Filters  map[string]any `json:"filters"`   // 筛选条件（JSON序列化）
	Type     string         `json:"type"`      // 导出类型，如 "payments"
	Language string         `json:"language"`  // 语言代码，如 "cn" 或 "en"
}

// ExportPayments 导出支付记录异步任务
type ExportPayments struct {
}

// errPaymentExportRecordMissing 导出记录被删除时的哨兵错误
var errPaymentExportRecordMissing = errors.New("payment export record missing (deleted)")

func (r *ExportPayments) markExportFailed(exportID uint, errorMsg string) {
	if exportID == 0 {
		return
	}
	var failedRecord models.Export
	if queryErr := facades.Orm().Query().Where("id", exportID).First(&failedRecord); queryErr == nil {
		failedRecord.Status = models.ExportStatusFailed
		failedRecord.ErrorMsg = errorMsg
		if saveErr := facades.Orm().Query().Save(&failedRecord); saveErr != nil {
			facades.Log().Errorf("更新导出记录失败状态失败: export_id=%d, error=%v", exportID, saveErr)
		}
	}
}

func (r *ExportPayments) isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// MySQL: Error 1146 (42S02): Table 'xxx' doesn't exist
	// 不同驱动/封装可能会改变格式，这里做宽松匹配，避免误判导致任务失败
	if strings.Contains(msg, "1146") || strings.Contains(msg, "42s02") {
		return true
	}
	return strings.Contains(msg, "table") && (strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "does not exist"))
}

func (r *ExportPayments) Signature() string {
	return "export_payments"
}

func (r *ExportPayments) Handle(args ...any) (retErr error) {
	var exportID uint
	// 防御性：任何 panic 都不应该把 worker 进程打崩；同时需要把导出记录标记为失败，避免一直“处理中”
	defer func() {
		if rec := recover(); rec != nil {
			errorMsg := fmt.Sprintf("panic: %v", rec)
			facades.Log().Errorf("ExportPayments Job panic: export_id=%v, panic=%v", func() any {
				if len(args) > 0 {
					return args[0]
				}
				return nil
			}(), rec)
			r.markExportFailed(exportID, errorMsg)
			retErr = fmt.Errorf("%s", errorMsg)
		}
	}()

	if len(args) < 1 {
		facades.Log().Errorf("ExportPayments Job 参数不足: args=%v", args)
		return apperrors.ErrInvalidArgument.WithMessage("missing export arguments")
	}

	// 解析参数
	var exportArgs ExportPaymentsArgs
	switch v := args[0].(type) {
	case ExportPaymentsArgs:
		exportArgs = v
	case string:
		// JSON 字符串，需要反序列化
		if err := json.Unmarshal([]byte(v), &exportArgs); err != nil {
			facades.Log().Errorf("反序列化参数失败: %v, JSON: %s", err, v)
			return apperrors.ErrInvalidArgument.WithMessage(fmt.Sprintf("failed to unmarshal export arguments: %v", err))
		}
	case map[string]any:
		if exportID, ok := utils.GetUint(v, "export_id"); ok {
			exportArgs.ExportID = exportID
		}
		if adminID, ok := utils.GetUint(v, "admin_id"); ok {
			exportArgs.AdminID = adminID
		}
		if filters, ok := utils.GetMap(v, "filters"); ok {
			exportArgs.Filters = filters
		}
		if exportType, ok := utils.GetString(v, "type"); ok {
			exportArgs.Type = exportType
		}
		if lang, ok := utils.GetString(v, "language"); ok {
			exportArgs.Language = lang
		}
	default:
		facades.Log().Errorf("不支持的参数类型: %T, 值: %+v", args[0], args[0])
		return apperrors.ErrInvalidArgument.WithMessage(fmt.Sprintf("invalid export arguments type: %T", args[0]))
	}

	if exportArgs.ExportID == 0 {
		facades.Log().Errorf("export_id 为 0，参数解析失败: %+v", exportArgs)
		return apperrors.ErrInvalidArgument.WithMessage("export_id is required")
	}
	exportID = exportArgs.ExportID

	// 检查导出记录是否存在
	exists, err := facades.Orm().Query().Model(&models.Export{}).Where("id", exportArgs.ExportID).Exists()
	if err != nil {
		errorlog.Record(context.TODO(), "export", "检查导出记录是否存在失败", map[string]any{
			"export_id": exportArgs.ExportID,
			"error":     err.Error(),
		}, "检查导出记录是否存在失败: %v", err)
		return err
	}
	if !exists {
		facades.Log().Infof("导出记录不存在，停止任务: export_id=%d", exportArgs.ExportID)
		return nil
	}

	var exportRecords []models.Export
	if err := facades.Orm().Query().Where("id", exportArgs.ExportID).Limit(1).Get(&exportRecords); err != nil {
		errorlog.Record(context.TODO(), "export", "查询导出记录失败", map[string]any{
			"export_id": exportArgs.ExportID,
			"error":     err.Error(),
		}, "查询导出记录失败: %v", err)
		return err
	}
	if len(exportRecords) == 0 {
		facades.Log().Infof("导出记录已被删除，停止任务: export_id=%d", exportArgs.ExportID)
		return nil
	}
	exportRecord := exportRecords[0]

	exportRecord.Status = models.ExportStatusProcessing
	exportRecord.ErrorMsg = ""
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		facades.Log().Errorf("更新导出状态为处理中失败: export_id=%d, error=%v", exportArgs.ExportID, err)
		return err
	}

	// 执行导出
	jobErr := r.exportPayments(exportArgs)

	if errors.Is(jobErr, errPaymentExportRecordMissing) {
		facades.Log().Infof("导出任务检测到导出记录已删除，停止任务: export_id=%d", exportArgs.ExportID)
		return nil
	}

	if jobErr != nil {
		errorMsg := jobErr.Error()
		if errorMsg == "" {
			errorMsg = "未知错误"
		}

		facades.Log().Errorf("导出任务失败: export_id=%d, error=%s", exportArgs.ExportID, errorMsg)
		errorlog.Record(context.TODO(), "export", "导出失败", map[string]any{
			"export_id": exportArgs.ExportID,
			"error":     errorMsg,
		}, "导出失败: %v", jobErr)

		r.markExportFailed(exportArgs.ExportID, errorMsg)

		return jobErr
	}

	return nil
}

// exportPayments 执行支付记录导出
func (r *ExportPayments) exportPayments(args ExportPaymentsArgs) error {
	// 构建筛选条件
	filters := services.PaymentFilters{}

	if paymentNo, ok := utils.GetString(args.Filters, "payment_no"); ok {
		filters.PaymentNo = paymentNo
	}
	if orderNo, ok := utils.GetString(args.Filters, "order_no"); ok {
		filters.OrderNo = orderNo
	}
	if paymentMethodID, ok := utils.GetUint(args.Filters, "payment_method_id"); ok {
		filters.PaymentMethodID = paymentMethodID
	}
	if userID, ok := utils.GetUint(args.Filters, "user_id"); ok {
		filters.UserID = userID
	}
	if status, ok := utils.GetString(args.Filters, "status"); ok {
		filters.Status = status
	}
	if orderBy, ok := utils.GetString(args.Filters, "order_by"); ok {
		filters.OrderBy = orderBy
	}

	// 解析时间
	if startTimeStr, ok := utils.GetString(args.Filters, "start_time"); ok && startTimeStr != "" {
		if t, err := utils.ParseDateTimeUTC(startTimeStr); err == nil {
			filters.StartTime = t
		}
	}
	if endTimeStr, ok := utils.GetString(args.Filters, "end_time"); ok && endTimeStr != "" {
		if t, err := utils.ParseDateTimeUTC(endTimeStr); err == nil {
			filters.EndTime = t
		}
	}

	// 时间范围兜底：与 Controller.buildFilters 保持一致
	// 未传 start_time 时默认最近 7 天；未传 end_time 时默认当前时间
	if filters.StartTime.IsZero() {
		filters.StartTime = time.Now().UTC().AddDate(0, 0, -7)
	}
	if filters.EndTime.IsZero() {
		filters.EndTime = time.Now().UTC()
	}

	// 准备表头
	headerKeys := []string{
		"export_header_id",
		"export_header_payment_no",
		"export_header_order_no",
		"export_header_payment_method",
		"export_header_user_id",
		"export_header_amount",
		"export_header_status",
		"export_header_third_party_no",
		"export_header_pay_time",
		"export_header_fail_reason",
		"export_header_remark",
		"export_header_created_at",
	}

	lang := args.Language
	if lang == "" {
		lang = facades.Config().GetString("app.locale", "cn")
	}
	lang = utils.NormalizeLanguage(lang)

	headers := utils.TranslateHeaders(headerKeys, lang)

	exportService := services.NewExportService(nil)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("payments_%d_%s.csv", args.ExportID, timestamp)
	filePath := path.Join("exports", filename)

	// 预写入文件信息
	{
		var exportRecord models.Export
		if err := facades.Orm().Query().Where("id", args.ExportID).First(&exportRecord); err == nil {
			changed := false
			if exportRecord.Path == "" {
				exportRecord.Path = filePath
				changed = true
			}
			if exportRecord.Filename == "" {
				exportRecord.Filename = filename
				changed = true
			}
			if exportRecord.Extension == "" {
				exportRecord.Extension = "csv"
				changed = true
			}
			if changed {
				_ = facades.Orm().Query().Save(&exportRecord)
			}
		}
	}

	lastUpdateAt := time.Now().Add(-10 * time.Second)
	lastExistCheckAt := time.Now().Add(-10 * time.Second)
	shouldStop := func() bool {
		if time.Since(lastExistCheckAt) < 2*time.Second {
			return false
		}
		lastExistCheckAt = time.Now()

		exists, err := facades.Orm().Query().Model(&models.Export{}).Where("id", args.ExportID).Exists()
		if err != nil {
			return false
		}
		return !exists
	}

	filePath, err := exportService.ExportToCSVStreamAtWithProgress(headers, filePath, func(w *csv.Writer) error {
		return r.writePaymentsToCSV(w, filters, lang, shouldStop)
	}, func(writtenBytes int64) {
		if shouldStop() {
			return
		}

		if time.Since(lastUpdateAt) < 3*time.Second {
			return
		}
		lastUpdateAt = time.Now()
		result, _ := facades.Orm().Query().Model(&models.Export{}).Where("id", args.ExportID).Update(map[string]any{
			"size": writtenBytes,
		})
		if result == nil || result.RowsAffected == 0 {
			lastExistCheckAt = time.Now().Add(-10 * time.Second)
		}
	}, true)
	if err != nil {
		if shouldStop() {
			return errPaymentExportRecordMissing
		}
		errorlog.Record(context.TODO(), "export", "导出文件失败", map[string]any{
			"export_id": args.ExportID,
			"filename":  filename,
			"error":     err.Error(),
		}, "导出文件失败: %v", err)
		return fmt.Errorf("导出文件失败: %v", err)
	}

	// 更新导出记录
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", args.ExportID).First(&exportRecord); err != nil {
		return nil
	}

	exportRecord.Path = filePath
	exportRecord.Filename = path.Base(filePath)

	if ext := path.Ext(filePath); ext != "" {
		exportRecord.Extension = ext[1:]
	} else {
		exportRecord.Extension = "csv"
	}

	// 获取文件大小
	if exportRecord.Disk == "" {
		// 用户"导出中删除"会删记录；如果这里还能查到记录但 disk 为空，直接跳过 Size，避免 Disk("") panic
		facades.Log().Warningf("导出记录 disk 为空，跳过获取文件大小: export_id=%d, path=%s", args.ExportID, filePath)
	} else {
		storage := facades.Storage().Disk(exportRecord.Disk)
		if fileInfo, err := storage.Size(filePath); err == nil {
			exportRecord.Size = fileInfo
		} else {
			facades.Log().Warningf("获取文件大小失败: export_id=%d, error=%v", args.ExportID, err)
			// 保留导出过程中的 size（实时更新值），避免这里覆盖为 0
		}
	}

	exportRecord.Status = models.ExportStatusSuccess
	exportRecord.ErrorMsg = ""

	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		facades.Log().Errorf("保存导出记录失败: export_id=%d, error=%v", args.ExportID, err)
		return fmt.Errorf("更新导出记录失败: %v", err)
	}

	return nil
}

// writePaymentsToCSV 按分表分批查询支付记录并写入 CSV
func (r *ExportPayments) writePaymentsToCSV(w *csv.Writer, filters services.PaymentFilters, lang string, shouldStop func() bool) error {
	orderBy := strings.TrimSpace(filters.OrderBy)
	if orderBy == "" {
		orderBy = "created_at:desc"
	}
	field, direction := r.parseOrderBy(orderBy)
	if field != "created_at" {
		field = "created_at"
	}

	// 获取分表列表
	// 如果精确指定了 payment_no，则可以从 payment_no 中解析日期直接定位分表，避免按月扫描
	var tableNames []string
	if filters.PaymentNo != "" && len(filters.PaymentNo) >= 11 {
		dateStr := filters.PaymentNo[3:11] // PAY + YYYYMMDD
		if t, err := time.Parse("20060102", dateStr); err == nil {
			tableNames = []string{utils.GetShardingTableName("payments", t)}
		}
	}
	if len(tableNames) == 0 {
		tableNames = utils.GetShardingTableNames("payments", filters.StartTime, filters.EndTime)
	}
	if len(tableNames) == 0 {
		return nil
	}

	// desc 时按月份倒序导出
	if direction == "desc" {
		for i, j := 0, len(tableNames)-1; i < j; i, j = i+1, j-1 {
			tableNames[i], tableNames[j] = tableNames[j], tableNames[i]
		}
	}

	// 预加载支付方式
	paymentMethodMap := r.loadPaymentMethods()

	const chunkSize = 2000

	for _, tableName := range tableNames {
		lastTimeStr := ""
		var lastID uint = 0

		for {
			if shouldStop != nil && shouldStop() {
				return errPaymentExportRecordMissing
			}

			query := services.BuildPaymentQuery(tableName, filters)

			// Keyset 分页
			if lastTimeStr != "" {
				if direction == "desc" {
					query = query.Where(fmt.Sprintf("(%s < ? OR (%s = ? AND id < ?))", field, field), lastTimeStr, lastTimeStr, lastID)
				} else {
					query = query.Where(fmt.Sprintf("(%s > ? OR (%s = ? AND id > ?))", field, field), lastTimeStr, lastTimeStr, lastID)
				}
			} else if lastID > 0 {
				// created_at 为空/不可用时，退化为仅使用 id 做 keyset
				if direction == "desc" {
					query = query.Where("id < ?", lastID)
				} else {
					query = query.Where("id > ?", lastID)
				}
			}

			if direction == "desc" {
				query = query.OrderByDesc(field).OrderByDesc("id")
			} else {
				query = query.OrderBy(field).OrderBy("id")
			}

			var payments []models.Payment
			if err := query.Limit(chunkSize).Get(&payments); err != nil {
				// 分表不存在时（如历史月份未建表），跳过该表，避免任务失败/卡死
				if r.isTableNotExistsError(err) {
					facades.Log().Warningf("支付记录分表不存在，跳过: table=%s, error=%v", tableName, err)
					break
				}
				return fmt.Errorf("查询支付记录失败: %v", err)
			}

			if len(payments) == 0 {
				break
			}

			// 写入 CSV
			for _, payment := range payments {
				paymentMethodName := ""
				if pm, ok := paymentMethodMap[payment.PaymentMethodID]; ok {
					paymentMethodName = pm.Name
				}

				payTime := ""
				if payment.PayTime != nil {
					payTime = payment.PayTime.Format("2006-01-02 15:04:05")
				}

				createdAt := ""
				if payment.CreatedAt != nil && !payment.CreatedAt.IsZero() {
					createdAt = payment.CreatedAt.ToDateTimeString()
				}

				row := []string{
					cast.ToString(payment.ID),
					payment.PaymentNo,
					payment.OrderNo,
					paymentMethodName,
					cast.ToString(payment.UserID),
					fmt.Sprintf("%.2f", payment.Amount),
					r.translateStatus(payment.Status, lang),
					payment.ThirdPartyNo,
					payTime,
					payment.FailReason,
					payment.Remark,
					createdAt,
				}
				if err := w.Write(row); err != nil {
					return fmt.Errorf("写入CSV失败: %v", err)
				}
			}

			// 更新游标
			lastPayment := payments[len(payments)-1]
			prevID := lastID
			if lastPayment.CreatedAt != nil && !lastPayment.CreatedAt.IsZero() {
				lastTimeStr = lastPayment.CreatedAt.ToDateTimeString()
			} else {
				lastTimeStr = ""
			}
			lastID = lastPayment.ID
			if lastID == prevID {
				return fmt.Errorf("导出游标未推进，可能导致死循环: table=%s, last_id=%d", tableName, lastID)
			}

			if len(payments) < chunkSize {
				break
			}
		}
	}

	return nil
}

// loadPaymentMethods 加载支付方式映射
func (r *ExportPayments) loadPaymentMethods() map[uint]models.PaymentMethod {
	var methods []models.PaymentMethod
	facades.Orm().Query().Model(&models.PaymentMethod{}).Get(&methods)

	result := make(map[uint]models.PaymentMethod)
	for _, m := range methods {
		result[m.ID] = m
	}
	return result
}

// translateStatus 翻译状态
func (r *ExportPayments) translateStatus(status, lang string) string {
	// 使用多语言翻译键，如 export_payment_status_pending
	key := fmt.Sprintf("export_payment_status_%s", status)
	return utils.TranslateKey(key, lang, status)
}

// parseOrderBy 解析排序字符串
func (r *ExportPayments) parseOrderBy(orderBy string) (field, direction string) {
	parts := strings.Split(orderBy, ":")
	field = "created_at"
	direction = "desc"

	if len(parts) >= 1 && parts[0] != "" {
		field = parts[0]
	}
	if len(parts) >= 2 && (parts[1] == "asc" || parts[1] == "desc") {
		direction = parts[1]
	}
	return
}
