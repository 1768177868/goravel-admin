package jobs

import (
	"encoding/csv"
	"errors"
	"fmt"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

// ExportArticles 文章异步导出（与列表共用 services.BuildArticleQuery + FillFiltersFromMap）
type ExportArticles struct{}

func (r *ExportArticles) Signature() string {
	return "export_articles"
}

func (r *ExportArticles) Handle(args ...any) (retErr error) {
	var exportID uint

	defer func() {
		if rec := recover(); rec != nil {
			errorMsg := fmt.Sprintf("panic: %v", rec)
			facades.Log().Errorf("ExportArticles Job panic: %v", rec)
			MarkExportFailed(exportID, errorMsg)
			retErr = fmt.Errorf("%s", errorMsg)
		}
	}()

	exportArgs, err := ParseArgs(args...)
	if err != nil {
		return err
	}
	exportID = exportArgs.ExportID

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
		FilePrefix: "articles",
		HeaderKeys: []string{
			"admin_id",
			"title",
			"content",
			"status",
			"created_at",
			"updated_at",
		},
		WriteData: r.writeArticlesToCSV,
	})

	jobErr := exporter.Execute(exportArgs)

	if errors.Is(jobErr, ErrExportRecordMissing) {
		return nil
	}

	if jobErr != nil {
		MarkExportFailed(exportID, jobErr.Error())
		return jobErr
	}

	return nil
}

func (r *ExportArticles) writeArticlesToCSV(w *csv.Writer, filters map[string]any, lang string, shouldStop func() bool) error {
	var articleFilters services.ArticleFilters
	utils.FillFiltersFromMap(filters, &articleFilters)

	timezone, _ := utils.GetString(filters, "_timezone")

	const chunkSize = 2000
	lastID := uint(0)

	for {
		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
		}

		q := services.BuildArticleQuery(articleFilters).With("Admin")
		if lastID > 0 {
			q = q.Where("id < ?", lastID)
		}
		q = q.Order("id desc").Limit(chunkSize)

		var rows []models.Article
		if err := q.Get(&rows); err != nil {
			return fmt.Errorf("query articles for export failed: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		for _, art := range rows {
			row := r.formatArticleRow(art, lang, timezone)
			if err := w.Write(row); err != nil {
				return fmt.Errorf("write csv row failed: %w", err)
			}
		}

		lastID = rows[len(rows)-1].ID
		if len(rows) < chunkSize {
			break
		}
	}

	return nil
}

func (r *ExportArticles) formatArticleRow(art models.Article, lang, timezone string) []string {
	statusKey := "disabled"
	if art.Status == 1 {
		statusKey = "enabled"
	}
	statusText := utils.TranslateKey(statusKey, lang, cast.ToString(art.Status))

	return []string{
		cast.ToString(art.AdminId),
		art.Title,
		art.Content,
		statusText,
		FormatCarbonWithTimezone(art.CreatedAt, timezone),
		FormatCarbonWithTimezone(art.UpdatedAt, timezone),
	}
}
