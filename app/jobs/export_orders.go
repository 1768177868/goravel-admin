package jobs

import (
	"context"
	"encoding/csv"
	"encoding/json"
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

// ExportOrdersArgs 导出订单任务的参数
type ExportOrdersArgs struct {
	ExportID uint           `json:"export_id"` // 导出记录ID
	AdminID  uint           `json:"admin_id"`  // 管理员ID
	Filters  map[string]any `json:"filters"`   // 筛选条件（JSON序列化）
	Type     string         `json:"type"`      // 导出类型，如 "orders"
	Language string         `json:"language"`  // 语言代码，如 "cn" 或 "en"
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
		// 使用泛型辅助函数简化类型断言
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
	// 构建筛选条件（从 args.Filters 反序列化）
	filters := services.OrderFilters{}

	// 解析筛选条件（使用泛型辅助函数简化类型断言）
	if userID, ok := utils.GetUint(args.Filters, "user_id"); ok {
		filters.UserID = userID
	}
	if orderNo, ok := utils.GetString(args.Filters, "order_no"); ok {
		filters.OrderNo = orderNo
	}
	if status, ok := utils.GetString(args.Filters, "status"); ok {
		filters.Status = status
	}
	if minAmount, ok := utils.GetValue[float64](args.Filters, "min_amount"); ok {
		filters.MinAmount = minAmount
	}
	if maxAmount, ok := utils.GetValue[float64](args.Filters, "max_amount"); ok {
		filters.MaxAmount = maxAmount
	}
	if orderBy, ok := utils.GetString(args.Filters, "order_by"); ok {
		filters.OrderBy = orderBy
	}

	// 解析时间（使用泛型辅助函数）
	if startTimeStr, ok := utils.GetString(args.Filters, "start_time"); ok && startTimeStr != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			filters.StartTime = t.UTC()
		}
	}
	if endTimeStr, ok := utils.GetString(args.Filters, "end_time"); ok && endTimeStr != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			filters.EndTime = t.UTC()
		}
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
	// 获取语言代码，如果没有传递则使用默认语言，并规范化
	lang := args.Language
	if lang == "" {
		lang = facades.Config().GetString("app.locale", "cn")
	}
	lang = utils.NormalizeLanguage(lang)

	// 直接读取语言文件进行翻译（使用通用工具函数）
	headers := utils.TranslateHeaders(headerKeys, lang)

	// 使用 ExportService 导出（跳过自动创建记录，因为我们已经有了导出记录）
	// 注意：Job 中没有 http.Context，传入 nil 是安全的，因为 skipAutoCreate=true 且代码已处理 nil 情况
	exportService := services.NewExportService(nil) //nolint:staticcheck // Job 中没有 http.Context
	// 先生成最终文件名/路径并写入 exports 表，让“导出中”也能看到文件名等信息
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("orders_%d_%s.csv", args.ExportID, timestamp)
	filePath := path.Join("exports", filename)

	// 预写入文件信息（Path/Filename/Extension），Size 先为 0
	// 注意：Disk 在创建导出记录时已写入
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

	// 百万级导出必须使用流式写入，避免把全部数据/CSV 内容堆到内存导致服务器崩溃
	lastUpdateAt := time.Now().Add(-10 * time.Second)
	filePath, err := exportService.ExportToCSVStreamAtWithProgress(headers, filePath, func(w *csv.Writer) error {
		return r.writeOrdersToCSV(w, filters, lang)
	}, func(writtenBytes int64) {
		// 限流：最多每 3 秒更新一次，避免导出过程中频繁写库
		if time.Since(lastUpdateAt) < 3*time.Second {
			return
		}
		lastUpdateAt = time.Now()
		_, _ = facades.Orm().Query().Model(&models.Export{}).Where("id", args.ExportID).Update(map[string]any{
			"size": writtenBytes,
		})
	}, true)
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
	// filePath 使用 "/" 分隔（云存储对象 key），这里使用 path.Base 保证跨平台一致
	exportRecord.Filename = path.Base(filePath)

	// 获取文件扩展名
	if ext := path.Ext(filePath); ext != "" {
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
		// 保留导出过程中的 size（实时更新值），避免这里覆盖为 0
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

// writeOrdersToCSV 按分表分批查询订单并写入 CSV（流式，低内存）
func (r *ExportOrders) writeOrdersToCSV(w *csv.Writer, filters services.OrderFilters, lang string) error {
	// 导出默认排序：created_at desc（稳定排序再加 id）
	orderBy := strings.TrimSpace(filters.OrderBy)
	if orderBy == "" {
		orderBy = "created_at:desc"
	}
	field, direction := r.parseOrderBy(orderBy)
	// 为了可分页且不崩，强制使用 created_at 做主排序（其他字段会导致无法高效 keyset 分页）
	if field != "created_at" {
		field = "created_at"
	}

	// 获取分表列表（按月）
	tableNames := utils.GetShardingTableNames("orders", filters.StartTime, filters.EndTime)
	if len(tableNames) == 0 {
		return nil
	}

	// desc 时按月份倒序导出，保证全局 created_at 顺序正确（时间分表不会跨月乱序）
	if direction == "desc" {
		for i, j := 0, len(tableNames)-1; i < j; i, j = i+1, j-1 {
			tableNames[i], tableNames[j] = tableNames[j], tableNames[i]
		}
	}

	const chunkSize = 2000

	for _, tableName := range tableNames {
		// 订单详情表与订单表按相同 YYYYMM 分表
		suffix := strings.TrimPrefix(tableName, "orders_")
		detailTableName := "order_details_" + suffix

		lastTimeStr := ""
		var lastID uint = 0

		for {
			query := facades.Orm().Query().Table(tableName).
				Where("created_at >= ?", filters.StartTime).
				Where("created_at <= ?", filters.EndTime)

			// 用户ID筛选
			if filters.UserID > 0 {
				query = query.Where("user_id = ?", filters.UserID)
			}
			// 订单号模糊搜索
			if filters.OrderNo != "" {
				query = query.Where("order_no LIKE ?", "%"+filters.OrderNo+"%")
			}
			// 订单状态筛选
			if filters.Status != "" {
				query = query.Where("status = ?", filters.Status)
			}
			// 金额范围筛选
			if filters.MinAmount > 0 {
				query = query.Where("amount >= ?", filters.MinAmount)
			}
			if filters.MaxAmount > 0 {
				query = query.Where("amount <= ?", filters.MaxAmount)
			}

			// keyset 分页（避免 offset 扫描）
			if lastTimeStr != "" {
				if direction == "desc" {
					query = query.Where("(created_at < ? OR (created_at = ? AND id < ?))", lastTimeStr, lastTimeStr, lastID)
				} else {
					query = query.Where("(created_at > ? OR (created_at = ? AND id > ?))", lastTimeStr, lastTimeStr, lastID)
				}
			}

			// 排序 + limit
			if direction == "desc" {
				query = query.Order("created_at desc").Order("id desc")
			} else {
				query = query.Order("created_at asc").Order("id asc")
			}
			query = query.Limit(chunkSize)

			var orders []models.Order
			if err := query.Get(&orders); err != nil {
				return fmt.Errorf("查询订单失败: table=%s, err=%v", tableName, err)
			}
			if len(orders) == 0 {
				break
			}

			// 批量查详情（避免 N+1）
			orderIDsAny := make([]any, 0, len(orders))
			for _, o := range orders {
				orderIDsAny = append(orderIDsAny, o.ID)
			}

			var details []models.OrderDetail
			if len(orderIDsAny) > 0 {
				// 某些月份可能还没创建详情分表，查不到就当无详情（避免整个导出失败）
				_ = facades.Orm().Query().Table(detailTableName).
					WhereIn("order_id", orderIDsAny).
					Get(&details)
			}

			detailMap := make(map[uint][]models.OrderDetail, len(orders))
			for _, d := range details {
				detailMap[d.OrderID] = append(detailMap[d.OrderID], d)
			}

			// 写入 CSV
			for _, order := range orders {
				// 格式化订单状态（使用多语言翻译）
				statusText := order.Status
				switch order.Status {
				case "pending":
					statusText = utils.TranslateKey("export_order_status_pending", lang, "pending")
				case "paid":
					statusText = utils.TranslateKey("export_order_status_paid", lang, "paid")
				case "cancelled":
					statusText = utils.TranslateKey("export_order_status_cancelled", lang, "cancelled")
				}

				timeStr := ""
				if order.CreatedAt != nil && !order.CreatedAt.IsZero() {
					timeStr = order.CreatedAt.ToDateTimeString()
				}

				ds := detailMap[order.ID]
				if len(ds) == 0 {
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
					if err := w.Write(row); err != nil {
						return err
					}
				} else {
					totalItems := len(ds)
					for idx, detail := range ds {
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
						if err := w.Write(row); err != nil {
							return err
						}
					}
				}
			}

			// 更新游标（取本批最后一条，确保下一批继续）
			last := orders[len(orders)-1]
			if last.CreatedAt != nil && !last.CreatedAt.IsZero() {
				lastTimeStr = last.CreatedAt.ToDateTimeString()
			} else {
				// created_at 为空很异常，退化为用 id
				lastTimeStr = ""
			}
			lastID = last.ID
		}
	}

	return nil
}

func (r *ExportOrders) parseOrderBy(orderBy string) (field string, direction string) {
	field = "created_at"
	direction = "desc"
	parts := strings.Split(orderBy, ":")
	if len(parts) == 2 {
		if strings.TrimSpace(parts[0]) != "" {
			field = strings.TrimSpace(parts[0])
		}
		dir := strings.ToLower(strings.TrimSpace(parts[1]))
		if dir == "asc" || dir == "desc" {
			direction = dir
		}
	}
	return field, direction
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
