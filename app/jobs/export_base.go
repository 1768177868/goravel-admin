package jobs

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/goravel/framework/facades"
	supportcarbon "github.com/goravel/framework/support/carbon"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

// ErrExportRecordMissing 导出记录被删除时的哨兵错误
var ErrExportRecordMissing = errors.New("export record missing (deleted)")

// ExportArgs 通用导出任务参数
type ExportArgs struct {
	ExportID uint           `json:"export_id"`
	AdminID  uint           `json:"admin_id"`
	Filters  map[string]any `json:"filters"`
	Type     string         `json:"type"`
	Language string         `json:"language"`
	Timezone string         `json:"timezone"` // 用户时区，用于时间格式化
}

// FormatTimeWithTimezone 使用指定时区格式化时间
func FormatTimeWithTimezone(t time.Time, timezone string) string {
	if t.IsZero() {
		return ""
	}
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return t.In(loc).Format("2006-01-02 15:04:05")
}

// FormatCarbonWithTimezone 使用指定时区格式化 Carbon 时间
func FormatCarbonWithTimezone(t *supportcarbon.DateTime, timezone string) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return FormatTimeWithTimezone(t.StdTime(), timezone)
}

// ExportConfig 导出配置
type ExportConfig struct {
	// 文件名前缀，如 "payments"、"orders"
	FilePrefix string
	// 表头翻译键列表
	HeaderKeys []string
	// 数据写入函数，接收 CSV writer、筛选条件 map、语言、停止检查函数
	WriteData func(w *csv.Writer, filters map[string]any, lang string, shouldStop func() bool) error
}

// BaseExporter 通用导出器基类
type BaseExporter struct {
	config ExportConfig
}

// NewBaseExporter 创建通用导出器
func NewBaseExporter(config ExportConfig) *BaseExporter {
	return &BaseExporter{config: config}
}

// ParseArgs 解析导出参数（通用逻辑）
func ParseArgs(args ...any) (ExportArgs, error) {
	if len(args) < 1 {
		return ExportArgs{}, apperrors.ErrInvalidArgument.WithMessage("missing export arguments")
	}

	var exportArgs ExportArgs
	switch v := args[0].(type) {
	case ExportArgs:
		exportArgs = v
	case string:
		if err := json.Unmarshal([]byte(v), &exportArgs); err != nil {
			facades.Log().Errorf("反序列化参数失败: %v, JSON: %s", err, v)
			return ExportArgs{}, apperrors.ErrInvalidArgument.WithMessage(fmt.Sprintf("failed to unmarshal export arguments: %v", err))
		}
	case map[string]any:
		if exportID, ok := utils.GetUint(v, "export_id"); ok {
			exportArgs.ExportID = exportID
		}
		if adminID, ok := utils.GetUint(v, "admin_id"); ok {
			exportArgs.AdminID = adminID
		}
		if filters, ok := utils.GetMap(v, "filters"); ok {
			exportArgs.Filters = filters
		}
		if exportType, ok := utils.GetString(v, "type"); ok {
			exportArgs.Type = exportType
		}
		if lang, ok := utils.GetString(v, "language"); ok {
			exportArgs.Language = lang
		}
	default:
		return ExportArgs{}, apperrors.ErrInvalidArgument.WithMessage(fmt.Sprintf("invalid export arguments type: %T", args[0]))
	}

	if exportArgs.ExportID == 0 {
		return ExportArgs{}, apperrors.ErrInvalidArgument.WithMessage("export_id is required")
	}

	return exportArgs, nil
}

// MarkExportFailed 标记导出失败
func MarkExportFailed(exportID uint, errorMsg string) {
	if exportID == 0 {
		return
	}
	var failedRecord models.Export
	if queryErr := facades.Orm().Query().Where("id", exportID).First(&failedRecord); queryErr == nil {
		failedRecord.Status = models.ExportStatusFailed
		failedRecord.ErrorMsg = errorMsg
		if saveErr := facades.Orm().Query().Save(&failedRecord); saveErr != nil {
			facades.Log().Errorf("更新导出记录失败状态失败: export_id=%d, error=%v", exportID, saveErr)
		}
	}
}

