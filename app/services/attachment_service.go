package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/utils"
)

type AttachmentService interface {
	// InitChunkUpload 初始化分片上传（不再使用服务端缓存）
	InitChunkUpload(filename string, totalSize int64, chunkSize int64, totalChunks int) (string, error)

	// UploadChunk 上传分片（不再使用服务端缓存）
	UploadChunk(chunkID string, chunkIndex int, chunkData []byte) error

	// MergeChunks 合并分片（不再使用服务端缓存，需要传入 totalChunks）
	MergeChunks(chunkID string, filename string, mimeType string, totalChunks int) (*models.Attachment, error)

	// GetChunkProgress 获取分片上传进度（不再使用服务端缓存，需要传入 totalChunks）
	GetChunkProgress(chunkID string, totalChunks int) (map[string]any, error)

	// UploadFile 普通文件上传（小文件）
	UploadFile(fileData []byte, filename string, mimeType string) (*models.Attachment, error)

	// GetFileURL 获取文件访问URL
	GetFileURL(attachment *models.Attachment) string

	// GetFileType 根据MIME类型判断文件类型
	GetFileType(mimeType string) string

	// DeleteFile 删除文件
	DeleteFile(attachment *models.Attachment) error
}

type AttachmentServiceImpl struct {
	ctx              http.Context
	disk             string
	systemLogService SystemLogService
}

func NewAttachmentService(ctx http.Context) AttachmentService {
	// 从数据库读取文件存储配置
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = "local"
	}

	return &AttachmentServiceImpl{
		ctx:              ctx,
		disk:             disk,
		systemLogService: NewSystemLogService(),
	}
}

// InitChunkUpload 初始化分片上传
// 注意：分片信息由客户端缓存，服务端只生成 chunkID
func (s *AttachmentServiceImpl) InitChunkUpload(filename string, totalSize int64, chunkSize int64, totalChunks int) (string, error) {
	// 检查存储驱动：大文件分片上传仅支持本地存储
	cloudStorageDrivers := []string{"s3", "oss", "cos", "minio", "qiniu"}
	for _, driver := range cloudStorageDrivers {
		if s.disk == driver {
			return "", fmt.Errorf("大文件分片上传仅支持本地存储，当前存储驱动为: %s", driver)
		}
	}

	// 生成唯一的分片ID
	hash := md5.Sum(fmt.Appendf(nil, "%s_%d_%d", filename, totalSize, time.Now().UnixNano()))
	chunkID := hex.EncodeToString(hash[:])

	// 不再使用服务端缓存，分片信息由客户端管理
	return chunkID, nil
}

// UploadChunk 上传分片
// 注意：不再使用服务端缓存，直接保存分片文件
func (s *AttachmentServiceImpl) UploadChunk(chunkID string, chunkIndex int, chunkData []byte) error {
	if chunkIndex < 0 {
		return fmt.Errorf("分片索引无效")
	}

	// 保存分片到临时目录
	storage := facades.Storage().Disk(s.disk)
	chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, chunkIndex)

	if err := storage.Put(chunkPath, string(chunkData)); err != nil {
		return fmt.Errorf("保存分片失败: %w", err)
	}

	return nil
}

