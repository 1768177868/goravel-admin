package admin

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"
)

type ExportController struct {
	exportRecordService services.ExportRecordService
}

func NewExportController() *ExportController {
	return &ExportController{
		exportRecordService: services.NewExportRecordService(),
	}
}

// Index 导出记录列表
func (r *ExportController) Index(ctx http.Context) http.Response {
	page, pageSize := helpers.PaginationFromQuery(ctx, helpers.PaginationLimits{})

	filters := r.buildFilters(ctx)

	exports, total, err := r.exportRecordService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 为每个导出记录生成可访问的 file_url
	exportService := services.NewExportService(ctx)
	type ExportWithURL struct {
		models.Export
		FileURL string `json:"file_url"`
	}

	var resultWithURL []ExportWithURL
	for _, e := range exports {
		fileURL := ""
		if e.Path != "" {
			// 对于 local 和 public 存储，使用下载接口
			if e.Disk == "local" || e.Disk == "public" {
				fileURL = fmt.Sprintf("/api/admin/exports/%d/download", e.ID)
			} else {
				// 对于云存储，使用 GetExportURL 生成 URL
				fileURL = exportService.GetExportURL(e.Path)
			}
		}
		resultWithURL = append(resultWithURL, ExportWithURL{
			Export:  e,
			FileURL: fileURL,
		})
	}

	return response.Paginate(ctx, resultWithURL, total, page, pageSize)
}

// buildFilters 构建导出记录查询过滤器
func (r *ExportController) buildFilters(ctx http.Context) services.ExportRecordFilters {
	adminID := ctx.Request().Query("admin_id", "")
	exportType := ctx.Request().Query("type", "")
	filename := ctx.Request().Query("filename", "")
	disk := ctx.Request().Query("disk", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.ExportRecordFilters{
		AdminID:   adminID,
		Type:      exportType,
		Filename:  filename,
		Disk:      disk,
		Status:    status,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Destroy 删除导出记录并删除源文件
func (r *ExportController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	export, err := r.exportRecordService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	// 尝试删除源文件（忽略失败，仅记录日志）
	if export.Path != "" && export.Disk != "" {
		storage := facades.Storage().Disk(export.Disk)
		if err := storage.Delete(export.Path); err != nil {
			// 删除源文件失败只记录日志，不影响主流程
			errorlog.RecordHTTP(ctx, "export", "Failed to delete export source file", map[string]any{
				"error": err.Error(),
				"disk":  export.Disk,
				"path":  export.Path,
			}, "Delete export source file error: %v", err)
		}
	}

	if err := r.exportRecordService.Delete(id); err != nil {
		return response.ErrorWithLog(ctx, "export", err, map[string]any{
			"exportId": id,
		})
	}

	return response.Success(ctx)
}

// Download 下载导出文件
func (r *ExportController) Download(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	export, err := r.exportRecordService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	if export.Path == "" || export.Disk == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrFilePathRequired.Code)
	}

	// 获取存储驱动
	storage := facades.Storage().Disk(export.Disk)

	// 读取文件内容
	content, err := storage.Get(export.Path)
	if err != nil {
		return response.ErrorWithLog(ctx, "export", err, map[string]any{
			"disk": export.Disk,
			"path": export.Path,
		})
	}

	// 设置响应头
	filename := export.Filename
	if filename == "" {
		filename = export.Path
	}

	// 根据文件扩展名设置 Content-Type
	contentType := "application/octet-stream"
	switch export.Extension {
	case "csv":
		contentType = "text/csv; charset=utf-8"
	case "xlsx", "xls":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	// 设置响应头，使用链式调用确保顺序正确
	response := ctx.Response().
		Header("Content-Type", contentType).
		Header("Content-Length", fmt.Sprintf("%d", len(content))).
		Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename)).
		Header("Cache-Control", "no-cache, no-store, must-revalidate").
		Header("Pragma", "no-cache").
		Header("Expires", "0")

	return response.String(http.StatusOK, content)
}

type ExportBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除导出记录并删除源文件
func (r *ExportController) BatchDestroy(ctx http.Context) http.Response {
	var req ExportBatchDestroyRequest

	// 使用结构体绑定
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsError.Code)
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDsRequired.Code)
	}

	ids := req.IDs

	// 查询要删除的导出记录（用于删除源文件）
	exports, err := r.exportRecordService.GetByIDs(ids)
	if err == nil {
		// 尝试删除源文件（忽略失败，仅记录日志）
		for _, export := range exports {
			if export.Path != "" && export.Disk != "" {
				storage := facades.Storage().Disk(export.Disk)
				if err := storage.Delete(export.Path); err != nil {
					errorlog.RecordHTTP(ctx, "export", "Failed to delete export source file in batch delete", map[string]any{
						"error": err.Error(),
						"disk":  export.Disk,
						"path":  export.Path,
					}, "Delete export source file in batch delete error: %v", err)
				}
			}
		}
	}

	// 批量删除数据库记录
	if err := r.exportRecordService.BatchDelete(ids); err != nil {
		return response.ErrorWithLog(ctx, "export", err, map[string]any{
			"ids": ids,
		})
	}

	return response.Success(ctx)
}