// CheckAndUpdateExportStatus 检查导出记录并更新状态为处理中
// 返回导出记录和错误，如果记录不存在返回 nil, nil
func CheckAndUpdateExportStatus(exportID uint) (*models.Export, error) {
	exists, err := facades.Orm().Query().Model(&models.Export{}).Where("id", exportID).Exists()
	if err != nil {
		errorlog.Record(context.TODO(), "export", "检查导出记录是否存在失败", map[string]any{
			"export_id": exportID,
			"error":     err.Error(),
		}, "检查导出记录是否存在失败: %v", err)
		return nil, err
	}
	if !exists {
		facades.Log().Infof("导出记录不存在，停止任务: export_id=%d", exportID)
		return nil, nil
	}

	var exportRecords []models.Export
	if err := facades.Orm().Query().Where("id", exportID).Limit(1).Get(&exportRecords); err != nil {
		errorlog.Record(context.TODO(), "export", "查询导出记录失败", map[string]any{
			"export_id": exportID,
			"error":     err.Error(),
		}, "查询导出记录失败: %v", err)
		return nil, err
	}
	if len(exportRecords) == 0 {
		facades.Log().Infof("导出记录已被删除，停止任务: export_id=%d", exportID)
		return nil, nil
	}

	exportRecord := &exportRecords[0]
	exportRecord.Status = models.ExportStatusProcessing
	exportRecord.ErrorMsg = ""
	if err := facades.Orm().Query().Save(exportRecord); err != nil {
		facades.Log().Errorf("更新导出状态为处理中失败: export_id=%d, error=%v", exportID, err)
		return nil, err
	}

	return exportRecord, nil
}

// Execute 执行导出（通用流程）
func (e *BaseExporter) Execute(args ExportArgs) error {
	// 获取语言
	lang := args.Language
	if lang == "" {
		lang = facades.Config().GetString("app.locale", "cn")
	}
	lang = utils.NormalizeLanguage(lang)

	// 将时区放入 filters 供 WriteData 使用
	if args.Timezone != "" {
		if args.Filters == nil {
			args.Filters = make(map[string]any)
		}
		args.Filters["_timezone"] = args.Timezone
	}

	// 翻译表头
	headers := utils.TranslateHeaders(e.config.HeaderKeys, lang)

	// 生成文件名和路径
	exportService := services.NewExportService(nil)
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%d_%s.csv", e.config.FilePrefix, args.ExportID, timestamp)
	filePath := path.Join("exports", filename)

	// 预写入文件信息
	e.preWriteFileInfo(args.ExportID, filePath, filename)

	// 创建 shouldStop 和进度更新回调
	lastUpdateAt := time.Now().Add(-10 * time.Second)
	lastExistCheckAt := time.Now().Add(-10 * time.Second)
	shouldStop := func() bool {
		if time.Since(lastExistCheckAt) < 2*time.Second {
			return false
		}
		lastExistCheckAt = time.Now()
		exists, err := facades.Orm().Query().Model(&models.Export{}).Where("id", args.ExportID).Exists()
		if err != nil {
			return false
		}
		return !exists
	}

	// 执行流式导出
	filePath, err := exportService.ExportToCSVStreamAtWithProgress(headers, filePath, func(w *csv.Writer) error {
		return e.config.WriteData(w, args.Filters, lang, shouldStop)
	}, func(writtenBytes int64) {
		if shouldStop() {
			return
		}
		if time.Since(lastUpdateAt) < 3*time.Second {
			return
		}
		lastUpdateAt = time.Now()
		result, _ := facades.Orm().Query().Model(&models.Export{}).Where("id", args.ExportID).Update(map[string]any{
			"size": writtenBytes,
		})
		if result == nil || result.RowsAffected == 0 {
			lastExistCheckAt = time.Now().Add(-10 * time.Second)
		}
	}, true)

	if err != nil {
		if shouldStop() {
			return ErrExportRecordMissing
		}
		errorlog.Record(context.TODO(), "export", "导出文件失败", map[string]any{
			"export_id": args.ExportID,
			"filename":  filename,
			"error":     err.Error(),
		}, "导出文件失败: %v", err)
		return fmt.Errorf("导出文件失败: %v", err)
	}

	// 更新导出记录为成功
	return e.finalizeExport(args.ExportID, filePath)
}

