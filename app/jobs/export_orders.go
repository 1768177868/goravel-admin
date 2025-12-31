package jobs

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"
)

// ExportOrdersArgs 导出订单任务的参数
type ExportOrdersArgs struct {
	ExportID uint                   `json:"export_id"` // 导出记录ID
	AdminID  uint                   `json:"admin_id"`  // 管理员ID
	Filters  map[string]interface{} `json:"filters"`   // 筛选条件（JSON序列化）
	Type     string                 `json:"type"`      // 导出类型，如 "orders"
}

// ExportOrders 导出订单异步任务
type ExportOrders struct {
}

func (r *ExportOrders) Signature() string {
	return "export_orders"
}

func (r *ExportOrders) Handle(args ...any) error {
	if len(args) < 1 {
		return apperrors.ErrInvalidArgument.WithMessage("missing export arguments")
	}

	// 解析参数
	var exportArgs ExportOrdersArgs
	switch v := args[0].(type) {
	case ExportOrdersArgs:
		exportArgs = v
	case map[string]any:
		if exportID, ok := v["export_id"].(float64); ok {
			exportArgs.ExportID = uint(exportID)
		} else if exportID, ok := v["export_id"].(uint); ok {
			exportArgs.ExportID = exportID
		}
		if adminID, ok := v["admin_id"].(float64); ok {
			exportArgs.AdminID = uint(adminID)
		} else if adminID, ok := v["admin_id"].(uint); ok {
			exportArgs.AdminID = adminID
		}
		if filters, ok := v["filters"].(map[string]interface{}); ok {
			exportArgs.Filters = filters
		}
		if exportType, ok := v["type"].(string); ok {
			exportArgs.Type = exportType
		}
	default:
		return apperrors.ErrInvalidArgument.WithMessage("invalid export arguments")
	}

	if exportArgs.ExportID == 0 {
		return apperrors.ErrInvalidArgument.WithMessage("export_id is required")
	}

	// 更新导出状态为处理中
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportArgs.ExportID).First(&exportRecord); err != nil {
		errorlog.Record(nil, "export", "导出记录不存在", map[string]any{
			"export_id": exportArgs.ExportID,
		}, "导出记录不存在: %v", err)
		return err
	}

	// 开始处理
	exportRecord.Status = models.ExportStatusProcessing
	exportRecord.ErrorMsg = ""
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		errorlog.Record(nil, "export", "更新导出状态失败", map[string]any{
			"export_id": exportArgs.ExportID,
		}, "更新导出状态失败: %v", err)
		return err
	}

	// 根据类型执行不同的导出逻辑
	var err error
	switch exportArgs.Type {
	case "orders":
		err = r.exportOrders(exportArgs)
	default:
		err = fmt.Errorf("不支持的导出类型: %s", exportArgs.Type)
	}

	// 更新导出状态
	if err != nil {
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return err
	}

	// 重新查询获取最新信息（包括文件路径等）
	if err := facades.Orm().Query().Where("id", exportArgs.ExportID).First(&exportRecord); err == nil {
		exportRecord.Status = models.ExportStatusSuccess
		facades.Orm().Query().Save(&exportRecord)
	}

	return nil
}

