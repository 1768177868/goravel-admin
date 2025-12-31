package services

import (
	"context"

	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

type UserService interface {
	// GetByID 根据ID获取用户
	GetByID(id uint) (*models.User, error)
	// GetList 获取用户列表
	GetList(filters UserFilters, page, pageSize int) ([]models.User, int64, error)
	// Create 创建用户
	Create(user *models.User) error
	// CreateWithValidation 创建用户（包含验证、密码加密、默认货币设置）
	CreateWithValidation(username, password, nickname, email, phone string, status uint8) (*models.User, error)
	// Update 更新用户
	Update(id uint, user *models.User) error
	// Delete 删除用户（软删除）
	Delete(id uint) error
	// UpdateBalance 更新用户余额（同时创建余额变动记录）
	UpdateBalance(userID uint, amount float64, logType string, source string, sourceID *uint, description string, operatorID *uint, remark string) error
	// ResetPassword 重置用户密码
	ResetPassword(userID uint, newPassword string) error
	// ValidateUserExists 验证用户是否存在（用户名、邮箱、手机号）
	ValidateUserExists(username, email, phone string, excludeID uint) error
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
		return nil, apperrors.ErrUserNotFound.WithError(err)
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
		query = query.Where("username LIKE ?", "%"+filters.Username+"%")
	}
	if filters.Email != "" {
		query = query.Where("email LIKE ?", "%"+filters.Email+"%")
	}
	if filters.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+filters.Phone+"%")
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
	// 使用 map 创建，确保 Status 为 0 时也能正确保存（与管理员创建方式一致）
	// GORM 在处理结构体时可能会忽略零值字段，使用 map 可以确保所有字段都被保存
	userData := map[string]any{
		"username":      user.Username,
		"password":      user.Password,
		"nickname":      user.Nickname,
		"avatar":        user.Avatar,
		"email":         user.Email,
		"phone":         user.Phone,
		"balance":       user.Balance,
		"currency_id":   user.CurrencyID,
		"status":        user.Status,
		"last_login_at": user.LastLoginAt,
	}
	if err := facades.Orm().Query().Table("users").Create(userData); err != nil {
		return err
	}
	// 将创建后的 ID 赋值回 user 对象（GORM 会将生成的 ID 填充到 map 中）
	if id, ok := userData["id"].(uint); ok {
		user.ID = id
	} else if id, ok := userData["id"].(uint64); ok {
		user.ID = uint(id)
	} else {
		// 如果 map 中没有 ID，通过用户名查询获取（与管理员创建方式一致）
		var createdUser models.User
		if err := facades.Orm().Query().Where("username", user.Username).First(&createdUser); err == nil {
			user.ID = createdUser.ID
		}
	}
	return nil
}

// Update 更新用户
func (s *UserServiceImpl) Update(id uint, user *models.User) error {
	// 如果密码为空，则不更新密码字段
	updateData := map[string]any{
		"nickname":    user.Nickname,
		"avatar":      user.Avatar,
		"email":       user.Email,
		"phone":       user.Phone,
		"status":      user.Status,
		"currency_id": user.CurrencyID,
	}

	// 只有用户名不为空时才更新用户名（通常用户名不允许修改，但保留此逻辑以防需要）
	if user.Username != "" {
		updateData["username"] = user.Username
	}

	// 只有密码不为空时才更新密码
	if user.Password != "" {
		updateData["password"] = user.Password
	}

	_, err := facades.Orm().Query().Model(&models.User{}).Where("id", id).Update(updateData)
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
			return apperrors.ErrInsufficientBalance.WithParams(map[string]any{
				"balance": user.Balance,
			})
		}
	default:
		return apperrors.ErrInvalidBalanceType.WithParams(map[string]any{
			"type": logType,
		})
	}

	// 1. 更新用户余额
	_, err = facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update(map[string]any{
		"balance": newBalance,
	})
	if err != nil {
		errorlog.Record(context.Background(), "user", "更新用户余额失败", map[string]any{
			"user_id":     userID,
			"amount":      amount,
			"log_type":    logType,
			"new_balance": newBalance,
			"error":       err.Error(),
		}, "更新用户余额失败: %v", err)
		return apperrors.ErrUpdateFailed.WithError(err)
	}

	// 2. 创建余额变动记录（使用 GORM Sharding）
	_, err = s.balanceLogService.CreateLog(userID, logType, amount, newBalance, source, sourceID, description, operatorID, "success", remark)
	if err != nil {
		// 如果创建记录失败，尝试回滚余额更新
		// 注意：由于涉及分表，无法使用跨表事务，这里手动回滚
		rollbackErr := s.rollbackBalance(userID, user.Balance)
		if rollbackErr != nil {
			errorlog.Record(context.Background(), "user", "回滚余额失败", map[string]any{
				"user_id":      userID,
				"old_balance":  user.Balance,
				"new_balance":  newBalance,
				"rollback_err": rollbackErr.Error(),
			}, "回滚余额失败: %v", rollbackErr)
		}

		errorlog.Record(context.Background(), "user", "创建余额变动记录失败", map[string]any{
			"user_id":     userID,
			"amount":      amount,
			"log_type":    logType,
			"new_balance": newBalance,
			"error":       err.Error(),
		}, "创建余额变动记录失败: %v", err)
		return apperrors.ErrCreateFailed.WithError(err)
	}

	return nil
}

