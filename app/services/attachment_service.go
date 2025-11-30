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
	// InitChunkUpload 初始化分片上传
	InitChunkUpload(filename string, totalSize int64, chunkSize int64, totalChunks int) (string, error)

	// UploadChunk 上传分片
	UploadChunk(chunkID string, chunkIndex int, chunkData []byte) error

	// MergeChunks 合并分片
	MergeChunks(chunkID string, filename string, mimeType string) (*models.Attachment, error)

	// GetChunkProgress 获取分片上传进度
	GetChunkProgress(chunkID string) (map[string]any, error)

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
	ctx  http.Context
	disk string
}

func NewAttachmentService(ctx http.Context) AttachmentService {
	// 从数据库读取文件存储配置
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = "local"
	}

	return &AttachmentServiceImpl{
		ctx:  ctx,
		disk: disk,
	}
}

// InitChunkUpload 初始化分片上传
func (s *AttachmentServiceImpl) InitChunkUpload(filename string, totalSize int64, chunkSize int64, totalChunks int) (string, error) {
	// 生成唯一的分片ID
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%d_%d", filename, totalSize, time.Now().UnixNano())))
	chunkID := hex.EncodeToString(hash[:])

	// 存储分片信息到缓存
	chunkInfo := map[string]any{
		"filename":    filename,
		"total_size":  totalSize,
		"chunk_size":  chunkSize,
		"total_chunks": totalChunks,
		"uploaded_chunks": make([]bool, totalChunks),
		"created_at":  time.Now().Unix(),
	}

	// 缓存24小时
	cacheKey := fmt.Sprintf("attachment:chunk:%s", chunkID)
	if err := facades.Cache().Put(cacheKey, chunkInfo, 24*time.Hour); err != nil {
		return "", fmt.Errorf("保存分片信息失败: %w", err)
	}

	return chunkID, nil
}

// UploadChunk 上传分片
func (s *AttachmentServiceImpl) UploadChunk(chunkID string, chunkIndex int, chunkData []byte) error {
	cacheKey := fmt.Sprintf("attachment:chunk:%s", chunkID)
	
	// 获取分片信息
	var chunkInfo map[string]any
	if err := facades.Cache().Get(cacheKey, &chunkInfo); err != nil {
		return fmt.Errorf("分片信息不存在或已过期: %w", err)
	}

	totalChunks := int(chunkInfo["total_chunks"].(int))
	if chunkIndex < 0 || chunkIndex >= totalChunks {
		return fmt.Errorf("分片索引超出范围")
	}

	// 保存分片到临时目录
	storage := facades.Storage().Disk(s.disk)
	chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, chunkIndex)
	
	if err := storage.Put(chunkPath, string(chunkData)); err != nil {
		return fmt.Errorf("保存分片失败: %w", err)
	}

	// 更新已上传分片标记
	uploadedChunks := chunkInfo["uploaded_chunks"].([]bool)
	if len(uploadedChunks) <= chunkIndex {
		// 扩展数组
		newArray := make([]bool, totalChunks)
		copy(newArray, uploadedChunks)
		uploadedChunks = newArray
	}
	uploadedChunks[chunkIndex] = true
	chunkInfo["uploaded_chunks"] = uploadedChunks

	// 更新缓存
	if err := facades.Cache().Put(cacheKey, chunkInfo, 24*time.Hour); err != nil {
		return fmt.Errorf("更新分片信息失败: %w", err)
	}

	return nil
}

// MergeChunks 合并分片
func (s *AttachmentServiceImpl) MergeChunks(chunkID string, filename string, mimeType string) (*models.Attachment, error) {
	cacheKey := fmt.Sprintf("attachment:chunk:%s", chunkID)
	
	// 获取分片信息
	var chunkInfo map[string]any
	if err := facades.Cache().Get(cacheKey, &chunkInfo); err != nil {
		return nil, fmt.Errorf("分片信息不存在或已过期: %w", err)
	}

	totalChunks := int(chunkInfo["total_chunks"].(int))
	uploadedChunks := chunkInfo["uploaded_chunks"].([]bool)

	// 检查所有分片是否都已上传
	for i := 0; i < totalChunks; i++ {
		if i >= len(uploadedChunks) || !uploadedChunks[i] {
			return nil, fmt.Errorf("分片 %d 未上传", i)
		}
	}

	storage := facades.Storage().Disk(s.disk)

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

	// 合并分片
	var mergedData []byte
	for i := 0; i < totalChunks; i++ {
		chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, i)
		chunkContent, err := storage.Get(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("读取分片 %d 失败: %w", i, err)
		}
		mergedData = append(mergedData, []byte(chunkContent)...)
	}

	// 保存合并后的文件
	if err := storage.Put(finalPath, string(mergedData)); err != nil {
		return nil, fmt.Errorf("保存合并文件失败: %w", err)
	}

	// 清理分片文件
	for i := 0; i < totalChunks; i++ {
		chunkPath := fmt.Sprintf("chunks/%s/%d", chunkID, i)
		_ = storage.Delete(chunkPath) // 忽略删除错误
	}

	// 清理缓存
	_ = facades.Cache().Forget(cacheKey)

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

	return attachment, nil
}

// GetChunkProgress 获取分片上传进度
func (s *AttachmentServiceImpl) GetChunkProgress(chunkID string) (map[string]any, error) {
	cacheKey := fmt.Sprintf("attachment:chunk:%s", chunkID)
	
	var chunkInfo map[string]any
	if err := facades.Cache().Get(cacheKey, &chunkInfo); err != nil {
		return nil, fmt.Errorf("分片信息不存在或已过期: %w", err)
	}

	uploadedChunks := chunkInfo["uploaded_chunks"].([]bool)
	uploadedCount := 0
	for _, uploaded := range uploadedChunks {
		if uploaded {
			uploadedCount++
		}
	}

	totalChunks := int(chunkInfo["total_chunks"].(int))
	progress := float64(uploadedCount) / float64(totalChunks) * 100

	return map[string]any{
		"chunk_id":       chunkID,
		"total_chunks":   totalChunks,
		"uploaded_count": uploadedCount,
		"progress":       progress,
		"completed":      uploadedCount == totalChunks,
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

