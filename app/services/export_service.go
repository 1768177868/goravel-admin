package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/minio/minio-go/v7"
	miniocreds "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/tencentyun/cos-go-sdk-v5"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

type ExportService interface {
	// ExportToCSV 导出数据到CSV文件
	// headers: CSV表头
	// data: 数据行，每行是一个字符串切片
	// filename: 文件名（不含扩展名）
	// skipAutoCreate: 是否跳过自动创建导出记录（用于异步任务）
	// 返回: 文件路径和错误
	ExportToCSV(headers []string, data [][]string, filename string, skipAutoCreate ...bool) (string, error)

	// ExportToCSVStream 流式导出到 CSV（只支持 local/public 磁盘，避免百万级导出把内存打爆）
	// headers: CSV 表头
	// filename: 文件名（不含扩展名）
	// write: 回调里持续调用 writer.Write(row) 写入数据；回调返回 error 会终止导出
	// skipAutoCreate: 是否跳过自动创建导出记录（用于异步任务）
	ExportToCSVStream(headers []string, filename string, write func(writer *csv.Writer) error, skipAutoCreate ...bool) (string, error)

	// ExportToCSVStreamAt 流式导出到指定 filePath（包含目录+文件名，如 exports/orders_1_20260107.csv）
	// 用于“导出中”就先写入 exports 表的 Path/Filename 等字段，完成后只更新 size/status。
	ExportToCSVStreamAt(headers []string, filePath string, write func(writer *csv.Writer) error, skipAutoCreate ...bool) (string, error)

	// ExportToCSVStreamAtWithProgress 流式导出到指定 filePath，并通过回调返回已写入字节数（可用于实时更新 exports.size）
	ExportToCSVStreamAtWithProgress(headers []string, filePath string, write func(writer *csv.Writer) error, onProgress func(writtenBytes int64), skipAutoCreate ...bool) (string, error)

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
	// 优先使用 file_disk，如果没有则使用 storage_disk（向后兼容），再尝试 export_disk，最后使用默认值 local
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = utils.GetConfigValue("storage", "storage_disk", "")
	}
	if disk == "" {
		// 向后兼容 export_disk
		disk = utils.GetConfigValue("storage", "export_disk", "")
	}
	// 如果都不存在，使用默认值 local
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

// ExportToCSVStream 流式导出 CSV（仅 local/public）
func (s *ExportServiceImpl) ExportToCSVStream(headers []string, filename string, write func(writer *csv.Writer) error, skipAutoCreate ...bool) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename = fmt.Sprintf("%s_%s.csv", filename, timestamp)
	// 注意：存储路径统一使用 "/"，避免 Windows 下 filepath.Join 生成 "\" 导致云存储对象 key 异常
	filePath := path.Join(s.path, filename)

	return s.ExportToCSVStreamAt(headers, filePath, write, skipAutoCreate...)
}

func (s *ExportServiceImpl) ExportToCSVStreamAt(headers []string, filePath string, write func(writer *csv.Writer) error, skipAutoCreate ...bool) (string, error) {
	return s.ExportToCSVStreamAtWithProgress(headers, filePath, write, nil, skipAutoCreate...)
}

