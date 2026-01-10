package jobs

import (
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/models"
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
			"export_header_id",
			"export_header_username",
			"export_header_nickname",
			"export_header_email",
			"export_header_phone",
			"export_header_balance",
			"export_header_currency",
			"export_header_status",
			"export_header_last_login_at",
			"export_header_created_at",
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
	// 解析筛选条件
	username, _ := utils.GetString(filters, "username")
	nickname, _ := utils.GetString(filters, "nickname")
	email, _ := utils.GetString(filters, "email")
	phone, _ := utils.GetString(filters, "phone")
	status, hasStatus := utils.GetUint(filters, "status")
	orderBy, _ := utils.GetString(filters, "order_by")
	_, direction := ParseOrderBy(orderBy)

	// 用户表是单表，时间范围仅在明确传递时使用（不设默认值）
	var startTime, endTime time.Time
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

	const chunkSize = 2000
	var lastID uint = 0

	for {
		if shouldStop != nil && shouldStop() {
			return ErrExportRecordMissing
		}

		// 每次循环重新构建查询（避免 Clone 问题）
		query := facades.Orm().Query().Model(&models.User{}).With("Currency")

		// 应用筛选条件
		if username != "" {
			query = query.Where("username LIKE ?", "%"+username+"%")
		}
		if nickname != "" {
			query = query.Where("nickname LIKE ?", "%"+nickname+"%")
		}
		if email != "" {
			query = query.Where("email LIKE ?", "%"+email+"%")
		}
		if phone != "" {
			query = query.Where("phone LIKE ?", "%"+phone+"%")
		}
		if hasStatus {
			query = query.Where("status", status)
		}
		if !startTime.IsZero() {
			query = query.Where("created_at >= ?", utils.FormatDateTime(startTime))
		}
		if !endTime.IsZero() {
			query = query.Where("created_at <= ?", utils.FormatDateTime(endTime))
		}

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
			row := r.formatUserRow(user, lang)
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
func (r *ExportUsers) formatUserRow(user models.User, lang string) []string {
	// 状态翻译
	statusKey := "export_user_status_disabled"
	if user.Status == 1 {
		statusKey = "export_user_status_enabled"
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
		lastLoginAt = user.LastLoginAt.Format("2006-01-02 15:04:05")
	}

	// 创建时间
	createdAt := ""
	if user.CreatedAt != nil && !user.CreatedAt.IsZero() {
		createdAt = user.CreatedAt.ToDateTimeString()
	}

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
