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

// ExportUsersArgs 导出用户任务的参数（类型别名）
type ExportUsersArgs = ExportArgs

// ExportUsers 用户导出任务
type ExportUsers struct{}

func (r *ExportUsers) Signature() string {
	return "export_users"
}

func (r *ExportUsers) Handle(args ...any) (retErr error) {
	var exportID uint

	defer func() {
		if rec := recover(); rec != nil {
			errorMsg := fmt.Sprintf("panic: %v", rec)
			facades.Log().Errorf("ExportUsers Job panic: %v", rec)
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
		FilePrefix: "users",
		HeaderKeys: []string{
			"id",
			"username",
			"nickname",
			"email",
			"phone",
			"balance",
			"currency",
			"status",
			"last_login_at",
			"created_at",
		},
		WriteData: r.writeUsersToCSV,
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

// writeUsersToCSV 写入用户数据到 CSV（单表，无分表）
func (r *ExportUsers) writeUsersToCSV(w *csv.Writer, filters map[string]any, lang string, shouldStop func() bool) error {
	// 构建筛选条件（自动填充，无需手动逐字段赋值）
	var userFilters services.UserFilters
	utils.FillFiltersFromMap(filters, &userFilters)

	orderBy, _ := utils.GetString(filters, "order_by")
	_, direction := ParseOrderBy(orderBy)

	// 获取时区（用于时间格式化）
	timezone, _ := utils.GetString(filters, "_timezone")

	const chunkSize = 2000
	var lastID uint = 0

	for {
		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
		}

		// 使用通用查询构建（复用 UserService 的逻辑）
		query := services.BuildUserQuery(userFilters).With("Currency")

		// Keyset 分页
		if lastID > 0 {
			if direction == "desc" {
				query = query.Where("id < ?", lastID)
			} else {
				query = query.Where("id > ?", lastID)
			}
		}

		if direction == "desc" {
			query = query.OrderByDesc("id")
		} else {
			query = query.OrderBy("id")
		}

		var users []models.User
		if err := query.Limit(chunkSize).Get(&users); err != nil {
			return fmt.Errorf("查询用户失败: %v", err)
		}

		if len(users) == 0 {
			break
		}

		// 写入 CSV
		for _, user := range users {
			row := r.formatUserRow(user, lang, timezone)
			if err := w.Write(row); err != nil {
				return fmt.Errorf("写入CSV失败: %v", err)
			}
		}

		// 更新游标
		lastID = users[len(users)-1].ID

		if len(users) < chunkSize {
			break
		}
	}

	return nil
}

// formatUserRow 格式化用户行数据
func (r *ExportUsers) formatUserRow(user models.User, lang, timezone string) []string {
	// 状态翻译
	statusKey := "disabled"
	if user.Status == 1 {
		statusKey = "enabled"
	}
	statusText := utils.TranslateKey(statusKey, lang, cast.ToString(user.Status))

	// 货币名称
	currencyName := ""
	if user.Currency != nil {
		currencyName = user.Currency.Name
	}

	// 最后登录时间
	lastLoginAt := ""
	if user.LastLoginAt != nil {
		lastLoginAt = FormatTimeWithTimezone(*user.LastLoginAt, timezone)
	}

	// 创建时间
	createdAt := FormatCarbonWithTimezone(user.CreatedAt, timezone)

	return []string{
		cast.ToString(user.ID),
		user.Username,
		user.Nickname,
		user.Email,
		user.Phone,
		fmt.Sprintf("%.8f", user.Balance),
		currencyName,
		statusText,
		lastLoginAt,
		createdAt,
	}
}