// rollbackBalance 回滚余额到指定值
func (s *UserServiceImpl) rollbackBalance(userID uint, balance float64) error {
	_, err := facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update(map[string]any{
		"balance": balance,
	})
	return err
}

// ValidateUserExists 验证用户是否存在（用户名、邮箱、手机号）
func (s *UserServiceImpl) ValidateUserExists(username, email, phone string, excludeID uint) error {
	// 检查用户名是否已存在
	if username != "" {
		query := facades.Orm().Query().Model(&models.User{}).Where("username", username)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		exists, err := query.Exists()
		if err != nil {
			return apperrors.ErrCreateFailed.WithError(err)
		}
		if exists {
			return apperrors.ErrUsernameExists
		}
	}

	// 检查邮箱是否已存在（如果提供了邮箱）
	if email != "" {
		query := facades.Orm().Query().Model(&models.User{}).Where("email", email)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		exists, err := query.Exists()
		if err != nil {
			return apperrors.ErrCreateFailed.WithError(err)
		}
		if exists {
			return apperrors.NewBusinessError("email_already_exists", "邮箱已存在")
		}
	}

	// 检查手机号是否已存在（如果提供了手机号）
	if phone != "" {
		query := facades.Orm().Query().Model(&models.User{}).Where("phone", phone)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		exists, err := query.Exists()
		if err != nil {
			return apperrors.ErrCreateFailed.WithError(err)
		}
		if exists {
			return apperrors.NewBusinessError("phone_already_exists", "手机号已存在")
		}
	}

	return nil
}

// CreateWithValidation 创建用户（包含验证、密码加密、默认货币设置）
func (s *UserServiceImpl) CreateWithValidation(username, password, nickname, email, phone string, status uint8) (*models.User, error) {
	// 验证用户是否存在
	if err := s.ValidateUserExists(username, email, phone, 0); err != nil {
		return nil, err
	}

	// 密码加密
	hashedPassword, err := facades.Hash().Make(password)
	if err != nil {
		return nil, apperrors.NewBusinessError("password_encrypt_failed", "密码加密失败").WithError(err)
	}

	// 如果未设置货币ID，默认使用人民币
	var currencyID uint
	var cnyCurrency models.Currency
	if err := facades.Orm().Query().Where("code", "CNY").First(&cnyCurrency); err == nil {
		currencyID = cnyCurrency.ID
	}

	// 创建用户
	user := &models.User{
		Username:   username,
		Password:   hashedPassword,
		Nickname:   nickname,
		Avatar:     "",
		Email:      email,
		Phone:      phone,
		Balance:    0,
		CurrencyID: currencyID,
		Status:     status,
	}

	if err := s.Create(user); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	// 查询创建后的用户（确保获取完整信息）
	createdUser, err := s.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// ResetPassword 重置用户密码
func (s *UserServiceImpl) ResetPassword(userID uint, newPassword string) error {
	// 检查用户是否存在
	_, err := s.GetByID(userID)
	if err != nil {
		return err
	}

	// 密码加密
	hashedPassword, err := facades.Hash().Make(newPassword)
	if err != nil {
		return apperrors.ErrPasswordEncryptFailed.WithError(err)
	}

	// 更新密码
	_, err = facades.Orm().Query().Model(&models.User{}).Where("id", userID).Update(map[string]any{
		"password": hashedPassword,
	})
	if err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}

	return nil
}
