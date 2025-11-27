package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"time"

	"github.com/goravel/framework/facades"
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
	disk   string
	path   string
	format string
}

func NewExportService() ExportService {
	disk := facades.Config().GetString("export.disk", "local")
	path := facades.Config().GetString("export.path", "exports")
	format := facades.Config().GetString("export.format", "csv")

	return &ExportServiceImpl{
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
	urlPrefix := facades.Config().GetString("export.url_prefix", "")
	if urlPrefix != "" {
		return urlPrefix + "/" + filePath
	}

	if s.disk == "local" || s.disk == "public" {
		storage := facades.Storage().Disk(s.disk)
		url := storage.Url(filePath)
		if url != "" {
			return url
		}
	}

	storage := facades.Storage().Disk(s.disk)
	if url, err := storage.TemporaryUrl(filePath, time.Now().Add(24*time.Hour)); err == nil {
		return url
	}

	return "/storage/" + filePath
}
