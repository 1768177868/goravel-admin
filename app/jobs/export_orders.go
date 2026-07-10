package jobs

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	appfacades "goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

// ExportOrdersArgs 导出订单任务的参数（类型别名）
type ExportOrdersArgs = ExportArgs

// ExportOrders 订单导出任务
type ExportOrders struct{}

func (r *ExportOrders) Signature() string {
	return "export_orders"
}

func (r *ExportOrders) Handle(args ...any) (retErr error) {
	var exportID uint

	defer func() {
		if rec := recover(); rec != nil {
			errorMsg := fmt.Sprintf("panic: %v", rec)
			facades.Log().Errorf("ExportOrders Job panic: %v", rec)
			MarkExportFailed(exportID, errorMsg)
			retErr = fmt.Errorf("%s", errorMsg)
		}
	}()

	exportArgs, err := ParseArgs(args...)
	if err != nil {
		return err
	}
	exportID = exportArgs.ExportID

	lock, err := AcquireExportExecutionLock(exportID)
	if err != nil {
		return err
	}
	if lock == nil {
		facades.Log().Infof("导出任务已在执行中，跳过重复投递: export_id=%d", exportID)
		return nil
	}
	defer lock.Release()

	exportRecord, err := CheckAndUpdateExportStatus(exportID)
	if err != nil {
		return err
	}
	if exportRecord == nil {
		return nil
	}

	exporter := NewBaseExporter(ExportConfig{
		FilePrefix: "orders",
		HeaderKeys: []string{
			"id",
			"order_no",
			"user_id",
			"amount",
			"status",
			"item_index",
			"product_id",
			"product_name",
			"price",
			"quantity",
			"subtotal",
			"remark",
			"created_at",
		},
		WriteData: r.writeOrdersToCSV,
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

// writeOrdersToCSV 写入订单数据到 CSV
func (r *ExportOrders) writeOrdersToCSV(w *csv.Writer, filters map[string]any, lang string, shouldStop func() bool) error {
	// 构建筛选条件（自动填充，无需手动逐字段赋值）
	var orderFilters services.OrderFilters
	utils.FillFiltersFromMap(filters, &orderFilters)

	// 时间范围需要特殊处理（分表依赖）
	orderFilters.StartTime, orderFilters.EndTime = GetDefaultTimeRange(filters)

	// 获取时区（用于时间格式化）
	timezone, _ := utils.GetString(filters, "_timezone")

	// 获取分表列表
	tableNames := utils.GetShardingTableNames("orders", orderFilters.StartTime, orderFilters.EndTime)
	if len(tableNames) == 0 {
		return nil
	}

	_, direction := ParseOrderBy(orderFilters.OrderBy)
	if direction == "desc" {
		for i, j := 0, len(tableNames)-1; i < j; i, j = i+1, j-1 {
			tableNames[i], tableNames[j] = tableNames[j], tableNames[i]
		}
	}

	const chunkSize = 2000

	for _, tableName := range tableNames {
		if err := r.exportTable(w, tableName, orderFilters, lang, timezone, direction, chunkSize, shouldStop); err != nil {
			return err
		}
	}

	return nil
}

// exportTable 导出单个分表
func (r *ExportOrders) exportTable(w *csv.Writer, tableName string, filters services.OrderFilters, lang, timezone, direction string, chunkSize int, shouldStop func() bool) error {
	suffix := strings.TrimPrefix(tableName, "orders_")
	detailTableName := "order_details_" + suffix

	lastTimeStr := ""
	var lastID uint = 0

	for {
		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
		}

		query := services.BuildOrderQuery(tableName, filters)

		if lastTimeStr != "" {
			if direction == "desc" {
				query = query.Where("(created_at < ? OR (created_at = ? AND id < ?))", lastTimeStr, lastTimeStr, lastID)
			} else {
				query = query.Where("(created_at > ? OR (created_at = ? AND id > ?))", lastTimeStr, lastTimeStr, lastID)
			}
		}

		if direction == "desc" {
			query = query.Order("created_at desc").Order("id desc")
		} else {
			query = query.Order("created_at asc").Order("id asc")
		}

		var orders []models.Order
		if err := query.Limit(chunkSize).Get(&orders); err != nil {
			return fmt.Errorf("查询订单失败: table=%s, err=%v", tableName, err)
		}

		if len(orders) == 0 {
			break
		}

		// 批量查详情
		orderIDsAny := make([]any, 0, len(orders))
		for _, o := range orders {
			orderIDsAny = append(orderIDsAny, o.ID)
		}

		var details []models.OrderDetail
		if len(orderIDsAny) > 0 {
			_ = appfacades.OrmQuery(context.Background()).Table(detailTableName).
				WhereIn("order_id", orderIDsAny).
				Get(&details)
		}

		detailMap := make(map[uint][]models.OrderDetail, len(orders))
		for _, d := range details {
			detailMap[d.OrderID] = append(detailMap[d.OrderID], d)
		}

		// 写入 CSV
		for _, order := range orders {
			if err := r.writeOrderRows(w, order, detailMap[order.ID], lang, timezone); err != nil {
				return err
			}
		}

		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
		}

		// 更新游标
		last := orders[len(orders)-1]
		if last.CreatedAt != nil && !last.CreatedAt.IsZero() {
			lastTimeStr = last.CreatedAt.ToDateTimeString()
		} else {
			lastTimeStr = ""
		}
		lastID = last.ID

		if len(orders) < chunkSize {
			break
		}
	}

	return nil
}

// writeOrderRows 写入订单行（包含详情）
func (r *ExportOrders) writeOrderRows(w *csv.Writer, order models.Order, details []models.OrderDetail, lang, timezone string) error {
	statusText := r.translateStatus(order.Status, lang)

	timeStr := FormatCarbonWithTimezone(order.CreatedAt, timezone)

	if len(details) == 0 {
		row := []string{
			cast.ToString(order.ID),
			order.OrderNo,
			cast.ToString(order.UserID),
			fmt.Sprintf("%.2f", order.Amount),
			statusText,
			"", "", "", "", "", "",
			order.Remark,
			timeStr,
		}
		return w.Write(row)
	}

	totalItems := len(details)
	for idx, detail := range details {
		row := []string{
			cast.ToString(order.ID),
			order.OrderNo,
			cast.ToString(order.UserID),
			fmt.Sprintf("%.2f", order.Amount),
			statusText,
			fmt.Sprintf("%d/%d", idx+1, totalItems),
			cast.ToString(detail.ProductID),
			detail.ProductName,
			fmt.Sprintf("%.2f", detail.Price),
			cast.ToString(detail.Quantity),
			fmt.Sprintf("%.2f", detail.Subtotal),
			order.Remark,
			timeStr,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// translateStatus 翻译订单状态
func (r *ExportOrders) translateStatus(status, lang string) string {
	switch status {
	case "pending":
		return utils.TranslateKey("order_status_pending", lang, "pending")
	case "paid":
		return utils.TranslateKey("order_status_paid", lang, "paid")
	case "cancelled":
		return utils.TranslateKey("order_status_cancelled", lang, "cancelled")
	default:
		return status
	}
}
