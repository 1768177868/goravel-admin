package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	Language string                 `json:"language"`  // 语言代码，如 "cn" 或 "en"
}

// ExportOrders 导出订单异步任务
type ExportOrders struct {
}

func (r *ExportOrders) Signature() string {
	return "export_orders"
}

func (r *ExportOrders) Handle(args ...any) error {
	if len(args) < 1 {
		facades.Log().Errorf("ExportOrders Job 参数不足: args=%v", args)
		return apperrors.ErrInvalidArgument.WithMessage("missing export arguments")
	}

	// 解析参数
	var exportArgs ExportOrdersArgs
	switch v := args[0].(type) {
	case ExportOrdersArgs:
		exportArgs = v
	case string:
		// JSON 字符串，需要反序列化
		if err := json.Unmarshal([]byte(v), &exportArgs); err != nil {
			facades.Log().Errorf("反序列化参数失败: %v, JSON: %s", err, v)
			return apperrors.ErrInvalidArgument.WithMessage(fmt.Sprintf("failed to unmarshal export arguments: %v", err))
		}
	case map[string]any:
		// 直接使用 map，框架应该已经解包了
		if exportID, ok := v["export_id"].(float64); ok {
			exportArgs.ExportID = uint(exportID)
		} else if exportID, ok := v["export_id"].(uint); ok {
			exportArgs.ExportID = exportID
		} else if exportID, ok := v["export_id"].(int); ok {
			exportArgs.ExportID = uint(exportID)
		}
		if adminID, ok := v["admin_id"].(float64); ok {
			exportArgs.AdminID = uint(adminID)
		} else if adminID, ok := v["admin_id"].(uint); ok {
			exportArgs.AdminID = adminID
		} else if adminID, ok := v["admin_id"].(int); ok {
			exportArgs.AdminID = uint(adminID)
		}
		if filters, ok := v["filters"].(map[string]interface{}); ok {
			exportArgs.Filters = filters
		}
		if exportType, ok := v["type"].(string); ok {
			exportArgs.Type = exportType
		}
	default:
		facades.Log().Errorf("不支持的参数类型: %T, 值: %+v", args[0], args[0])
		return apperrors.ErrInvalidArgument.WithMessage(fmt.Sprintf("invalid export arguments type: %T", args[0]))
	}

	if exportArgs.ExportID == 0 {
		facades.Log().Errorf("export_id 为 0，参数解析失败: %+v", exportArgs)
		return apperrors.ErrInvalidArgument.WithMessage("export_id is required")
	}

	// 更新导出状态为处理中
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportArgs.ExportID).First(&exportRecord); err != nil {
		errorlog.Record(context.TODO(), "export", "导出记录不存在", map[string]any{
			"export_id": exportArgs.ExportID,
		}, "导出记录不存在: %v", err)
		return err
	}

	exportRecord.Status = models.ExportStatusProcessing
	exportRecord.ErrorMsg = "" // 清空之前的错误信息
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		facades.Log().Errorf("更新导出状态为处理中失败: export_id=%d, error=%v", exportArgs.ExportID, err)
		errorlog.Record(context.TODO(), "export", "更新导出状态失败", map[string]any{
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
		// 记录错误日志（先记录，确保有日志）
		errorMsg := err.Error()
		if errorMsg == "" {
			errorMsg = "未知错误"
		}

		facades.Log().Errorf("导出任务失败: export_id=%d, error=%s", exportArgs.ExportID, errorMsg)
		errorlog.Record(context.TODO(), "export", "导出失败", map[string]any{
			"export_id": exportArgs.ExportID,
			"error":     errorMsg,
		}, "导出失败: %v", err)

		// 重新查询导出记录并更新
		var failedRecord models.Export
		if queryErr := facades.Orm().Query().Where("id", exportArgs.ExportID).First(&failedRecord); queryErr == nil {
			failedRecord.Status = models.ExportStatusFailed
			failedRecord.ErrorMsg = errorMsg // 确保不是空字符串
			if saveErr := facades.Orm().Query().Save(&failedRecord); saveErr != nil {
				facades.Log().Errorf("更新导出记录失败状态失败: export_id=%d, error=%v", exportArgs.ExportID, saveErr)
			} else {
				facades.Log().Infof("已更新导出记录为失败状态: export_id=%d", exportArgs.ExportID)
			}
		} else {
			facades.Log().Errorf("查询导出记录失败: export_id=%d, error=%v", exportArgs.ExportID, queryErr)
		}

		return err
	}

	// 导出成功，状态已在 exportOrders 方法中更新
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
		errorlog.Record(context.TODO(), "export", "获取订单数据失败", map[string]any{
			"export_id": args.ExportID,
			"filters":   args.Filters,
			"error":     err.Error(),
		}, "获取订单数据失败: %v", err)
		return fmt.Errorf("获取订单数据失败: %v", err)
	}

	// 检查是否有数据
	if len(ordersWithDetails) == 0 {
		facades.Log().Infof("导出订单: 没有符合条件的订单数据, export_id=%d", args.ExportID)
		// 即使没有数据，也创建一个空的CSV文件
	}

	// 准备表头（翻译键，需要翻译）
	headerKeys := []string{
		"export_header_id",
		"export_header_order_no",
		"export_header_user_id",
		"export_header_amount",
		"export_header_status",
		"export_header_item_index",   // 商品序号
		"export_header_product_id",   // 商品ID
		"export_header_product_name", // 商品名称
		"export_header_price",        // 单价
		"export_header_quantity",     // 数量
		"export_header_subtotal",     // 小计
		"export_header_remark",
		"export_header_created_at",
	}

	// 翻译表头（使用传递的语言）
	// 获取语言代码，如果没有传递则使用默认语言
	lang := args.Language
	if lang == "" {
		lang = facades.Config().GetString("app.locale", "cn")
	}
	// 验证语言是否支持
	if lang != "cn" && lang != "en" {
		lang = facades.Config().GetString("app.locale", "cn")
	}

	// 直接读取语言文件进行翻译
	headers := r.translateHeaders(headerKeys, lang)

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
	// 注意：Job 中没有 http.Context，传入 nil 是安全的，因为 skipAutoCreate=true 且代码已处理 nil 情况
	exportService := services.NewExportService(nil) //nolint:staticcheck // Job 中没有 http.Context
	filename := fmt.Sprintf("orders_%d", time.Now().Unix())

	filePath, err := exportService.ExportToCSV(headers, data, filename, true) // skipAutoCreate=true
	if err != nil {
		errorlog.Record(context.TODO(), "export", "导出文件失败", map[string]any{
			"export_id": args.ExportID,
			"filename":  filename,
			"error":     err.Error(),
		}, "导出文件失败: %v", err)
		return fmt.Errorf("导出文件失败: %v", err)
	}

	// 更新导出记录的文件路径和大小
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", args.ExportID).First(&exportRecord); err != nil {
		facades.Log().Errorf("查询导出记录失败: export_id=%d, error=%v", args.ExportID, err)
		return fmt.Errorf("查询导出记录失败: %v", err)
	}

	exportRecord.Path = filePath
	exportRecord.Filename = filepath.Base(filePath)

	// 获取文件扩展名
	if ext := filepath.Ext(filePath); ext != "" {
		exportRecord.Extension = ext[1:] // 去掉点号
	} else {
		exportRecord.Extension = "csv" // 默认扩展名
	}

	// 获取文件大小
	storage := facades.Storage().Disk(exportRecord.Disk)
	if fileInfo, err := storage.Size(filePath); err == nil {
		exportRecord.Size = fileInfo
	} else {
		facades.Log().Warningf("获取文件大小失败: export_id=%d, error=%v", args.ExportID, err)
		exportRecord.Size = 0
	}

	// 更新状态为成功
	exportRecord.Status = models.ExportStatusSuccess
	exportRecord.ErrorMsg = "" // 清空错误信息

	// 保存更新
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		facades.Log().Errorf("保存导出记录失败: export_id=%d, error=%v", args.ExportID, err)
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

