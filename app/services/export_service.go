package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/utils"
)

type ExportService interface {
	// ExportToCSV 导出数据到CSV文件
	// headers: CSV表头
	// data: 数据行，每行是一个字符串切片
	// filename: 文件名（不含扩展名）
	// 返回: 文件路径和错误
	ExportToCSV(headers []string, data [][]string, filename string) (string, error)

	// ExportToFile 导出数据到文件（根据配置的格式）
	// headers: 表头
	// data: 数据行
	// filename: 文件名（不含扩展名）
	// 返回: 文件路径和错误
	ExportToFile(headers []string, data [][]string, filename string) (string, error)

	// GetExportURL 获取导出文件的访问URL
	// filePath: 文件路径
	// 返回: 访问URL
	GetExportURL(filePath string) string
}

type ExportServiceImpl struct {
	ctx    http.Context
	disk   string
	path   string
	format string
}

func NewExportService(ctx http.Context) ExportService {
	// 从数据库读取文件存储配置，如果不存在则使用默认值
	// 优先使用新的 file_disk，向后兼容 export_disk
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		// 如果 file_disk 为空，尝试读取 export_disk
		disk = utils.GetConfigValue("storage", "export_disk", "")
	}
	// 如果两个都为空或不存在，使用默认值 local
	if disk == "" {
		disk = "local"
	}

	// 记录使用的存储驱动（用于调试）
	// facades.Log().Debugf("ExportService: using storage disk: %s", disk)

	// 文件路径默认使用 exports，不再从配置读取
	path := "exports"
	// 文件格式默认使用 csv，不再从配置读取
	format := "csv"

	return &ExportServiceImpl{
		ctx:    ctx,
		disk:   disk,
		path:   path,
		format: format,
	}
}

func (s *ExportServiceImpl) ExportToCSV(headers []string, data [][]string, filename string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename = fmt.Sprintf("%s_%s.csv", filename, timestamp)
	filePath := filepath.Join(s.path, filename)

	// 创建CSV内容缓冲区
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入表头
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return "", fmt.Errorf("写入CSV表头失败: %w", err)
		}
	}

	// 写入数据
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("写入CSV数据失败: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV写入失败: %w", err)
	}

	// 获取存储驱动
	storage := facades.Storage().Disk(s.disk)

	// 写入文件
	if err := storage.Put(filePath, buf.String()); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	// 获取文件大小（如果存储驱动支持 Size 方法）
	var size int64
	if fileInfo, err := storage.Size(filePath); err == nil {
		size = fileInfo
	}

	// 记录导出日志到数据库（尽量避免影响主流程，错误仅记日志）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				facades.Log().Errorf("ExportService: panic while recording export log: %v", r)
			}
		}()

		adminID := uint(0)
		if s.ctx != nil {
			if id, err := helpers.GetAdminIDFromContext(s.ctx); err == nil {
				adminID = id
			}
		}

		ext := ""
		if dot := strings.LastIndex(filename, "."); dot != -1 {
			ext = filename[dot+1:]
		} else if dot := strings.LastIndex(filePath, "."); dot != -1 {
			ext = filePath[dot+1:]
		}

		exportRecord := models.Export{
			AdminID:   adminID,
			Disk:      s.disk,
			Path:      filePath,
			Filename:  filepath.Base(filePath),
			Extension: ext,
			Size:      size,
			Status:    1,
		}

		if err := facades.Orm().Query().Create(&exportRecord); err != nil {
			facades.Log().Errorf("ExportService: failed to record export log: %v", err)
		}
	}()

	return filePath, nil
}

// ExportToFile 导出数据到文件（根据配置的格式）
func (s *ExportServiceImpl) ExportToFile(headers []string, data [][]string, filename string) (string, error) {
	switch s.format {
	case "csv":
		return s.ExportToCSV(headers, data, filename)
	case "xlsx":
		return "", fmt.Errorf("Excel导出功能暂未实现，请使用CSV格式")
	default:
		return s.ExportToCSV(headers, data, filename)
	}
}

func (s *ExportServiceImpl) GetExportURL(filePath string) string {
	// 根据不同的存储类型从配置读取 URL
	var configURL string
	switch s.disk {
	case "s3":
		configURL = utils.GetConfigValue("storage", "s3_url", "")
	case "oss":
		configURL = utils.GetConfigValue("storage", "oss_url", "")
	case "cos":
		configURL = utils.GetConfigValue("storage", "cos_url", "")
	case "qiniu":
		configURL = utils.GetConfigValue("storage", "qiniu_domain", "")
	case "minio":
		configURL = utils.GetConfigValue("storage", "minio_url", "")
	}

	if configURL != "" {
		// 确保 URL 以 / 结尾，然后拼接文件路径
		if !strings.HasSuffix(configURL, "/") {
			configURL += "/"
		}
		return configURL + filePath
	}

	// 对于 local 和 public 存储，使用下载接口而不是直接文件路径
	// 这样可以避免被前端路由拦截
	if s.disk == "local" || s.disk == "public" {
		// 返回下载接口 URL，需要从 context 中获取导出记录 ID
		// 但这里没有 ID，所以需要修改调用方式
		// 暂时返回一个占位符，实际 URL 在 ExportController.Index 中生成
		return ""
	}

	storage := facades.Storage().Disk(s.disk)
	if url, err := storage.TemporaryUrl(filePath, time.Now().Add(24*time.Hour)); err == nil {
		return url
	}

	return "/storage/" + filePath
}
