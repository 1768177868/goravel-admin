package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type UserController struct {
	userService services.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: services.NewUserService(),
	}
}

// Index 用户列表
func (r *UserController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "20"))

	filters := services.UserFilters{
		Username: ctx.Request().Query("username", ""),
		Email:    ctx.Request().Query("email", ""),
		Phone:    ctx.Request().Query("phone", ""),
		Status:   ctx.Request().Query("status", ""),
	}

	users, total, err := r.userService.GetList(filters, page, pageSize)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 用户详情
func (r *UserController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	user, err := r.userService.GetByID(id)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusNotFound, businessErr.Code)
		}
		return response.Error(ctx, http.StatusNotFound, err.Error())
	}

	return response.Success(ctx, http.Json{
		"user": user,
	})
}

// Store 创建用户
func (r *UserController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var userCreate adminrequests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 检查用户名是否已存在（虽然验证规则中有 not_exists，但这里显式检查以提供更好的错误处理）
	exists, err := facades.Orm().Query().Model(&models.User{}).Where("username", userCreate.Username).Exists()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}
	if exists {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrUsernameExists.Code)
	}

	// 密码加密
	hashedPassword, err := facades.Hash().Make(userCreate.Password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	// 如果未设置货币ID，默认使用人民币
	var currencyID uint
	var cnyCurrency models.Currency
	if err := facades.Orm().Query().Where("code", "CNY").First(&cnyCurrency); err == nil {
		currencyID = cnyCurrency.ID
	}

	now := carbon.Now()
	userData := map[string]any{
		"username":    userCreate.Username,
		"password":    hashedPassword,
		"nickname":    userCreate.Nickname,
		"avatar":      "",
		"email":       userCreate.Email,
		"phone":       userCreate.Phone,
		"balance":     0,
		"currency_id": currencyID,
		"status":      userCreate.Status,
		"created_at":  now,
		"updated_at":  now,
	}

	var user models.User
	if err := facades.Orm().Query().Table("users").Create(userData); err != nil {
		return response.ErrorWithLog(ctx, "user", err, map[string]any{
			"username": userCreate.Username,
		})
	}

	// 查询创建后的用户
	if err := facades.Orm().Query().Where("username", userCreate.Username).First(&user); err != nil {
		return response.ErrorWithLog(ctx, "user", err, map[string]any{
			"username": userCreate.Username,
		})
	}

	return response.Success(ctx, http.Json{
		"user": user,
	})
}

// Update 更新用户
func (r *UserController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))

	// 使用请求验证
	var userUpdate adminrequests.UserUpdate
	errors, err := ctx.Request().ValidateRequest(&userUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 检查邮箱是否已存在（排除当前用户）
	if userUpdate.Email != "" {
		var existingUser models.User
		if err := facades.Orm().Query().Where("email", userUpdate.Email).Where("id", "!=", id).First(&existingUser); err == nil {
			return response.Error(ctx, http.StatusBadRequest, "email_already_exists")
		}
	}

	user := models.User{
		Nickname: userUpdate.Nickname,
		Email:    userUpdate.Email,
		Phone:    userUpdate.Phone,
		Status:   userUpdate.Status,
	}

	// 如果提供了密码，则加密
	if userUpdate.Password != "" {
		hashedPassword, err := facades.Hash().Make(userUpdate.Password)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
		}
		user.Password = hashedPassword
	}

	if err := r.userService.Update(id, &user); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"user": user,
	})
}

// Destroy 删除用户
func (r *UserController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	if err := r.userService.Delete(id); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, "delete_success", http.Json{})
}

// UpdateBalance 更新用户余额
func (r *UserController) UpdateBalance(ctx http.Context) http.Response {
	// 从路由参数获取 user_id
	userID := cast.ToUint(ctx.Request().Route("id"))
	amount := cast.ToFloat64(ctx.Request().Input("amount", "0"))
	logType := ctx.Request().Input("type", "")
	source := ctx.Request().Input("source", "manual")
	description := ctx.Request().Input("description", "")
	remark := ctx.Request().Input("remark", "")

	var sourceID *uint
	if sourceIDStr := ctx.Request().Input("source_id", ""); sourceIDStr != "" {
		id := cast.ToUint(sourceIDStr)
		sourceID = &id
	}

	var operatorID *uint
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err == nil && adminID > 0 {
		operatorID = &adminID
	}

	if userID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "user_id_required")
	}
	if amount == 0 {
		return response.Error(ctx, http.StatusBadRequest, "amount_cannot_be_zero")
	}
	if logType == "" {
		return response.Error(ctx, http.StatusBadRequest, "balance_type_required")
	}

	if err := r.userService.UpdateBalance(userID, amount, logType, source, sourceID, description, operatorID, remark); err != nil {
		// response.Error 会自动检测 BusinessError 并处理占位符替换
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	return response.Success(ctx, "balance_update_success", http.Json{})
}