type progressWriter struct {
	w        io.Writer
	written  int64
	lastTick time.Time
	interval time.Duration
	cb       func(int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	if n > 0 {
		p.written += int64(n)
		if p.cb != nil && (p.interval <= 0 || time.Since(p.lastTick) >= p.interval) {
			p.lastTick = time.Now()
			p.cb(p.written)
		}
	}
	return n, err
}

func (s *ExportServiceImpl) ExportToCSVStreamAtWithProgress(headers []string, filePath string, write func(writer *csv.Writer) error, onProgress func(writtenBytes int64), skipAutoCreate ...bool) (string, error) {
	// 规范化
	filePath = path.Clean(strings.ReplaceAll(filePath, "\\", "/"))
	displayName := path.Base(filePath)

	// local/public：直接写到磁盘 root 下（最省资源）
	if s.disk == "local" || s.disk == "public" {
		// 获取磁盘 root，直接写到 root 下，避免先缓存在内存再 Put
		root := facades.Config().GetString(fmt.Sprintf("filesystems.disks.%s.root", s.disk), "")
		if root == "" {
			return "", fmt.Errorf("filesystems.disks.%s.root is empty, can't stream write", s.disk)
		}

		absPath := filepath.Join(root, filepath.FromSlash(filePath))
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			if s.ctx != nil {
				errorlog.RecordHTTP(s.ctx, "export", "创建导出目录失败", map[string]any{
					"abs_path": absPath,
					"error":    err.Error(),
				}, "创建导出目录失败: %w", err)
			}
			return "", fmt.Errorf("创建导出目录失败: %w", err)
		}

		f, err := os.Create(absPath)
		if err != nil {
			if s.ctx != nil {
				errorlog.RecordHTTP(s.ctx, "export", "创建导出文件失败", map[string]any{
					"abs_path": absPath,
					"error":    err.Error(),
				}, "创建导出文件失败: %w", err)
			}
			return "", fmt.Errorf("创建导出文件失败: %w", err)
		}
		defer func() {
			_ = f.Close()
		}()

		pw := &progressWriter{
			w:        f,
			lastTick: time.Now(),
			interval: 2 * time.Second,
			cb:       onProgress,
		}
		writer := csv.NewWriter(pw)

		// 写入表头
		if len(headers) > 0 {
			if err := writer.Write(headers); err != nil {
				if s.ctx != nil {
					errorlog.RecordHTTP(s.ctx, "export", "写入CSV表头失败", map[string]any{
						"filename": displayName,
						"error":    err.Error(),
					}, "写入CSV表头失败: %w", err)
				}
				return "", apperrors.ErrWriteCSVHeaderFailed.WithError(err)
			}
		}

		// 由调用方流式写入数据
		if err := write(writer); err != nil {
			if s.ctx != nil {
				errorlog.RecordHTTP(s.ctx, "export", "写入CSV数据失败", map[string]any{
					"filename": displayName,
					"error":    err.Error(),
				}, "写入CSV数据失败: %w", err)
			}
			return "", apperrors.ErrWriteCSVDataFailed.WithError(err)
		}

		writer.Flush()
		if onProgress != nil {
			onProgress(pw.written)
		}
		if err := writer.Error(); err != nil {
			if s.ctx != nil {
				errorlog.RecordHTTP(s.ctx, "export", "CSV写入失败", map[string]any{
					"filename": displayName,
					"error":    err.Error(),
				}, "CSV写入失败: %w", err)
			}
			return "", apperrors.ErrCSVWriteFailed.WithError(err)
		}

		// 记录导出日志到数据库（尽量避免影响主流程，错误仅记日志）
		shouldSkip := len(skipAutoCreate) > 0 && skipAutoCreate[0]
		if !shouldSkip {
			// 复用 ExportToCSV 的异步记录逻辑：这里先简单沿用（不阻塞主流程）
			go func() {
				defer func() {
					if r := recover(); r != nil {
						facades.Log().Errorf("ExportService: panic while recording export log: %v", r)
					}
				}()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				s.recordExportLogWithContext(ctx, filePath, absPath)
			}()
		}

		return filePath, nil
	}

	// 云存储（s3/oss/cos/minio）：先流式写到本地临时文件，再“流式/文件上传”到云盘，然后删除临时文件
	tmpFile, err := os.CreateTemp(os.TempDir(), "export-*.csv")
	if err != nil {
		return "", fmt.Errorf("创建临时导出文件失败: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
	}()

	pw := &progressWriter{
		w:        tmpFile,
		lastTick: time.Now(),
		interval: 2 * time.Second,
		cb:       onProgress,
	}
	writer := csv.NewWriter(pw)
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			_ = os.Remove(tmpPath)
			return "", apperrors.ErrWriteCSVHeaderFailed.WithError(err)
		}
	}
	if err := write(writer); err != nil {
		writer.Flush()
		if onProgress != nil {
			onProgress(pw.written)
		}
		_ = os.Remove(tmpPath)
		return "", apperrors.ErrWriteCSVDataFailed.WithError(err)
	}
	writer.Flush()
	if onProgress != nil {
		onProgress(pw.written)
	}
	if err := writer.Error(); err != nil {
		_ = os.Remove(tmpPath)
		return "", apperrors.ErrCSVWriteFailed.WithError(err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("关闭临时导出文件失败: %w", err)
	}

	if err := s.uploadLocalFileToCloudDisk(tmpPath, filePath); err != nil {
		// 上传失败：保留临时文件便于排查（同时记录日志）
		if s.ctx != nil {
			errorlog.RecordHTTP(s.ctx, "export", "上传云存储失败", map[string]any{
				"disk":      s.disk,
				"tmp_path":  tmpPath,
				"dest_path": filePath,
				"error":     err.Error(),
			}, "上传云存储失败: %v", err)
		} else {
			facades.Log().Errorf("export upload failed: disk=%s tmp=%s dest=%s err=%v", s.disk, tmpPath, filePath, err)
		}
		return "", err
	}

	_ = os.Remove(tmpPath)

	shouldSkip := len(skipAutoCreate) > 0 && skipAutoCreate[0]
	if !shouldSkip {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					facades.Log().Errorf("ExportService: panic while recording export log: %v", r)
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.recordExportLogWithContext(ctx, filePath, tmpPath)
		}()
	}

	return filePath, nil
}

