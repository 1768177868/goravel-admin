package services

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

// ExportOrderService 订单导出服务
type ExportOrderService struct {
	ctx http.Context
}

func NewExportOrderService(ctx http.Context) *ExportOrderService {
	return &ExportOrderService{ctx: ctx}
}

// ExportOrders 同步导出订单
func (s *ExportOrderService) ExportOrders(exportID uint, filters OrderFilters) error {
	// 更新导出状态为处理中
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportID).FirstOrFail(&exportRecord); err != nil {
		return fmt.Errorf("查询导出记录失败: %v", err)
	}

	facades.Log().Infof("开始处理导出任务: export_id=%d", exportID)

	exportRecord.Status = models.ExportStatusProcessing
	exportRecord.ErrorMsg = ""
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		return fmt.Errorf("更新导出状态失败: %v", err)
	}

	// 获取订单数据
	orderService := NewOrderService()
	ordersWithDetails, err := orderService.GetAllOrdersWithDetailsForExport(filters)
	if err != nil {
		errorMsg := fmt.Sprintf("获取订单数据失败: %v", err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = errorMsg
		facades.Orm().Query().Save(&exportRecord)
		return errors.New(errorMsg)
	}

	// 检查是否有数据
	if len(ordersWithDetails) == 0 {
		facades.Log().Infof("导出订单: 没有符合条件的订单数据, export_id=%d", exportID)
	}

	// 准备表头（翻译键，需要翻译）
	headerKeys := []string{
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
	}

	// 翻译表头
	headers := make([]string, len(headerKeys))
	for i, key := range headerKeys {
		translated := facades.Lang(s.ctx).Get("messages." + key)
		if translated == "messages."+key || translated == "" {
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
				"",
				"",
				"",
				"",
				"",
				"",
				order.Remark,
				timeStr,
			}
			data = append(data, row)
		} else {
			// 每个商品一行，添加商品序号
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

	// 使用 ExportService 导出（跳过自动创建记录）
	exportService := NewExportService(s.ctx)
	filename := fmt.Sprintf("orders_%d", time.Now().Unix())

	facades.Log().Infof("开始导出文件: export_id=%d, filename=%s, data_rows=%d", exportID, filename, len(data))

	filePath, err := exportService.ExportToCSV(headers, data, filename, true) // skipAutoCreate=true
	if err != nil {
		errorMsg := fmt.Sprintf("导出文件失败: %v", err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = errorMsg
		facades.Orm().Query().Save(&exportRecord)
		errorlog.RecordHTTP(s.ctx, "export", "导出文件失败", map[string]any{
			"export_id": exportID,
			"filename":  filename,
			"error":     err.Error(),
		}, "导出文件失败: %v", err)
		return fmt.Errorf("导出文件失败: %v", err)
	}

	facades.Log().Infof("文件导出成功: export_id=%d, file_path=%s", exportID, filePath)

	// 更新导出记录的文件路径和大小
	if err := facades.Orm().Query().Where("id", exportID).FirstOrFail(&exportRecord); err != nil {
		facades.Log().Errorf("查询导出记录失败: export_id=%d, error=%v", exportID, err)
		return fmt.Errorf("查询导出记录失败: %v", err)
	}

	facades.Log().Infof("开始更新导出记录: export_id=%d, file_path=%s", exportID, filePath)

	exportRecord.Path = filePath
	exportRecord.Filename = filepath.Base(filePath)

	// 获取文件扩展名
	if ext := filepath.Ext(filePath); ext != "" {
		exportRecord.Extension = ext[1:]
	} else {
		exportRecord.Extension = "csv"
	}

	// 获取文件大小
	storage := facades.Storage().Disk(exportRecord.Disk)
	if fileInfo, err := storage.Size(filePath); err == nil {
		exportRecord.Size = fileInfo
		facades.Log().Infof("文件大小: export_id=%d, size=%d", exportID, fileInfo)
	} else {
		facades.Log().Warningf("获取文件大小失败: export_id=%d, error=%v", exportID, err)
		exportRecord.Size = 0
	}

	// 更新状态为成功
	exportRecord.Status = models.ExportStatusSuccess
	exportRecord.ErrorMsg = ""

	// 保存更新
	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		facades.Log().Errorf("保存导出记录失败: export_id=%d, error=%v", exportID, err)
		return fmt.Errorf("更新导出记录失败: %v", err)
	}

	facades.Log().Infof("导出成功: export_id=%d, file_path=%s, filename=%s, size=%d, extension=%s",
		exportID, filePath, exportRecord.Filename, exportRecord.Size, exportRecord.Extension)

	return nil
}
