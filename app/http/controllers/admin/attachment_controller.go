package admin

import (
	"fmt"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"
)

type AttachmentController struct {
}

func NewAttachmentController() *AttachmentController {
	return &AttachmentController{}
}

// Index 附件列表
func (r *AttachmentController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var attachments []models.Attachment
	attachmentService := services.NewAttachmentService(ctx)
	type AttachmentWithURL struct {
		models.Attachment
		FileURL string `json:"file_url"`
	}

	return response.PaginateQuery(ctx, query, &attachments, &response.PaginateQueryOptions{
		WithRelations: []string{"Admin"},
		ErrorModule:   "attachment",
		Transform: func(data any) any {
			attachments := data.(*[]models.Attachment)
			var resultWithURL []AttachmentWithURL
			for _, a := range *attachments {
				fileURL := attachmentService.GetFileURL(&a)
				resultWithURL = append(resultWithURL, AttachmentWithURL{
					Attachment: a,
					FileURL:    fileURL,
				})
			}
			return resultWithURL
		},
	})
}

// buildQuery 构建附件查询
func (r *AttachmentController) buildQuery(ctx http.Context) orm.Query {
	query := facades.Orm().Query().Model(&models.Attachment{})

	adminID := ctx.Request().Query("admin_id", "")
	filename := ctx.Request().Query("filename", "")
	displayName := ctx.Request().Query("display_name", "")
	fileType := ctx.Request().Query("file_type", "")
	extension := ctx.Request().Query("extension", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	if adminID != "" {
		query = query.Where("admin_id", adminID)
	}
	if filename != "" {
		query = query.Where("filename LIKE ?", "%"+filename+"%")
	}
	if displayName != "" {
		query = query.Where("display_name LIKE ?", "%"+displayName+"%")
	}
	if fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	if extension != "" {
		query = query.Where("extension = ?", extension)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 id desc
	query = helpers.ApplySort(query, orderBy, "id:desc")

	return query
}

// Upload 普通文件上传（小文件）
func (r *AttachmentController) Upload(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "file_required")
	}

	filename := file.GetClientOriginalName()
	if filename == "" {
		filename = "uploaded_file"
	}

	// 读取文件内容：先将文件保存到临时位置，然后读取
	storage := facades.Storage().Disk("local")

	// 保存文件到临时位置，PutFile 返回保存后的路径
	savedPath, err := storage.PutFile("", file)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"filename": filename,
		})
	}

	// 读取文件内容
	fileDataStr, err := storage.Get(savedPath)
	if err != nil {
		// 清理临时文件
		_ = storage.Delete(savedPath)
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"filename": filename,
		})
	}

	// 清理临时文件
	_ = storage.Delete(savedPath)

	// 转换为字节数组
	fileData := []byte(fileDataStr)

	// 获取MIME类型：直接根据文件扩展名推断（multipart/form-data 的 Content-Type 不是文件本身的 MIME 类型）
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	attachmentService := services.NewAttachmentService(ctx)
	attachment, err := attachmentService.UploadFile(fileData, filename, mimeType)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"filename": filename,
		})
	}

	fileURL := attachmentService.GetFileURL(attachment)

	return response.Success(ctx, "upload_success", http.Json{
		"id":        attachment.ID,
		"filename":  attachment.Filename,
		"size":      attachment.Size,
		"mime_type": attachment.MimeType,
		"file_type": attachment.FileType,
		"file_url":  fileURL,
	})
}