// MergeChunks 合并分片
// 注意：不再使用服务端缓存，通过检查实际文件系统来验证分片
func (s *AttachmentServiceImpl) MergeChunks(chunkID string, filename string, mimeType string, totalChunks int) (*models.Attachment, error) {
	storage := facades.Storage().Disk(s.disk)

	// 检查所有分片文件是否存在
	indices := make([]int, totalChunks)
	for i := range indices {
		chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, i)
		if !storage.Exists(chunkPath) {
			return nil, fmt.Errorf("分片 %d 不存在", i)
		}
	}

	// 生成最终文件路径
	ext := filepath.Ext(filename)
	if ext == "" {
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[0]
		}
	}

	// 生成唯一文件名
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%d", filename, time.Now().UnixNano())))
	uniqueName := hex.EncodeToString(hash[:]) + ext
	datePath := time.Now().Format("2006/01/02")
	finalPath := fmt.Sprintf("attachments/%s/%s", datePath, uniqueName)

	// 合并分片（必须按顺序读取所有分片）
	var mergedData []byte
	var missingChunks []int // 记录缺失的分片索引
	for i := range make([]int, totalChunks) {
		chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, i)
		if !storage.Exists(chunkPath) {
			missingChunks = append(missingChunks, i)
			continue
		}
		chunkContent, err := storage.Get(chunkPath)
		if err != nil {
			// 读取失败，记录到系统日志
			if s.ctx != nil {
				_ = s.systemLogService.RecordHTTP(s.ctx, "warning", "attachment", fmt.Sprintf("Failed to read chunk %d for chunkID %s", i, chunkID), map[string]any{
					"chunk_index": i,
					"chunk_id":    chunkID,
					"error":       err.Error(),
				})
			}
			missingChunks = append(missingChunks, i)
			continue
		}
		mergedData = append(mergedData, []byte(chunkContent)...)
	}

	// 检查是否有缺失的分片
	if len(missingChunks) > 0 {
		return nil, fmt.Errorf("分片缺失: %v (共 %d 个分片缺失)", missingChunks, len(missingChunks))
	}

	// 检查是否有数据被合并
	if len(mergedData) == 0 {
		return nil, fmt.Errorf("没有可合并的分片数据")
	}

	// 保存合并后的文件
	if err := storage.Put(finalPath, string(mergedData)); err != nil {
		return nil, fmt.Errorf("保存合并文件失败: %w", err)
	}

	// 获取文件大小
	fileSize := int64(len(mergedData))
	if size, err := storage.Size(finalPath); err == nil {
		fileSize = size
	}

	// 创建附件记录
	adminID := uint(0)
	if s.ctx != nil {
		if id, err := helpers.GetAdminIDFromContext(s.ctx); err == nil {
			adminID = id
		}
	}

	fileType := s.GetFileType(mimeType)
	attachment := &models.Attachment{
		AdminID:   adminID,
		Disk:      s.disk,
		Path:      finalPath,
		Filename:  filename,
		Extension: strings.TrimPrefix(ext, "."),
		MimeType:  mimeType,
		Size:      fileSize,
		Status:    1,
		FileType:  fileType,
		ChunkID:   chunkID,
	}

	if err := facades.Orm().Query().Create(attachment); err != nil {
		// 如果创建记录失败，删除已上传的文件
		_ = storage.Delete(finalPath)
		return nil, fmt.Errorf("创建附件记录失败: %w", err)
	}

	// 合并成功后才清理分片文件（确保数据已保存到数据库）
	// 清理所有分片文件（包括可能缺失的）
	cleanupSuccess := true
	cleanupCount := 0
	for i := range make([]int, totalChunks) {
		chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, i)
		if storage.Exists(chunkPath) {
			if err := storage.Delete(chunkPath); err != nil {
				// 记录删除失败到系统日志，但不影响整体流程
				if s.ctx != nil {
					_ = s.systemLogService.RecordHTTP(s.ctx, "warning", "attachment", fmt.Sprintf("Failed to delete chunk file %s", chunkPath), map[string]any{
						"chunk_path": chunkPath,
						"chunk_id":   chunkID,
						"error":      err.Error(),
					})
				}
				cleanupSuccess = false
			} else {
				cleanupCount++
			}
		}
	}

	if !cleanupSuccess && s.ctx != nil {
		// 记录部分分片删除失败的警告到系统日志
		_ = s.systemLogService.RecordHTTP(s.ctx, "warning", "attachment", fmt.Sprintf("Some chunk files failed to delete for chunkID %s", chunkID), map[string]any{
			"chunk_id":      chunkID,
			"cleaned_count": cleanupCount,
			"total_chunks":  totalChunks,
		})
	}

	return attachment, nil
}

// GetChunkProgress 获取分片上传进度
// 注意：不再使用服务端缓存，通过检查实际文件系统来获取进度
// 优化：如果分片数量很大，可以考虑限制返回的索引数量或使用并发检查
func (s *AttachmentServiceImpl) GetChunkProgress(chunkID string, totalChunks int) (map[string]any, error) {
	storage := facades.Storage().Disk(s.disk)

	uploadedCount := 0
	uploadedIndices := []int{} // 已上传的分片索引数组

	// 优化：如果分片数量超过1000，只返回前100个已上传的索引，减少响应大小
	maxIndices := 100
	if totalChunks > 1000 {
		maxIndices = 100
	}

	// 检查每个分片文件是否存在
	indices := make([]int, totalChunks)
	for i := range indices {
		chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, i)
		if storage.Exists(chunkPath) {
			uploadedCount++
			// 只记录前 maxIndices 个索引，减少响应大小
			if len(uploadedIndices) < maxIndices {
				uploadedIndices = append(uploadedIndices, i)
			}
		}
	}

	progress := float64(uploadedCount) / float64(totalChunks) * 100

	return map[string]any{
		"chunk_id":        chunkID,
		"total_chunks":    totalChunks,
		"uploaded_count":  uploadedCount,
		"uploaded_chunks": uploadedIndices, // 返回已上传的分片索引数组（最多100个）
		"progress":        progress,
		"completed":       uploadedCount == totalChunks,
	}, nil
}