// exportOrders 执行订单导出
func (r *ExportOrders) exportOrders(args ExportOrdersArgs) error {
	// 获取订单服务
	orderService := services.NewOrderService()

	// 构建筛选条件（从 args.Filters 反序列化）
	filters := services.OrderFilters{}

	// 解析筛选条件
	if userID, ok := args.Filters["user_id"].(float64); ok {
		filters.UserID = uint(userID)
	} else if userID, ok := args.Filters["user_id"].(uint); ok {
		filters.UserID = userID
	}
	if orderNo, ok := args.Filters["order_no"].(string); ok {
		filters.OrderNo = orderNo
	}
	if status, ok := args.Filters["status"].(string); ok {
		filters.Status = status
	}
	if minAmount, ok := args.Filters["min_amount"].(float64); ok {
		filters.MinAmount = minAmount
	}
	if maxAmount, ok := args.Filters["max_amount"].(float64); ok {
		filters.MaxAmount = maxAmount
	}
	if orderBy, ok := args.Filters["order_by"].(string); ok {
		filters.OrderBy = orderBy
	}

	// 解析时间
	if startTimeStr, ok := args.Filters["start_time"].(string); ok && startTimeStr != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			filters.StartTime = t.UTC()
		}
	}
	if endTimeStr, ok := args.Filters["end_time"].(string); ok && endTimeStr != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			filters.EndTime = t.UTC()
		}
	}

	// 获取订单数据
	ordersWithDetails, err := orderService.GetAllOrdersWithDetailsForExport(filters)
	if err != nil {
		return fmt.Errorf("获取订单数据失败: %v", err)
	}

	// 准备表头（翻译键，需要翻译）
	headerKeys := []string{
		"export_header_id",
		"export_header_order_no",
		"export_header_user_id",
		"export_header_amount",
		"export_header_status",
		"export_header_item_index",    // 商品序号
		"export_header_product_id",    // 商品ID
		"export_header_product_name",  // 商品名称
		"export_header_price",         // 单价
		"export_header_quantity",      // 数量
		"export_header_subtotal",      // 小计
		"export_header_remark",
		"export_header_created_at",
	}
	
	// 翻译表头（使用默认语言 cn）
	headers := make([]string, len(headerKeys))
	for i, key := range headerKeys {
		translated := facades.Lang(nil).Get("messages." + key)
		if translated == "messages."+key || translated == "" {
			// 如果翻译失败，使用原始键
			headers[i] = key
		} else {
			headers[i] = translated
		}
	}

	// 准备数据（展开模式：每个商品一行）
	var data [][]string
	for _, orderWithDetails := range ordersWithDetails {
		order := orderWithDetails.Order
		details := orderWithDetails.Details

		// 格式化订单状态
		statusText := order.Status
		switch order.Status {
		case "pending":
			statusText = "待支付"
		case "paid":
			statusText = "已支付"
		case "cancelled":
			statusText = "已取消"
		}

		// 格式化时间
		timeStr := ""
		if order.CreatedAt != nil && !order.CreatedAt.IsZero() {
			timeStr = order.CreatedAt.ToDateTimeString()
		}

		// 如果订单没有商品，至少输出一行订单信息
		if len(details) == 0 {
			row := []string{
				cast.ToString(order.ID),
				order.OrderNo,
				cast.ToString(order.UserID),
				fmt.Sprintf("%.2f", order.Amount),
				statusText,
				"", // 商品序号
				"", // 商品ID
				"", // 商品名称
				"", // 单价
				"", // 数量
				"", // 小计
				order.Remark,
				timeStr,
			}
			data = append(data, row)
		} else {
			// 每个商品一行，添加商品序号（如：1/2, 2/2）
			totalItems := len(details)
			for idx, detail := range details {
				itemIndex := fmt.Sprintf("%d/%d", idx+1, totalItems)
				row := []string{
					cast.ToString(order.ID),
					order.OrderNo,
					cast.ToString(order.UserID),
					fmt.Sprintf("%.2f", order.Amount),
					statusText,
					itemIndex,
					cast.ToString(detail.ProductID),
					detail.ProductName,
					fmt.Sprintf("%.2f", detail.Price),
					cast.ToString(detail.Quantity),
					fmt.Sprintf("%.2f", detail.Subtotal),
					order.Remark,
					timeStr,
				}
				data = append(data, row)
			}
		}
	}

	// 使用 ExportService 导出（跳过自动创建记录，因为我们已经有了导出记录）
	exportService := services.NewExportService(nil) // 异步任务没有 context
	filename := fmt.Sprintf("orders_%d", time.Now().Unix())
	filePath, err := exportService.ExportToCSV(headers, data, filename, true) // skipAutoCreate=true
	if err != nil {
		return fmt.Errorf("导出文件失败: %v", err)
	}

	// 更新导出记录的文件路径和大小
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", args.ExportID).First(&exportRecord); err != nil {
		return fmt.Errorf("查询导出记录失败: %v", err)
	}

	exportRecord.Path = filePath
	exportRecord.Filename = filepath.Base(filePath)

	// 获取文件扩展名
	if ext := filepath.Ext(filePath); ext != "" {
		exportRecord.Extension = ext[1:] // 去掉点号
	}

	// 获取文件大小
	storage := facades.Storage().Disk(exportRecord.Disk)
	if fileInfo, err := storage.Size(filePath); err == nil {
		exportRecord.Size = fileInfo
	}

	// 保存更新
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		return fmt.Errorf("更新导出记录失败: %v", err)
	}

	return nil
}

func (r *ExportOrders) ShouldRetry(err error) bool {
	// 如果是业务错误，不重试
	if apperrors.IsBusinessError(err) {
		return false
	}
	// 其他错误可以重试
	return true
}

func (r *ExportOrders) RetryAfter(err error) time.Duration {
	// 递增延迟：第1次重试延迟5秒，第2次10秒，第3次20秒...
	return 5 * time.Second
}

