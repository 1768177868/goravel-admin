package admin

import (
	"fmt"
	"mime"
	"path/filepath"
	"strconv"
	"time"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/errorlog"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type AttachmentController struct {
	attachmentService services.AttachmentService
}

func NewAttachmentController() *AttachmentController {
	return &AttachmentController{}
}

// Index 附件列表
func (r *AttachmentController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := r.buildFilters(ctx)

	attachmentService := services.NewAttachmentService(ctx)
	attachments, total, err := attachmentService.GetList(filters, page, pageSize)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment", err)
	}

	type AttachmentWithURL struct {
		models.Attachment
		FileURL string `json:"file_url"`
	}

	var resultWithURL []AttachmentWithURL
	for _, a := range attachments {
		fileURL := attachmentService.GetFileURL(&a)
		resultWithURL = append(resultWithURL, AttachmentWithURL{
			Attachment: a,
			FileURL:    fileURL,
		})
	}

	return response.Success(ctx, http.Json{
		"list":      resultWithURL,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// buildFilters 构建查询过滤器
func (r *AttachmentController) buildFilters(ctx http.Context) services.AttachmentFilters {
	adminID := ctx.Request().Query("admin_id", "")
	filename := ctx.Request().Query("filename", "")
	displayName := ctx.Request().Query("display_name", "")
	fileType := ctx.Request().Query("file_type", "")
	extension := ctx.Request().Query("extension", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.AttachmentFilters{
		AdminID:     adminID,
		Filename:    filename,
		DisplayName: displayName,
		FileType:    fileType,
		Extension:   extension,
		StartTime:   startTime,
		EndTime:     endTime,
		OrderBy:     orderBy,
	}
}

// Upload 普通文件上传（小文件）
func (r *AttachmentController) Upload(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrFileRequired.Code)
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

	// 生成预览URL（如果 GetFileURL 返回的是相对路径，且需要通过 preview 接口访问）
	// 注意：GetFileURL 通常返回的是 storage/public 的直接访问地址（如果是 public 驱动），或者是 api/admin/attachments/{id}/preview
	// 这里我们显式返回一个可直接访问的预览地址，优先使用 GetFileURL 的结果
	// 如果 GetFileURL 返回的是相对路径（如 /storage/xxx），前端会自动拼接域名

	// 为了确保富文本编辑器能直接显示，我们再返回一个 absolute_url 字段
	// 如果 fileURL 已经是完整的 http 地址（例如 OSS），则直接使用
	// 如果是相对地址，则拼接当前服务的域名（在前端做，后端这里只返回 path）

	// 但用户的需求是返回一个“可预览的完整地址”。
	// 附件管理列表能显示是因为前端拼接了域名。
	// 这里我们直接返回 file_url，它通常已经是正确的路径（相对或绝对）。
	// 我们添加一个 preview_url 字段，明确指向 preview 接口（如果是本地存储且未公开 storage）
	// 注意：这里需要确保返回的是完整的 URL，包含 api/admin 前缀
	previewURL := fmt.Sprintf("/api/admin/public/images/%d", attachment.ID)
	// 如果使用的是云存储，GetFileURL 返回的通常是 HTTP 链接，直接用那个更好
	if attachment.Disk != "local" && attachment.Disk != "public" {
		previewURL = fileURL
	}

	// 统一返回 file_url 为 previewURL，与附件列表保持一致
	fileURL = previewURL

	return response.Success(ctx, "upload_success", http.Json{
		"id":          attachment.ID,
		"filename":    attachment.Filename,
		"size":        attachment.Size,
		"mime_type":   attachment.MimeType,
		"file_type":   attachment.FileType,
		"file_url":    fileURL,
		"preview_url": previewURL, // 新增字段
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
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrFilenameRequired.Code)
		}

		totalSizeStr := ctx.Request().Input("total_size", "0")
		totalSize, err := strconv.ParseInt(totalSizeStr, 10, 64)
		if err != nil {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalSize.Code)
		}
		if totalSize <= 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalSize.Code)
		}

		chunkSizeStr := ctx.Request().Input("chunk_size", "0")
		chunkSize, err := strconv.ParseInt(chunkSizeStr, 10, 64)
		if err != nil {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidChunkSize.Code)
		}
		if chunkSize <= 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidChunkSize.Code)
		}

		totalChunksStr := ctx.Request().Input("total_chunks", "0")
		totalChunks, err := strconv.Atoi(totalChunksStr)
		if err != nil {
			// 尝试作为浮点数解析（以防传入 "5.0" 格式）
			if floatVal, floatErr := strconv.ParseFloat(totalChunksStr, 64); floatErr == nil {
				totalChunks = int(floatVal)
			} else {
				return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalChunks.Code)
			}
		}
		if totalChunks <= 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalChunks.Code)
		}

		// 验证分片数量计算的合理性
		expectedChunks := int((totalSize + chunkSize - 1) / chunkSize) // 向上取整
		if totalChunks != expectedChunks {
			// 不返回错误，使用客户端提供的值（可能是由于浮点数计算差异）
		}

		chunkID, err := attachmentService.InitChunkUpload(filename, totalSize, chunkSize, totalChunks)
		if err != nil {
			// 使用业务错误类型，直接提取错误码
			if businessErr, ok := apperrors.GetBusinessError(err); ok {
				return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
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
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrChunkIDRequired.Code)
		}

		chunkIndex, err := strconv.Atoi(ctx.Request().Input("chunk_index", "-1"))
		if err != nil || chunkIndex < 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidChunkIndex.Code)
		}

		file, err := ctx.Request().File("chunk")
		if err != nil {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrChunkFileRequired.Code)
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
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrChunkIDRequired.Code)
		}

		filename := ctx.Request().Input("filename", "")
		if filename == "" {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrFilenameRequired.Code)
		}

		totalChunksStr := ctx.Request().Input("total_chunks", "0")
		totalChunks, err := strconv.Atoi(totalChunksStr)
		if err != nil {
			// 尝试作为浮点数解析（以防传入 "5.0" 格式）
			if floatVal, floatErr := strconv.ParseFloat(totalChunksStr, 64); floatErr == nil {
				totalChunks = int(floatVal)
			} else {
				return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalChunks.Code)
			}
		}
		if totalChunks <= 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalChunks.Code)
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

		// 合并成功后，记录日志（仅 Debug 模式）
		facades.Log().Debugf("Successfully merged chunks for chunkID %s, filename: %s, total_chunks: %d", chunkID, filename, totalChunks)

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
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrChunkIDRequired.Code)
		}

		totalChunks, err := strconv.Atoi(ctx.Request().Query("total_chunks", "0"))
		if totalChunks == 0 {
			totalChunks, err = strconv.Atoi(ctx.Request().Input("total_chunks", "0"))
		}
		if err != nil || totalChunks <= 0 {
			return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidTotalChunks.Code)
		}

		progress, err := attachmentService.GetChunkProgress(chunkID, totalChunks)
		if err != nil {
			return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
				"chunk_id": chunkID,
			})
		}

		return response.Success(ctx, progress)

	default:
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrInvalidAction.Code)
	}
}