// translateHeaders 翻译表头（直接读取语言文件）
func (r *ExportOrders) translateHeaders(headerKeys []string, lang string) []string {
	headers := make([]string, len(headerKeys))

	// 获取语言文件路径（使用框架配置的路径）
	langPath := facades.Config().GetString("app.lang_path", "lang")
	langFile := filepath.Join(langPath, fmt.Sprintf("%s.json", lang))

	// 尝试读取语言文件
	langData, err := os.ReadFile(langFile)
	if err != nil {
		// 如果读取失败，使用原始键
		facades.Log().Warningf("读取语言文件失败: %s, error=%v, 使用原始键", langFile, err)
		return append([]string(nil), headerKeys...)
	}

	// 解析 JSON
	var langMap map[string]interface{}
	if err := json.Unmarshal(langData, &langMap); err != nil {
		facades.Log().Warningf("解析语言文件失败: %s, error=%v, 使用原始键", langFile, err)
		return append([]string(nil), headerKeys...)
	}

	// 获取 messages 对象
	messages, ok := langMap["messages"].(map[string]interface{})
	if !ok {
		facades.Log().Warningf("语言文件中没有 messages 对象: %s, 使用原始键", langFile)
		return append([]string(nil), headerKeys...)
	}

	// 翻译每个键
	for i, key := range headerKeys {
		fullKey := "messages." + key
		if value, ok := messages[key].(string); ok && value != "" {
			headers[i] = value
		} else {
			// 如果翻译失败，使用原始键
			headers[i] = key
			facades.Log().Debugf("翻译键未找到: %s (语言: %s)", fullKey, lang)
		}
	}

	return headers
}