// ExportToCSV 导出数据到CSV文件
// skipAutoCreate: 是否跳过自动创建导出记录（用于异步任务，避免重复创建）
func (s *ExportServiceImpl) ExportToCSV(headers []string, data [][]string, filename string, skipAutoCreate ...bool) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename = fmt.Sprintf("%s_%s.csv", filename, timestamp)
	filePath := path.Join(s.path, filename)

	// 创建CSV内容缓冲区
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入表头
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			if s.ctx != nil {
				errorlog.RecordHTTP(s.ctx, "export", "写入CSV表头失败", map[string]any{
					"filename": filename,
					"error":    err.Error(),
				}, "写入CSV表头失败: %w", err)
			}
			return "", apperrors.ErrWriteCSVHeaderFailed.WithError(err)
		}
	}

	// 写入数据
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			if s.ctx != nil {
				errorlog.RecordHTTP(s.ctx, "export", "写入CSV数据失败", map[string]any{
					"filename": filename,
					"error":    err.Error(),
				}, "写入CSV数据失败: %w", err)
			}
			return "", apperrors.ErrWriteCSVDataFailed.WithError(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		if s.ctx != nil {
			errorlog.RecordHTTP(s.ctx, "export", "CSV写入失败", map[string]any{
				"filename": filename,
				"error":    err.Error(),
			}, "CSV写入失败: %w", err)
		}
		return "", apperrors.ErrCSVWriteFailed.WithError(err)
	}

	// 获取存储驱动
	storage := facades.Storage().Disk(s.disk)

	// 写入文件
	if err := storage.Put(filePath, buf.String()); err != nil {
		if s.ctx != nil {
			errorlog.RecordHTTP(s.ctx, "export", "保存文件失败", map[string]any{
				"filename":  filename,
				"file_path": filePath,
				"error":     err.Error(),
			}, "保存文件失败: %w", err)
		}
		return "", apperrors.ErrSaveFileFailed.WithError(err)
	}

	// 获取文件大小（如果存储驱动支持 Size 方法）
	var size int64
	if fileInfo, err := storage.Size(filePath); err == nil {
		size = fileInfo
	}

	// 记录导出日志到数据库（尽量避免影响主流程，错误仅记日志）
	// 如果 skipAutoCreate 为 true，则跳过自动创建（用于异步任务）
	shouldSkip := len(skipAutoCreate) > 0 && skipAutoCreate[0]
	if !shouldSkip {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					facades.Log().Errorf("ExportService: panic while recording export log: %v", r)
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

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

			// 尝试从 context 中获取导出类型
			exportType := ""
			if s.ctx != nil {
				if typeValue := s.ctx.Value("export_type"); typeValue != nil {
					if typeStr, ok := typeValue.(string); ok {
						exportType = typeStr
					}
				}
			}

			exportRecord := models.Export{
				AdminID:   adminID,
				Type:      exportType,
				Disk:      s.disk,
				Path:      filePath,
				Filename:  filepath.Base(filePath),
				Extension: ext,
				Size:      size,
				Status:    1,
			}

			// 使用带超时的 context 执行数据库操作
			// 注意：虽然 ORM 可能不支持 WithContext，但超时控制通过 goroutine 的 context 实现
			done := make(chan error, 1)
			go func() {
				done <- facades.Orm().Query().Create(&exportRecord)
			}()
			
			select {
			case err := <-done:
				if err != nil {
					facades.Log().Errorf("ExportService: failed to record export log: %v", err)
				}
			case <-ctx.Done():
				facades.Log().Errorf("ExportService: timeout while recording export log")
			}
		}()
	}

	return filePath, nil
}