// StreamExportProgress SSE 实时推送导出任务进度
// 监控导出任务的状态变化，实时推送进度信息
func (r *ExportController) StreamExportProgress(ctx http.Context) http.Response {
	// 获取参数
	exportID := helpers.GetUintRoute(ctx, "id")
	if exportID == 0 {
		// 尝试从查询参数获取
		exportID = helpers.GetUintQuery(ctx, "id", 0)
		if exportID == 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
		}
	}

	// 获取推送间隔（毫秒），默认 1 秒
	interval := 1000
	if intervalStr := ctx.Request().Query("interval", ""); intervalStr != "" {
		if parsed, err := time.ParseDuration(intervalStr + "ms"); err == nil {
			interval = max(int(parsed.Milliseconds()), 500)
			if interval > 5000 {
				interval = 5000
			}
		}
	}

	// 设置 SSE 响应头
	writer := ctx.Response().Writer()
	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 发送初始连接消息
	initMsg := map[string]any{
		"type":      "connected",
		"message":   "SSE连接已建立，开始监控导出任务进度",
		"export_id": exportID,
		"interval":  interval,
	}
	initData, _ := json.Marshal(initMsg)
	fmt.Fprintf(writer, "data: %s\n\n", string(initData))
	if flusher, ok := writer.(nethttp.Flusher); ok {
		flusher.Flush()
	}

	// 创建 ticker，定期检查导出任务状态
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	// 检测客户端断开连接
	clientGone := ctx.Request().Origin().Context().Done()

	// 记录上次的状态，避免重复推送
	lastStatus := uint8(255) // 使用一个不可能的值作为初始值
	lastPath := ""

	exportService := services.NewExportService(ctx)

	for {
		select {
		case <-clientGone:
			// 客户端断开连接
			return nil
		case <-ticker.C:
			// 查询导出任务
			export, err := r.exportRecordService.GetByID(exportID)
			if err != nil {
				// 导出任务不存在或已删除
				errorMsg := map[string]any{
					"type":    "error",
					"message": "导出任务不存在或已删除",
					"error":   err.Error(),
				}
				errorData, _ := json.Marshal(errorMsg)
				fmt.Fprintf(writer, "data: %s\n\n", string(errorData))
				if flusher, ok := writer.(nethttp.Flusher); ok {
					flusher.Flush()
				}
				// 继续监控，可能任务还在创建中
				continue
			}

			// 检查状态是否有变化
			if export.Status == lastStatus && export.Path == lastPath {
				// 状态和路径都没有变化，跳过本次推送
				// 但如果已完成，可以继续推送完成状态
				if export.Status == 1 && lastStatus == 1 {
					continue
				}
			}

			// 更新记录
			lastStatus = export.Status
			lastPath = export.Path

			// 构造进度消息
			message := map[string]any{
				"type":      "progress",
				"export_id": export.ID,
				"status":    export.Status,
				"timestamp": time.Now().Format(time.RFC3339),
			}

			// 根据状态设置不同的消息
			switch export.Status {
			case 1:
				// 导出成功
				message["type"] = "completed"
				message["message"] = "导出任务已完成"
				message["status_text"] = "成功"

				// 生成下载链接
				fileURL := ""
				if export.Path != "" {
					if export.Disk == "local" || export.Disk == "public" {
						fileURL = fmt.Sprintf("/api/admin/exports/%d/download", export.ID)
					} else {
						fileURL = exportService.GetExportURL(export.Path)
					}
				}

				message["file_url"] = fileURL
				message["filename"] = export.Filename
				message["size"] = export.Size
			case 0:
				// 导出失败
				message["type"] = "failed"
				message["message"] = "导出任务失败"
				message["status_text"] = "失败"
			default:
				// 处理中（Status 可能是其他值，或者我们不知道的状态）
				message["message"] = "导出任务处理中"
				message["status_text"] = "处理中"
			}

			// 如果文件路径已存在，说明正在生成文件
			if export.Path != "" && export.Status != 1 {
				message["message"] = "正在生成导出文件"
				// 可以尝试检查文件大小来判断进度（如果存储驱动支持）
				if export.Disk != "" {
					storage := facades.Storage().Disk(export.Disk)
					if size, err := storage.Size(export.Path); err == nil {
						message["file_size"] = size
						if export.Size > 0 {
							progress := float64(size) / float64(export.Size) * 100
							if progress > 100 {
								progress = 100
							}
							message["progress"] = progress
						}
					}
				}
			}

			messageData, err := json.Marshal(message)
			if err != nil {
				errorlog.RecordHTTP(ctx, "export", "Failed to marshal progress", map[string]any{
					"error": err.Error(),
				}, "Marshal progress error: %v", err)
				continue
			}

			// 发送 SSE 消息
			fmt.Fprintf(writer, "data: %s\n\n", string(messageData))

			// 刷新缓冲区
			if flusher, ok := writer.(nethttp.Flusher); ok {
				flusher.Flush()
			}

			// 如果已完成或失败，可以选择继续推送一段时间后关闭，或者保持连接
			// 这里选择继续推送，让前端决定何时关闭
		}
	}
}