// preWriteFileInfo 预写入文件信息
func (e *BaseExporter) preWriteFileInfo(exportID uint, filePath, filename string) {
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportID).First(&exportRecord); err == nil {
		changed := false
		if exportRecord.Path == "" {
			exportRecord.Path = filePath
			changed = true
		}
		if exportRecord.Filename == "" {
			exportRecord.Filename = filename
			changed = true
		}
		if exportRecord.Extension == "" {
			exportRecord.Extension = "csv"
			changed = true
		}
		if changed {
			_ = facades.Orm().Query().Save(&exportRecord)
		}
	}
}

// finalizeExport 完成导出，更新记录
func (e *BaseExporter) finalizeExport(exportID uint, filePath string) error {
	var exportRecord models.Export
	if err := facades.Orm().Query().Where("id", exportID).First(&exportRecord); err != nil {
		return nil
	}

	exportRecord.Path = filePath
	exportRecord.Filename = path.Base(filePath)

	if ext := path.Ext(filePath); ext != "" {
		exportRecord.Extension = ext[1:]
	} else {
		exportRecord.Extension = "csv"
	}

	if exportRecord.Disk != "" {
		storage := facades.Storage().Disk(exportRecord.Disk)
		if fileInfo, err := storage.Size(filePath); err == nil {
			exportRecord.Size = fileInfo
		}
	}

	exportRecord.Status = models.ExportStatusSuccess
	exportRecord.ErrorMsg = ""

	if err := facades.Orm().Query().Save(&exportRecord); err != nil {
		facades.Log().Errorf("保存导出记录失败: export_id=%d, error=%v", exportID, err)
		return fmt.Errorf("更新导出记录失败: %v", err)
	}

	return nil
}

// IsTableNotExistsError 判断是否是表不存在错误
func IsTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "1146") || strings.Contains(msg, "42s02") {
		return true
	}
	return strings.Contains(msg, "table") && (strings.Contains(msg, "doesn't exist") || strings.Contains(msg, "does not exist"))
}

// ParseOrderBy 解析排序字符串
func ParseOrderBy(orderBy string) (field, direction string) {
	parts := strings.Split(orderBy, ":")
	field = "created_at"
	direction = "desc"

	if len(parts) >= 1 && parts[0] != "" {
		field = parts[0]
	}
	if len(parts) >= 2 && (parts[1] == "asc" || parts[1] == "desc") {
		direction = parts[1]
	}
	return
}

// GetDefaultTimeRange 获取默认时间范围（7天前到现在）
func GetDefaultTimeRange(filters map[string]any) (startTime, endTime time.Time) {
	if startTimeStr, ok := utils.GetString(filters, "start_time"); ok && startTimeStr != "" {
		if t, err := utils.ParseDateTimeUTC(startTimeStr); err == nil {
			startTime = t
		}
	}
	if endTimeStr, ok := utils.GetString(filters, "end_time"); ok && endTimeStr != "" {
		if t, err := utils.ParseDateTimeUTC(endTimeStr); err == nil {
			endTime = t
		}
	}

	if startTime.IsZero() {
		startTime = time.Now().UTC().AddDate(0, 0, -7)
	}
	if endTime.IsZero() {
		endTime = time.Now().UTC()
	}
	return
}