func (s *ExportServiceImpl) recordExportLog(filePath string, absOrTmpPathForSize string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.recordExportLogWithContext(ctx, filePath, absOrTmpPathForSize)
}

// recordExportLogWithContext 使用指定的 context 记录导出日志（支持超时控制）
func (s *ExportServiceImpl) recordExportLogWithContext(ctx context.Context, filePath string, absOrTmpPathForSize string) {
	defer func() {
		if r := recover(); r != nil {
			facades.Log().Errorf("ExportService: panic while recording export log: %v", r)
		}
	}()

	size := int64(0)
	if fi, err := os.Stat(absOrTmpPathForSize); err == nil {
		size = fi.Size()
	}

	adminID := uint(0)
	if s.ctx != nil {
		if id, err := helpers.GetAdminIDFromContext(s.ctx); err == nil {
			adminID = id
		}
	}

	ext := "csv"
	exportType := ""
	if s.ctx != nil {
		if typeValue := s.ctx.Value("export_type"); typeValue != nil {
			if typeStr, ok := typeValue.(string); ok {
				exportType = typeStr
			}
		}
	}

	exportRecord := models.Export{
		AdminID:   adminID,
		Type:      exportType,
		Disk:      s.disk,
		Path:      filePath,
		Filename:  path.Base(filePath),
		Extension: ext,
		Size:      size,
		Status:    1,
	}

	// 使用带超时的 context 执行数据库操作
	// 注意：虽然 ORM 可能不支持 WithContext，但超时控制通过 goroutine 的 context 实现
	done := make(chan error, 1)
	go func() {
		done <- facades.Orm().Query().Create(&exportRecord)
	}()
	
	select {
	case err := <-done:
		if err != nil {
			facades.Log().Errorf("ExportService: failed to record export log: %v", err)
		}
	case <-ctx.Done():
		facades.Log().Errorf("ExportService: timeout while recording export log")
	}
}

func (s *ExportServiceImpl) uploadLocalFileToCloudDisk(localFilePath string, destPath string) error {
	destPath = strings.TrimPrefix(destPath, "/")
	switch s.disk {
	case "s3":
		return s.uploadToS3(localFilePath, destPath)
	case "minio":
		return s.uploadToMinio(localFilePath, destPath)
	case "oss":
		return s.uploadToOss(localFilePath, destPath)
	case "cos":
		return s.uploadToCos(localFilePath, destPath)
	default:
		return fmt.Errorf("unsupported cloud disk for stream export: %s", s.disk)
	}
}