// ChunkUpload 大文件分片上传统一接口
// 通过 action 参数区分不同操作：init（初始化）、upload（上传分片）、merge（合并分片）、progress（获取进度）
func (r *AttachmentController) ChunkUpload(ctx http.Context) http.Response {
	action := ctx.Request().Input("action", "")
	if action == "" {
		// 兼容 GET 请求获取进度
		action = ctx.Request().Query("action", "progress")
	}

	attachmentService := services.NewAttachmentService(ctx)

	switch action {
	case "init":
		// 初始化分片上传
		filename := ctx.Request().Input("filename", "")
		if filename == "" {
			return response.Error(ctx, http.StatusBadRequest, "filename_required")
		}

		totalSizeStr := ctx.Request().Input("total_size", "0")
		totalSize, err := strconv.ParseInt(totalSizeStr, 10, 64)
		if err != nil {
			return response.Error(ctx, http.StatusBadRequest, "invalid_total_size")
		}
		if totalSize <= 0 {
			return response.Error(ctx, http.StatusBadRequest, "invalid_total_size")
		}

		chunkSizeStr := ctx.Request().Input("chunk_size", "0")
		chunkSize, err := strconv.ParseInt(chunkSizeStr, 10, 64)
		if err != nil {
			return response.Error(ctx, http.StatusBadRequest, "invalid_chunk_size")
		}
		if chunkSize <= 0 {
			return response.Error(ctx, http.StatusBadRequest, "invalid_chunk_size")
		}

		totalChunksStr := ctx.Request().Input("total_chunks", "0")
		totalChunks, err := strconv.Atoi(totalChunksStr)
		if err != nil {
			// 尝试作为浮点数解析（以防传入 "5.0" 格式）
			if floatVal, floatErr := strconv.ParseFloat(totalChunksStr, 64); floatErr == nil {
				totalChunks = int(floatVal)
			} else {
				return response.Error(ctx, http.StatusBadRequest, "invalid_total_chunks")
			}
		}
		if totalChunks <= 0 {
			return response.Error(ctx, http.StatusBadRequest, "invalid_total_chunks")
		}

		// 验证分片数量计算的合理性
		expectedChunks := int((totalSize + chunkSize - 1) / chunkSize) // 向上取整
		if totalChunks != expectedChunks {
			// 不返回错误，使用客户端提供的值（可能是由于浮点数计算差异）
		}

		chunkID, err := attachmentService.InitChunkUpload(filename, totalSize, chunkSize, totalChunks)
		if err != nil {
			// 检查是否是存储驱动不支持的错误
			if strings.Contains(err.Error(), "大文件分片上传仅支持本地存储") {
				return response.Error(ctx, http.StatusBadRequest, "chunk_upload_only_local_storage")
			}

			// 返回详细的错误信息
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"filename":     filename,
				"total_size":   totalSize,
				"chunk_size":   chunkSize,
				"total_chunks": totalChunks,
			})
		}

		return response.Success(ctx, "init_chunk_upload_success", http.Json{
			"chunk_id": chunkID,
		})

	case "upload":
		// 上传分片
		chunkID := ctx.Request().Input("chunk_id", "")
		if chunkID == "" {
			return response.Error(ctx, http.StatusBadRequest, "chunk_id_required")
		}

		chunkIndex, err := strconv.Atoi(ctx.Request().Input("chunk_index", "-1"))
		if err != nil || chunkIndex < 0 {
			return response.Error(ctx, http.StatusBadRequest, "invalid_chunk_index")
		}

		file, err := ctx.Request().File("chunk")
		if err != nil {
			return response.Error(ctx, http.StatusBadRequest, "chunk_file_required")
		}

		// 读取分片数据：先将文件保存到临时位置，然后读取
		storage := facades.Storage().Disk("local")

		// 保存文件到临时位置
		savedPath, err := storage.PutFile("", file)
		if err != nil {
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"chunk_id":    chunkID,
				"chunk_index": chunkIndex,
			})
		}

		// 读取文件内容
		chunkDataStr, err := storage.Get(savedPath)
		if err != nil {
			// 清理临时文件
			_ = storage.Delete(savedPath)
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"chunk_id":    chunkID,
				"chunk_index": chunkIndex,
			})
		}

		// 清理临时文件
		_ = storage.Delete(savedPath)

		// 转换为字节数组
		chunkData := []byte(chunkDataStr)

		if err := attachmentService.UploadChunk(chunkID, chunkIndex, chunkData); err != nil {
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"chunk_id":    chunkID,
				"chunk_index": chunkIndex,
			})
		}

		return response.Success(ctx, "upload_chunk_success")

	case "merge":
		// 合并分片
		chunkID := ctx.Request().Input("chunk_id", "")
		if chunkID == "" {
			return response.Error(ctx, http.StatusBadRequest, "chunk_id_required")
		}

		filename := ctx.Request().Input("filename", "")
		if filename == "" {
			return response.Error(ctx, http.StatusBadRequest, "filename_required")
		}

		totalChunksStr := ctx.Request().Input("total_chunks", "0")
		totalChunks, err := strconv.Atoi(totalChunksStr)
		if err != nil {
			// 尝试作为浮点数解析（以防传入 "5.0" 格式）
			if floatVal, floatErr := strconv.ParseFloat(totalChunksStr, 64); floatErr == nil {
				totalChunks = int(floatVal)
			} else {
				return response.Error(ctx, http.StatusBadRequest, "invalid_total_chunks")
			}
		}
		if totalChunks <= 0 {
			return response.Error(ctx, http.StatusBadRequest, "invalid_total_chunks")
		}

		// 获取MIME类型：直接根据文件扩展名推断（前端传递的 mime_type 可能不准确）
		ext := filepath.Ext(filename)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		attachment, err := attachmentService.MergeChunks(chunkID, filename, mimeType, totalChunks)
		if err != nil {
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"chunk_id":     chunkID,
				"filename":     filename,
				"total_chunks": totalChunks,
			})
		}

		// 合并成功后，记录日志（用于调试）
		facades.Log().Infof("Successfully merged chunks for chunkID %s, filename: %s, total_chunks: %d", chunkID, filename, totalChunks)

		fileURL := attachmentService.GetFileURL(attachment)

		return response.Success(ctx, "merge_chunks_success", http.Json{
			"id":        attachment.ID,
			"filename":  attachment.Filename,
			"size":      attachment.Size,
			"mime_type": attachment.MimeType,
			"file_type": attachment.FileType,
			"file_url":  fileURL,
		})

	case "progress":
		// 获取分片上传进度
		chunkID := ctx.Request().Query("chunk_id", "")
		if chunkID == "" {
			chunkID = ctx.Request().Input("chunk_id", "")
		}
		if chunkID == "" {
			return response.Error(ctx, http.StatusBadRequest, "chunk_id_required")
		}

		totalChunks, err := strconv.Atoi(ctx.Request().Query("total_chunks", "0"))
		if totalChunks == 0 {
			totalChunks, err = strconv.Atoi(ctx.Request().Input("total_chunks", "0"))
		}
		if err != nil || totalChunks <= 0 {
			return response.Error(ctx, http.StatusBadRequest, "invalid_total_chunks")
		}

		progress, err := attachmentService.GetChunkProgress(chunkID, totalChunks)
		if err != nil {
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"chunk_id": chunkID,
			})
		}

		return response.Success(ctx, progress)

	default:
		return response.Error(ctx, http.StatusBadRequest, "invalid_action")
	}
}