// Download 下载附件文件
func (r *AttachmentController) Download(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	attachmentService := services.NewAttachmentService(ctx)
	attachment, err := attachmentService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	if attachment.Path == "" || attachment.Disk == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrFilePathRequired.Code)
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
		directURL := attachmentService.GetFileURL(attachment)
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
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	attachmentService := services.NewAttachmentService(ctx)
	attachment, err := attachmentService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	if attachment.Path == "" || attachment.Disk == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrFilePathRequired.Code)
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
		directURL := attachmentService.GetFileURL(attachment)
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
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	attachmentService := services.NewAttachmentService(ctx)
	attachment, err := attachmentService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	if err := attachmentService.DeleteFile(attachment); err != nil {
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
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsError.Code)
	}

	if len(req.IDs) == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDsRequired.Code)
	}

	ids := req.IDs

	// 查询要删除的附件
	attachmentService := services.NewAttachmentService(ctx)
	attachments, err := attachmentService.GetByIDs(ids)
	if err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"ids": ids,
		})
	}

	// 删除文件和记录
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
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrIDRequired.Code)
	}

	attachmentService := services.NewAttachmentService(ctx)
	displayName := ctx.Request().Input("display_name", "")

	if err := attachmentService.UpdateDisplayName(id, displayName); err != nil {
		return response.ErrorWithLog(ctx, "attachment", err, map[string]any{
			"attachId": id,
		})
	}

	// 重新获取更新后的附件
	attachment, err := attachmentService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	return response.Success(ctx, http.Json{
		"attachment": attachment,
	})
}
