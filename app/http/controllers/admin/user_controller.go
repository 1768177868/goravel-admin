package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
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

	// 使用服务方法创建用户（包含验证、密码加密、默认货币设置）
	user, err := r.userService.CreateWithValidation(
		userCreate.Username,
		userCreate.Password,
		userCreate.Nickname,
		userCreate.Email,
		userCreate.Phone,
		userCreate.Status,
	)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
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

	// 使用服务方法验证用户是否存在（排除当前用户）
	if err := r.userService.ValidateUserExists("", userUpdate.Email, userUpdate.Phone, id); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
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

// ResetPassword 重置用户密码（管理员操作）
func (r *UserController) ResetPassword(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))

	// 使用请求验证
	var resetPasswordRequest adminrequests.ResetPassword
	errors, err := ctx.Request().ValidateRequest(&resetPasswordRequest)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用服务方法重置密码
	if err := r.userService.ResetPassword(id, resetPasswordRequest.Password); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.ErrorWithLog(ctx, "password", err, map[string]any{
			"user_id": id,
		})
	}

	return response.Success(ctx, "password_reset_success", http.Json{})
}