// UploadFile 普通文件上传（小文件）
func (s *AttachmentServiceImpl) UploadFile(fileData []byte, filename string, mimeType string) (*models.Attachment, error) {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 生成文件路径
	ext := filepath.Ext(filename)
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%d", filename, time.Now().UnixNano())))
	uniqueName := hex.EncodeToString(hash[:]) + ext
	datePath := time.Now().Format("2006/01/02")
	finalPath := fmt.Sprintf("attachments/%s/%s", datePath, uniqueName)

	// 保存文件
	storage := facades.Storage().Disk(s.disk)
	if err := storage.Put(finalPath, string(fileData)); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 获取文件大小
	fileSize := int64(len(fileData))
	if size, err := storage.Size(finalPath); err == nil {
		fileSize = size
	}

	// 创建附件记录
	adminID := uint(0)
	if s.ctx != nil {
		if id, err := helpers.GetAdminIDFromContext(s.ctx); err == nil {
			adminID = id
		}
	}

	fileType := s.GetFileType(mimeType)
	attachment := &models.Attachment{
		AdminID:   adminID,
		Disk:      s.disk,
		Path:      finalPath,
		Filename:  filename,
		Extension: strings.TrimPrefix(ext, "."),
		MimeType:  mimeType,
		Size:      fileSize,
		Status:    1,
		FileType:  fileType,
	}

	if err := facades.Orm().Query().Create(attachment); err != nil {
		_ = storage.Delete(finalPath)
		return nil, fmt.Errorf("创建附件记录失败: %w", err)
	}

	return attachment, nil
}

// GetFileURL 获取文件访问URL
func (s *AttachmentServiceImpl) GetFileURL(attachment *models.Attachment) string {
	// 对于本地存储，返回下载接口URL
	if attachment.Disk == "local" || attachment.Disk == "public" {
		return fmt.Sprintf("/api/admin/attachments/%d/preview", attachment.ID)
	}

	// 对于云存储，生成临时URL或直接URL
	storage := facades.Storage().Disk(attachment.Disk)

	// 尝试生成临时URL（24小时有效）
	if url, err := storage.TemporaryUrl(attachment.Path, time.Now().Add(24*time.Hour)); err == nil {
		return url
	}

	// 如果生成临时URL失败，尝试从配置获取基础URL
	var configURL string
	switch attachment.Disk {
	case "s3":
		configURL = utils.GetConfigValue("storage", "s3_url", "")
	case "oss":
		configURL = utils.GetConfigValue("storage", "oss_url", "")
	case "cos":
		configURL = utils.GetConfigValue("storage", "cos_url", "")
	case "minio":
		configURL = utils.GetConfigValue("storage", "minio_url", "")
	}

	if configURL != "" {
		if !strings.HasSuffix(configURL, "/") {
			configURL += "/"
		}
		return configURL + attachment.Path
	}

	// 默认返回下载接口URL
	return fmt.Sprintf("/api/admin/attachments/%d/preview", attachment.ID)
}

// GetFileType 根据MIME类型判断文件类型
func (s *AttachmentServiceImpl) GetFileType(mimeType string) string {
	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	}
	if strings.HasPrefix(mimeType, "video/") {
		return "video"
	}
	if strings.HasPrefix(mimeType, "application/pdf") ||
		strings.HasPrefix(mimeType, "application/msword") ||
		strings.HasPrefix(mimeType, "application/vnd.openxmlformats-officedocument") ||
		strings.HasPrefix(mimeType, "application/vnd.ms-excel") ||
		strings.HasPrefix(mimeType, "text/") {
		return "document"
	}
	return "other"
}

// DeleteFile 删除文件
func (s *AttachmentServiceImpl) DeleteFile(attachment *models.Attachment) error {
	// 删除文件
	if attachment.Path != "" && attachment.Disk != "" {
		storage := facades.Storage().Disk(attachment.Disk)
		if err := storage.Delete(attachment.Path); err != nil {
			return fmt.Errorf("删除文件失败: %w", err)
		}
	}

	// 删除数据库记录
	if _, err := facades.Orm().Query().Delete(attachment); err != nil {
		return fmt.Errorf("删除附件记录失败: %w", err)
	}

	return nil
}
