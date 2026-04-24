package jobs

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"
)

// ExportGenericArgs 通用导出任务参数
type ExportGenericArgs struct {
	ExportArgs
	Table       string            `json:"table"`
	FilePrefix  string            `json:"file_prefix"`
	HeaderKeys  []string          `json:"header_keys"`
	Columns     []string          `json:"columns"`
	SearchTypes map[string]string `json:"search_types"`
}

// ExportGeneric 通用导出任务
type ExportGeneric struct{}

func (r *ExportGeneric) Signature() string {
	return "export_generic"
}

func (r *ExportGeneric) Handle(args ...any) (retErr error) {
	exportID := tryExtractExportID(args...)

	defer func() {
		if rec := recover(); rec != nil {
			errorMsg := fmt.Sprintf("panic: %v", rec)
			facades.Log().Errorf("ExportGeneric Job panic: %v", rec)
			MarkExportFailed(exportID, errorMsg)
			retErr = fmt.Errorf("%s", errorMsg)
		}
	}()

	jobArgs, err := parseExportGenericArgs(args...)
	if err != nil {
		if exportID > 0 {
			MarkExportFailed(exportID, err.Error())
		}
		return err
	}
	exportID = jobArgs.ExportID

	lock, err := AcquireExportExecutionLock(exportID)
	if err != nil {
		return err
	}
	if lock == nil {
		facades.Log().Infof("导出任务已在执行中，跳过重复投递: export_id=%d", exportID)
		return nil
	}
	defer lock.Release()

	exportRecord, err := CheckAndUpdateExportStatus(exportID)
	if err != nil {
		return err
	}
	if exportRecord == nil {
		return nil
	}

	exporter := NewBaseExporter(ExportConfig{
		FilePrefix: jobArgs.FilePrefix,
		HeaderKeys: jobArgs.HeaderKeys,
		WriteData: func(w *csv.Writer, filters map[string]any, lang string, shouldStop func() bool) error {
			return r.writeGenericRows(w, jobArgs, filters, shouldStop)
		},
	})

	jobErr := exporter.Execute(jobArgs.ExportArgs)

	if errors.Is(jobErr, ErrExportRecordMissing) {
		return nil
	}

	if jobErr != nil {
		MarkExportFailed(exportID, jobErr.Error())
		return jobErr
	}

	return nil
}

func tryExtractExportID(args ...any) uint {
	if len(args) == 0 {
		return 0
	}

	switch v := args[0].(type) {
	case ExportGenericArgs:
		return v.ExportID
	case string:
		var payload map[string]any
		if err := json.Unmarshal([]byte(v), &payload); err == nil {
			return cast.ToUint(payload["export_id"])
		}
	case map[string]any:
		return cast.ToUint(v["export_id"])
	}
	return 0
}

func parseExportGenericArgs(args ...any) (ExportGenericArgs, error) {
	if len(args) < 1 {
		return ExportGenericArgs{}, errors.New("missing export arguments")
	}

	switch v := args[0].(type) {
	case ExportGenericArgs:
		return v, nil
	case string:
		var parsed ExportGenericArgs
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return ExportGenericArgs{}, fmt.Errorf("failed to unmarshal export generic args: %w", err)
		}
		return parsed, nil
	case map[string]any:
		raw, err := json.Marshal(v)
		if err != nil {
			return ExportGenericArgs{}, err
		}
		var parsed ExportGenericArgs
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return ExportGenericArgs{}, err
		}
		return parsed, nil
	default:
		return ExportGenericArgs{}, fmt.Errorf("invalid export generic args type: %T", args[0])
	}
}

func (r *ExportGeneric) writeGenericRows(w *csv.Writer, args ExportGenericArgs, filters map[string]any, shouldStop func() bool) error {
	if args.Table == "" {
		return errors.New("table is required")
	}
	if len(args.Columns) == 0 {
		return errors.New("columns is required")
	}

	const chunkSize = 1000
	page := 0
	applySoftDeleteFilter := true

	for {
		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
		}

		query := facades.Orm().Query().Table(args.Table).Select(args.Columns...)
		if applySoftDeleteFilter {
			query = query.Where("deleted_at IS NULL")
		}
		query = applyGenericExportFilters(query, filters, args.SearchTypes)
		query = query.Offset(page * chunkSize).Limit(chunkSize).Order("id asc")

		var rows []map[string]any
		if err := query.Find(&rows); err != nil {
			// Some legacy tables may not have deleted_at.
			if applySoftDeleteFilter && strings.Contains(strings.ToLower(err.Error()), "unknown column") && strings.Contains(strings.ToLower(err.Error()), "deleted_at") {
				applySoftDeleteFilter = false
				continue
			}
			return fmt.Errorf("query export rows failed: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			record := make([]string, 0, len(args.Columns))
			for _, col := range args.Columns {
				record = append(record, formatGenericExportValue(row[col]))
			}
			if err := w.Write(record); err != nil {
				return fmt.Errorf("write csv row failed: %w", err)
			}
		}

		if len(rows) < chunkSize {
			break
		}
		page++
	}

	return nil
}

func applyGenericExportFilters(query orm.Query, filters map[string]any, searchTypes map[string]string) orm.Query {
	for key, value := range filters {
		if value == nil {
			continue
		}
		if key == "_timezone" || key == "page" || key == "page_size" || key == "order_by" {
			continue
		}

		valueStr := cast.ToString(value)
		if strings.TrimSpace(valueStr) == "" {
			continue
		}

		if strings.HasSuffix(key, "_start") {
			field := strings.TrimSuffix(key, "_start")
			query = query.Where(field+" >= ?", value)
			continue
		}
		if strings.HasSuffix(key, "_end") {
			field := strings.TrimSuffix(key, "_end")
			query = query.Where(field+" <= ?", value)
			continue
		}

		switch searchTypes[key] {
		case "like":
			query = query.Where(key+" LIKE ?", "%"+valueStr+"%")
		case "in":
			items := strings.Split(valueStr, ",")
			query = query.Where(key+" IN ?", items)
		case ">", ">=", "<", "<=", "!=":
			query = query.Where(key+" "+searchTypes[key]+" ?", value)
		default:
			query = query.Where(key+" = ?", value)
		}
	}

	return query
}

func formatGenericExportValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case time.Time:
		if v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02 15:04:05")
	case []byte:
		return string(v)
	default:
		return cast.ToString(v)
	}
}
