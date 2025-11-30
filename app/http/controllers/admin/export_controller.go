package admin

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"
)

type ExportController struct {
}

func NewExportController() *ExportController {
	return &ExportController{}
}

// Index 导出记录列表
func (r *ExportController) Index(ctx http.Context) http.Response {
	page, pageSize := helpers.ValidatePagination(
		helpers.GetIntQuery(ctx, "page", 1),
		helpers.GetIntQuery(ctx, "page_size", 10),
	)

	query := r.buildQuery(ctx)

	total, err := query.Count()
	if err != nil {
		errorlog.RecordHTTP(ctx, "export", "Failed to count exports", map[string]any{
			"error": err.Error(),
		}, "Count exports error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	orderBy := ctx.Request().Query("order_by", "id:desc")
	query = helpers.ApplySort(query, orderBy, "id:desc")

	offset := (page - 1) * pageSize

	var exports []models.Export
	if err := query.With("Admin").Offset(offset).Limit(pageSize).Get(&exports); err != nil {
		errorlog.RecordHTTP(ctx, "export", "Failed to query exports", map[string]any{
			"error": err.Error(),
		}, "Query exports error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
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

	return response.Paginate(ctx, "get_success", resultWithURL, total, page, pageSize)
}

// buildQuery 构建导出记录查询
func (r *ExportController) buildQuery(ctx http.Context) orm.Query {
	query := facades.Orm().Query().Model(&models.Export{})

	adminID := ctx.Request().Query("admin_id", "")
	filename := ctx.Request().Query("filename", "")
	disk := ctx.Request().Query("disk", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	if adminID != "" {
		query = query.Where("admin_id", adminID)
	}
	if filename != "" {
		query = query.Where("filename LIKE ?", "%"+filename+"%")
	}
	if disk != "" {
		query = query.Where("disk = ?", disk)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	return query
}

// Destroy 删除导出记录并删除源文件
func (r *ExportController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var export models.Export
	if err := facades.Orm().Query().Where("id", id).First(&export); err != nil {
		errorlog.RecordHTTP(ctx, "export", "Export record not found for delete", map[string]any{
			"error":    err.Error(),
			"exportId": id,
		}, "Export record not found: %v", err)
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	// 尝试删除源文件（忽略失败，仅记录日志）
	if export.Path != "" && export.Disk != "" {
		storage := facades.Storage().Disk(export.Disk)
		if err := storage.Delete(export.Path); err != nil {
			errorlog.RecordHTTP(ctx, "export", "Failed to delete export source file", map[string]any{
				"error": err.Error(),
				"disk":  export.Disk,
				"path":  export.Path,
			}, "Delete export source file error: %v", err)
		}
	}

	if _, err := facades.Orm().Query().Delete(&export); err != nil {
		errorlog.RecordHTTP(ctx, "export", "Failed to delete export record", map[string]any{
			"error":    err.Error(),
			"exportId": export.ID,
		}, "Delete export record error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// Download 下载导出文件
func (r *ExportController) Download(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var export models.Export
	if err := facades.Orm().Query().Where("id", id).First(&export); err != nil {
		errorlog.RecordHTTP(ctx, "export", "Export record not found for download", map[string]any{
			"error":    err.Error(),
			"exportId": id,
		}, "Export record not found: %v", err)
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	if export.Path == "" || export.Disk == "" {
		return response.Error(ctx, http.StatusBadRequest, "file_path_required")
	}

	// 获取存储驱动
	storage := facades.Storage().Disk(export.Disk)

	// 读取文件内容
	content, err := storage.Get(export.Path)
	if err != nil {
		errorlog.RecordHTTP(ctx, "export", "Failed to read export file", map[string]any{
			"error": err.Error(),
			"disk":  export.Disk,
			"path":  export.Path,
		}, "Read export file error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "file_read_failed")
	}

	// 设置响应头
	filename := export.Filename
	if filename == "" {
		filename = export.Path
	}

	// 根据文件扩展名设置 Content-Type
	contentType := "application/octet-stream"
	if export.Extension == "csv" {
		contentType = "text/csv; charset=utf-8"
	} else if export.Extension == "xlsx" || export.Extension == "xls" {
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
		errorlog.RecordHTTP(ctx, "export", "Failed to bind batch delete request", map[string]any{
			"error": err.Error(),
		}, "Bind batch delete request error: %v", err)
		return response.Error(ctx, http.StatusBadRequest, "params_error")
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "ids_required")
	}

	ids := req.IDs
	idsAny := helpers.ConvertUintSliceToAny(ids)

	// 查询要删除的导出记录
	var exports []models.Export
	if err := facades.Orm().Query().WhereIn("id", idsAny).Get(&exports); err != nil {
		errorlog.RecordHTTP(ctx, "export", "Failed to query exports for batch delete", map[string]any{
			"error": err.Error(),
			"ids":   ids,
		}, "Query exports for batch delete error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

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

	// 批量删除数据库记录
	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.Export{}); err != nil {
		errorlog.RecordHTTP(ctx, "export", "Failed to batch delete export records", map[string]any{
			"error": err.Error(),
			"ids":   ids,
		}, "Batch delete export records error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