// Download 下载附件文件
func (r *AttachmentController) Download(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var attachment models.Attachment
	if err := facades.Orm().Query().Where("id", id).First(&attachment); err != nil {
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	if attachment.Path == "" || attachment.Disk == "" {
		return response.Error(ctx, http.StatusBadRequest, "file_path_required")
	}

	// 对于云存储，尝试生成临时URL并重定向，避免通过服务器中转
	if attachment.Disk != "local" && attachment.Disk != "public" {
		storage := facades.Storage().Disk(attachment.Disk)

		// 尝试生成临时URL（24小时有效）
		if url, err := storage.TemporaryUrl(attachment.Path, time.Now().Add(24*time.Hour)); err == nil {
			return ctx.Response().Redirect(http.StatusFound, url)
		}

		// 如果生成临时URL失败，尝试从配置获取基础URL
		attachmentService := services.NewAttachmentService(ctx)
		directURL := attachmentService.GetFileURL(&attachment)
		if directURL != "" && directURL != fmt.Sprintf("/api/admin/attachments/%d/preview", attachment.ID) {
			return ctx.Response().Redirect(http.StatusFound, directURL)
		}
		// 如果都失败，继续使用服务器中转方式
	}

	// 对于本地存储或临时URL生成失败的情况，使用服务器中转
	storage := facades.Storage().Disk(attachment.Disk)

	// 读取文件内容
	content, err := storage.Get(attachment.Path)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"disk": attachment.Disk,
			"path": attachment.Path,
		})
	}

	// 设置响应头
	filename := attachment.Filename
	if filename == "" {
		filename = attachment.Path
	}

	// 根据MIME类型设置 Content-Type
	contentType := attachment.MimeType
	if contentType == "" {
		contentType = "application/octet-stream"
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

// Preview 预览文件（图片、视频、文档）
func (r *AttachmentController) Preview(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var attachment models.Attachment
	if err := facades.Orm().Query().Where("id", id).First(&attachment); err != nil {
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	if attachment.Path == "" || attachment.Disk == "" {
		return response.Error(ctx, http.StatusBadRequest, "file_path_required")
	}

	// 对于云存储，尝试生成临时URL并重定向，避免通过服务器中转
	// 这样可以减少服务器带宽和内存占用，提高性能
	if attachment.Disk != "local" && attachment.Disk != "public" {
		storage := facades.Storage().Disk(attachment.Disk)

		// 尝试生成临时URL（24小时有效）
		if url, err := storage.TemporaryUrl(attachment.Path, time.Now().Add(24*time.Hour)); err == nil {
			return ctx.Response().Redirect(http.StatusFound, url)
		}

		// 如果生成临时URL失败，尝试从配置获取基础URL
		attachmentService := services.NewAttachmentService(ctx)
		directURL := attachmentService.GetFileURL(&attachment)
		if directURL != "" && directURL != fmt.Sprintf("/api/admin/attachments/%d/preview", attachment.ID) {
			return ctx.Response().Redirect(http.StatusFound, directURL)
		}
		// 如果都失败，继续使用服务器中转方式
	}

	// 对于本地存储或临时URL生成失败的情况，使用服务器中转
	storage := facades.Storage().Disk(attachment.Disk)

	// 读取文件内容
	content, err := storage.Get(attachment.Path)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"disk": attachment.Disk,
			"path": attachment.Path,
		})
	}

	// 设置响应头
	mimeType := attachment.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 设置响应头
	response := ctx.Response().
		Header("Content-Type", mimeType).
		Header("Content-Length", fmt.Sprintf("%d", len(content))).
		Header("Cache-Control", "public, max-age=3600")

	// 对于图片和视频，支持范围请求（Range request）
	if attachment.FileType == "image" || attachment.FileType == "video" {
		response = response.Header("Accept-Ranges", "bytes")
	}

	return response.String(http.StatusOK, content)
}

