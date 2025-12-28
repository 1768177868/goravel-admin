package services

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

type UserBalanceLogService interface {
	// CreateLog 创建余额变动记录（使用 GORM Sharding，必须提供 user_id）
	CreateLog(userID uint, logType string, amount float64, balance float64, source string, sourceID *uint, description string, operatorID *uint, status string, remark string) (*models.UserBalanceLog, error)
	// GetLogs 查询余额变动记录列表（必须提供 user_id）
	GetLogs(filters UserBalanceLogFilters, page, pageSize int) ([]models.UserBalanceLog, int64, error)
	// GetUserBalance 获取用户当前余额（从 users 表获取）
	GetUserBalance(userID uint) (float64, error)
	// GetUserStatistics 获取用户余额统计
	GetUserStatistics(userID uint, startTime, endTime time.Time) (*UserBalanceStatistics, error)
}

type UserBalanceLogFilters struct {
	UserID     uint      // 用户ID（必填，用于分表路由）
	Type       string    // 变动类型
	Source     string    // 来源
	Status     string    // 状态
	StartTime  time.Time // 开始时间
	EndTime    time.Time // 结束时间
	OperatorID *uint     // 操作员ID
}

type UserBalanceStatistics struct {
	TotalIncome    float64 `json:"total_income"`    // 总收入
	TotalExpense   float64 `json:"total_expense"`   // 总支出
	TotalRefund    float64 `json:"total_refund"`    // 总退款
	CurrentBalance float64 `json:"current_balance"` // 当前余额
}

type UserBalanceLogServiceImpl struct {
}

func NewUserBalanceLogService() UserBalanceLogService {
	return &UserBalanceLogServiceImpl{}
}

// CreateLog 创建余额变动记录
// 注意：必须提供 user_id，GORM Sharding 插件会自动路由到对应分表
func (s *UserBalanceLogServiceImpl) CreateLog(
	userID uint,
	logType string,
	amount float64,
	balance float64,
	source string,
	sourceID *uint,
	description string,
	operatorID *uint,
	status string,
	remark string,
) (*models.UserBalanceLog, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id 不能为空，GORM Sharding 需要 ShardingKey")
	}

	// 默认状态为 success
	if status == "" {
		status = "success"
	}

	log := &models.UserBalanceLog{
		UserID:      userID,
		Type:        logType,
		Amount:      amount,
		Balance:     balance,
		Source:      source,
		SourceID:    sourceID,
		Description: description,
		OperatorID:  operatorID,
		Status:      status,
		Remark:      remark,
	}

	// 使用 GORM Sharding 插件，必须带上 user_id 条件
	// 插件会自动路由到对应的分表
	err := facades.Orm().Query().Create(log)
	if err != nil {
		errorlog.Record(context.Background(), "user-balance-log", "创建余额变动记录失败", map[string]any{
			"user_id":  userID,
			"log_type": logType,
			"amount":   amount,
			"error":    err.Error(),
		}, "创建余额变动记录失败: %v", err)
		return nil, fmt.Errorf("创建余额变动记录失败: %v", err)
	}

	return log, nil
}

// GetLogs 查询余额变动记录列表
// 注意：必须提供 user_id，否则 GORM Sharding 插件会报错
func (s *UserBalanceLogServiceImpl) GetLogs(filters UserBalanceLogFilters, page, pageSize int) ([]models.UserBalanceLog, int64, error) {
	if filters.UserID == 0 {
		return nil, 0, fmt.Errorf("user_id 不能为空，GORM Sharding 需要 ShardingKey")
	}

	// 构建查询，必须包含 user_id（ShardingKey）
	query := facades.Orm().Query().Model(&models.UserBalanceLog{}).
		Where("user_id", filters.UserID)

	// 添加其他筛选条件
	if filters.Type != "" {
		query = query.Where("type", filters.Type)
	}
	if filters.Source != "" {
		query = query.Where("source", filters.Source)
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}
	if !filters.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filters.EndTime)
	}
	if filters.OperatorID != nil && *filters.OperatorID > 0 {
		query = query.Where("operator_id", *filters.OperatorID)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var logs []models.UserBalanceLog
	err = query.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetUserBalance 获取用户当前余额（从 users 表获取）
func (s *UserBalanceLogServiceImpl) GetUserBalance(userID uint) (float64, error) {
	if userID == 0 {
		return 0, fmt.Errorf("user_id 不能为空")
	}

	var user models.User
	err := facades.Orm().Query().Where("id", userID).First(&user)
	if err != nil {
		return 0, fmt.Errorf("用户不存在: %v", err)
	}

	return user.Balance, nil
}

// GetUserStatistics 获取用户余额统计
func (s *UserBalanceLogServiceImpl) GetUserStatistics(userID uint, startTime, endTime time.Time) (*UserBalanceStatistics, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id 不能为空")
	}

	query := facades.Orm().Query().Model(&models.UserBalanceLog{}).
		Where("user_id", userID).
		Where("status", "success")

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	var stats UserBalanceStatistics

	// 统计收入
	var incomeResult struct {
		Total float64
	}
	incomeQuery := query.Where("type", "income")
	err := incomeQuery.Select("COALESCE(SUM(amount), 0) as total").Scan(&incomeResult)
	if err == nil {
		stats.TotalIncome = incomeResult.Total
	}

	// 统计支出
	var expenseResult struct {
		Total float64
	}
	expenseQuery := query.Where("type", "expense")
	err = expenseQuery.Select("COALESCE(SUM(amount), 0) as total").Scan(&expenseResult)
	if err == nil {
		stats.TotalExpense = expenseResult.Total
	}

	// 统计退款
	var refundResult struct {
		Total float64
	}
	refundQuery := query.Where("type", "refund")
	err = refundQuery.Select("COALESCE(SUM(amount), 0) as total").Scan(&refundResult)
	if err == nil {
		stats.TotalRefund = refundResult.Total
	}

	// 获取当前余额
	balance, err := s.GetUserBalance(userID)
	if err == nil {
		// 获取用户信息以获取货币信息
		var user models.User
		if err := facades.Orm().Query().Where("id", userID).First(&user); err == nil {
			// 加载货币信息
			if user.CurrencyID > 0 {
				var currency models.Currency
				if err := facades.Orm().Query().Where("id", user.CurrencyID).First(&currency); err == nil {
					user.Currency = &currency
				}
			}
			stats.CurrentBalance = utils.FormatBalance(balance, user.Currency)
			// 格式化统计金额
			stats.TotalIncome = utils.FormatBalance(stats.TotalIncome, user.Currency)
			stats.TotalExpense = utils.FormatBalance(stats.TotalExpense, user.Currency)
			stats.TotalRefund = utils.FormatBalance(stats.TotalRefund, user.Currency)
		} else {
			stats.CurrentBalance = balance
		}
	}

	return &stats, nil
}
