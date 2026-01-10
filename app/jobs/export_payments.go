package jobs

import (
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

// ExportPaymentsArgs 导出支付记录任务的参数（类型别名，保持向后兼容）
type ExportPaymentsArgs = ExportArgs

// ExportPayments 支付记录导出任务
type ExportPayments struct{}

func (r *ExportPayments) Signature() string {
	return "export_payments"
}

func (r *ExportPayments) Handle(args ...any) (retErr error) {
	var exportID uint

	// panic 保护
	defer func() {
		if rec := recover(); rec != nil {
			errorMsg := fmt.Sprintf("panic: %v", rec)
			facades.Log().Errorf("ExportPayments Job panic: %v", rec)
			MarkExportFailed(exportID, errorMsg)
			retErr = fmt.Errorf("%s", errorMsg)
		}
	}()

	// 解析参数
	exportArgs, err := ParseArgs(args...)
	if err != nil {
		return err
	}
	exportID = exportArgs.ExportID

	// 检查并更新导出状态
	exportRecord, err := CheckAndUpdateExportStatus(exportID)
	if err != nil {
		return err
	}
	if exportRecord == nil {
		return nil // 记录不存在，正常结束
	}

	// 创建导出器并执行
	exporter := NewBaseExporter(ExportConfig{
		FilePrefix: "payments",
		HeaderKeys: []string{
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
		},
		WriteData: r.writePaymentsToCSV,
	})

	jobErr := exporter.Execute(exportArgs)

	if errors.Is(jobErr, ErrExportRecordMissing) {
		facades.Log().Infof("导出任务检测到导出记录已删除: export_id=%d", exportID)
		return nil
	}

	if jobErr != nil {
		MarkExportFailed(exportID, jobErr.Error())
		return jobErr
	}

	return nil
}

// writePaymentsToCSV 写入支付数据到 CSV（业务特定逻辑）
func (r *ExportPayments) writePaymentsToCSV(w *csv.Writer, filters map[string]any, lang string, shouldStop func() bool) error {
	// 构建筛选条件（自动填充，无需手动逐字段赋值）
	var paymentFilters services.PaymentFilters
	utils.FillFiltersFromMap(filters, &paymentFilters)

	// 时间范围需要特殊处理（分表依赖）
	paymentFilters.StartTime, paymentFilters.EndTime = GetDefaultTimeRange(filters)

	// 获取分表列表
	tableNames := r.getTableNames(paymentFilters)
	if len(tableNames) == 0 {
		return nil
	}

	// 排序处理
	_, direction := ParseOrderBy(paymentFilters.OrderBy)
	if direction == "desc" {
		for i, j := 0, len(tableNames)-1; i < j; i, j = i+1, j-1 {
			tableNames[i], tableNames[j] = tableNames[j], tableNames[i]
		}
	}

	// 预加载支付方式
	paymentMethodMap := r.loadPaymentMethods()

	const chunkSize = 2000

	for _, tableName := range tableNames {
		if err := r.exportTable(w, tableName, paymentFilters, paymentMethodMap, lang, direction, chunkSize, shouldStop); err != nil {
			return err
		}
	}

	return nil
}

// getTableNames 获取分表列表
func (r *ExportPayments) getTableNames(filters services.PaymentFilters) []string {
	// 如果精确指定了 payment_no，直接定位分表
	if filters.PaymentNo != "" && len(filters.PaymentNo) >= 11 {
		dateStr := filters.PaymentNo[3:11]
		if t, err := time.Parse("20060102", dateStr); err == nil {
			return []string{utils.GetShardingTableName("payments", t)}
		}
	}
	return utils.GetShardingTableNames("payments", filters.StartTime, filters.EndTime)
}

// exportTable 导出单个分表
func (r *ExportPayments) exportTable(w *csv.Writer, tableName string, filters services.PaymentFilters, paymentMethodMap map[uint]models.PaymentMethod, lang, direction string, chunkSize int, shouldStop func() bool) error {
	lastTimeStr := ""
	var lastID uint = 0
	field := "created_at"

	for {
		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
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
			if IsTableNotExistsError(err) {
				facades.Log().Warningf("支付记录分表不存在，跳过: table=%s", tableName)
				return nil
			}
			return fmt.Errorf("查询支付记录失败: %v", err)
		}

		if len(payments) == 0 {
			break
		}

		// 写入 CSV
		for _, payment := range payments {
			row := r.formatPaymentRow(payment, paymentMethodMap, lang)
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
			return fmt.Errorf("导出游标未推进: table=%s, last_id=%d", tableName, lastID)
		}

		if len(payments) < chunkSize {
			break
		}
	}

	return nil
}

// formatPaymentRow 格式化单行数据
func (r *ExportPayments) formatPaymentRow(payment models.Payment, paymentMethodMap map[uint]models.PaymentMethod, lang string) []string {
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

	statusKey := fmt.Sprintf("export_payment_status_%s", payment.Status)

	return []string{
		cast.ToString(payment.ID),
		payment.PaymentNo,
		payment.OrderNo,
		paymentMethodName,
		cast.ToString(payment.UserID),
		fmt.Sprintf("%.2f", payment.Amount),
		utils.TranslateKey(statusKey, lang, payment.Status),
		payment.ThirdPartyNo,
		payTime,
		payment.FailReason,
		payment.Remark,
		createdAt,
	}
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