func (s *ExportServiceImpl) uploadToS3(localFilePath, key string) error {
	cfg := facades.Config()
	accessKeyId := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.key", s.disk))
	accessKeySecret := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.secret", s.disk))
	region := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.region", s.disk))
	bucket := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.bucket", s.disk))
	token := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.token", s.disk))
	endpoint := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.endpoint", s.disk))
	usePathStyle := cfg.GetBool(fmt.Sprintf("filesystems.disks.%s.use_path_style", s.disk), true)
	objectCannedACL := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.object_canned_acl", s.disk))
	if accessKeyId == "" || accessKeySecret == "" || region == "" || bucket == "" {
		return fmt.Errorf("please set %s configuration first", s.disk)
	}

	options := s3.Options{
		Region: region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, token)),
		UsePathStyle: usePathStyle,
	}
	if endpoint != "" {
		options.BaseEndpoint = aws.String(endpoint)
	}
	client := s3.New(options)

	f, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          f,
		ContentLength: aws.Int64(fi.Size()),
		ContentType:   aws.String("text/csv; charset=utf-8"),
	}
	if objectCannedACL != "" {
		input.ACL = types.ObjectCannedACL(objectCannedACL)
	}

	_, err = client.PutObject(context.Background(), input)
	return err
}

func (s *ExportServiceImpl) uploadToMinio(localFilePath, key string) error {
	cfg := facades.Config()
	accessKeyId := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.key", s.disk))
	accessKeySecret := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.secret", s.disk))
	region := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.region", s.disk))
	bucket := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.bucket", s.disk))
	endpoint := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.endpoint", s.disk))
	ssl := cfg.GetBool(fmt.Sprintf("filesystems.disks.%s.ssl", s.disk), false)
	if accessKeyId == "" || accessKeySecret == "" || bucket == "" || endpoint == "" {
		return fmt.Errorf("please set %s configuration first", s.disk)
	}
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  miniocreds.NewStaticV4(accessKeyId, accessKeySecret, ""),
		Secure: ssl,
		Region: region,
	})
	if err != nil {
		return err
	}

	_, err = client.FPutObject(context.Background(), bucket, key, localFilePath, minio.PutObjectOptions{
		ContentType: "text/csv; charset=utf-8",
	})
	return err
}

func (s *ExportServiceImpl) uploadToOss(localFilePath, key string) error {
	cfg := facades.Config()
	accessKeyId := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.key", s.disk))
	accessKeySecret := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.secret", s.disk))
	bucket := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.bucket", s.disk))
	endpoint := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.endpoint", s.disk))
	if accessKeyId == "" || accessKeySecret == "" || bucket == "" || endpoint == "" {
		return fmt.Errorf("please set %s configuration first", s.disk)
	}
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}
	bucketInstance, err := client.Bucket(bucket)
	if err != nil {
		return err
	}
	return bucketInstance.PutObjectFromFile(key, localFilePath)
}

func (s *ExportServiceImpl) uploadToCos(localFilePath, key string) error {
	cfg := facades.Config()
	accessKeyId := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.key", s.disk))
	accessKeySecret := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.secret", s.disk))
	cosUrl := cfg.GetString(fmt.Sprintf("filesystems.disks.%s.url", s.disk))
	if accessKeyId == "" || accessKeySecret == "" || cosUrl == "" {
		return fmt.Errorf("please set %s configuration first", s.disk)
	}
	u, err := url.Parse(cosUrl)
	if err != nil {
		return err
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &stdhttp.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKeyId,
			SecretKey: accessKeySecret,
		},
	})
	_, _, err = client.Object.Upload(context.Background(), key, localFilePath, nil)
	return err
}

// ExportToFile 导出数据到文件（根据配置的格式）
func (s *ExportServiceImpl) ExportToFile(headers []string, data [][]string, filename string) (string, error) {
	switch s.format {
	case "csv":
		return s.ExportToCSV(headers, data, filename)
	case "xlsx":
		return "", apperrors.ErrExcelNotImplemented
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
