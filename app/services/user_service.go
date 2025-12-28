package services

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/utils"
)

type UserService interface {
	// GetByID 根据ID获取用户
	GetByID(id uint) (*models.User, error)
	// GetList 获取用户列表
	GetList(filters UserFilters, page, pageSize int) ([]models.User, int64, error)
	// Create 创建用户
	Create(user *models.User) error
	// Update 更新用户
	Update(id uint, user *models.User) error
	// Delete 删除用户（软删除）
	Delete(id uint) error
	// UpdateBalance 更新用户余额（同时创建余额变动记录）
	UpdateBalance(userID uint, amount float64, logType string, source string, sourceID *uint, description string, operatorID *uint, remark string) error
}

type UserFilters struct {
	Username string
	Email    string
	Phone    string
	Status   string
}

type UserServiceImpl struct {
	balanceLogService UserBalanceLogService
}

func NewUserService() UserService {
	return &UserServiceImpl{
		balanceLogService: NewUserBalanceLogService(),
	}
}

// GetByID 根据ID获取用户
func (s *UserServiceImpl) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := facades.Orm().Query().Where("id", id).First(&user); err != nil {
		return nil, fmt.Errorf("用户不存在: %v", err)
	}

	// 加载货币信息
	if user.CurrencyID > 0 {
		var currency models.Currency
		if err := facades.Orm().Query().Where("id", user.CurrencyID).First(&currency); err == nil {
			user.Currency = &currency
		}
	}

	// 格式化余额
	user.Balance = utils.FormatBalance(user.Balance, user.Currency)
	return &user, nil
}

// GetList 获取用户列表
func (s *UserServiceImpl) GetList(filters UserFilters, page, pageSize int) ([]models.User, int64, error) {
	query := facades.Orm().Query().Model(&models.User{})

	if filters.Username != "" {
		query.Where("username", filters.Username)
	}
	if filters.Email != "" {
		query.Where("email", filters.Email)
	}
	if filters.Phone != "" {
		query.Where("phone", filters.Phone)
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var users []models.User
	err = query.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users)
	if err != nil {
		return nil, 0, err
	}

	// 加载货币信息并格式化每个用户的余额
	for i := range users {
		if users[i].CurrencyID > 0 {
			var currency models.Currency
			if err := facades.Orm().Query().Where("id", users[i].CurrencyID).First(&currency); err == nil {
				users[i].Currency = &currency
			}
		}
		users[i].Balance = utils.FormatBalance(users[i].Balance, users[i].Currency)
	}

	return users, total, nil
}

// Create 创建用户
func (s *UserServiceImpl) Create(user *models.User) error {
	// 如果未设置货币ID，默认使用人民币
	if user.CurrencyID == 0 {
		var cnyCurrency models.Currency
		if err := facades.Orm().Query().Where("code", "CNY").First(&cnyCurrency); err == nil {
			user.CurrencyID = cnyCurrency.ID
		}
	}
	return facades.Orm().Query().Create(user)
}

// Update 更新用户
func (s *UserServiceImpl) Update(id uint, user *models.User) error {
	_, err := facades.Orm().Query().Where("id", id).Update(user)
	return err
}

// Delete 删除用户（软删除）
func (s *UserServiceImpl) Delete(id uint) error {
	_, err := facades.Orm().Query().Where("id", id).Delete(&models.User{})
	return err
}

// UpdateBalance 更新用户余额（同时创建余额变动记录）
func (s *UserServiceImpl) UpdateBalance(userID uint, amount float64, logType string, source string, sourceID *uint, description string, operatorID *uint, remark string) error {
	// 获取当前用户
	user, err := s.GetByID(userID)
	if err != nil {
		return err
	}

	// 计算新余额
	var newBalance float64
	switch logType {
	case "income", "refund":
		newBalance = user.Balance + amount
	case "expense":
		newBalance = user.Balance - amount
		if newBalance < 0 {
			return fmt.Errorf("余额不足，当前余额: %.2f", user.Balance)
		}
	default:
		return fmt.Errorf("无效的变动类型: %s", logType)
	}

	// 1. 更新用户余额
	_, err = facades.Orm().Query().Where("id", userID).Update(map[string]interface{}{
		"balance": newBalance,
	})
	if err != nil {
		return fmt.Errorf("更新用户余额失败: %v", err)
	}

	// 2. 创建余额变动记录（使用 GORM Sharding）
	_, err = s.balanceLogService.CreateLog(userID, logType, amount, newBalance, source, sourceID, description, operatorID, "success", remark)
	if err != nil {
		// 如果创建记录失败，尝试回滚余额（这里简化处理，实际应该使用事务）
		// 注意：Goravel 框架的事务可能需要特殊处理
		return fmt.Errorf("创建余额变动记录失败: %v", err)
	}

	return nil
}