// Destroy 删除附件
func (r *AttachmentController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var attachment models.Attachment
	if err := facades.Orm().Query().Where("id", id).First(&attachment); err != nil {
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	attachmentService := services.NewAttachmentService(ctx)
	if err := attachmentService.DeleteFile(&attachment); err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"attachId": attachment.ID,
		})
	}

	return response.Success(ctx)
}

type AttachmentBatchDestroyRequest struct {
	IDs []uint `json:"ids"`
}

// BatchDestroy 批量删除附件
func (r *AttachmentController) BatchDestroy(ctx http.Context) http.Response {
	var req AttachmentBatchDestroyRequest

	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "params_error")
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "ids_required")
	}

	ids := req.IDs
	idsAny := helpers.ConvertUintSliceToAny(ids)

	// 查询要删除的附件
	var attachments []models.Attachment
	if err := facades.Orm().Query().WhereIn("id", idsAny).Get(&attachments); err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"ids": ids,
		})
	}

	// 删除文件和记录
	attachmentService := services.NewAttachmentService(ctx)
	for _, attachment := range attachments {
		if err := attachmentService.DeleteFile(&attachment); err != nil {
			// 批量删除中单个文件删除失败只记录日志，不影响主流程
			errorlog.RecordHTTP(ctx, "attachment", "Failed to delete attachment in batch delete", map[string]any{
				"error":    err.Error(),
				"attachId": attachment.ID,
			}, "Delete attachment in batch delete error: %v", err)
		}
	}

	return response.Success(ctx)
}

// UpdateDisplayName 更新显示名称
func (r *AttachmentController) UpdateDisplayName(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, "id_required")
	}

	var attachment models.Attachment
	if err := facades.Orm().Query().Where("id", id).First(&attachment); err != nil {
		return response.Error(ctx, http.StatusNotFound, "record_not_found")
	}

	displayName := ctx.Request().Input("display_name", "")
	attachment.DisplayName = displayName

	if err := facades.Orm().Query().Save(&attachment); err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"attachId": id,
		})
	}

	return response.Success(ctx, http.Json{
		"attachment": attachment,
	})
}
