package admin

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/traceid"
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

	// 密码加密
	hashedPassword, err := facades.Hash().Make(userCreate.Password)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
	}

	user := models.User{
		Username: userCreate.Username,
		Password: hashedPassword,
		Nickname: userCreate.Nickname,
		Email:    userCreate.Email,
		Phone:    userCreate.Phone,
		Status:   userCreate.Status,
	}

	if err := r.userService.Create(&user); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
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
		// 检查是否是业务错误
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			// 获取翻译后的消息
			messageKey := businessErr.Code
			message := trans.Get(ctx, messageKey)

			// 如果有参数，替换参数占位符
			if len(businessErr.Params) > 0 {
				for key, value := range businessErr.Params {
					// 支持 {key} 和 ${key} 格式
					placeholder1 := fmt.Sprintf("{%s}", key)
					placeholder2 := fmt.Sprintf("${%s}", key)

					var valueStr string
					switch v := value.(type) {
					case float64:
						valueStr = fmt.Sprintf("%.2f", v)
					case float32:
						valueStr = fmt.Sprintf("%.2f", v)
					default:
						valueStr = fmt.Sprintf("%v", v)
					}

					message = strings.ReplaceAll(message, placeholder1, valueStr)
					message = strings.ReplaceAll(message, placeholder2, valueStr)
				}
			}

			// 如果翻译后的消息和 key 相同，说明没有找到翻译，使用默认消息
			if message == messageKey {
				message = businessErr.Message
				// 如果默认消息中有参数，也替换（使用 fmt.Sprintf 格式化）
				if len(businessErr.Params) > 0 {
					// 对于余额不足错误，格式化余额值
					if balance, ok := businessErr.Params["balance"]; ok {
						var balanceStr string
						switch v := balance.(type) {
						case float64:
							balanceStr = fmt.Sprintf("%.2f", v)
						case float32:
							balanceStr = fmt.Sprintf("%.2f", v)
						default:
							balanceStr = fmt.Sprintf("%v", v)
						}
						message = fmt.Sprintf("%s，当前余额: %s", message, balanceStr)
					}
					// 对于无效类型错误，格式化类型值
					if typeVal, ok := businessErr.Params["type"]; ok {
						typeStr := fmt.Sprintf("%v", typeVal)
						message = fmt.Sprintf("%s: %s", message, typeStr)
					}
				}
			}

			// 直接构造响应，使用替换后的 message
			responseData := http.Json{
				"code":       http.StatusBadRequest,
				"message":    message,
				"error_code": messageKey,
			}

			// 自动包含 trace_id
			if traceID := traceid.FromHTTPContext(ctx); traceID != "" {
				responseData["trace_id"] = traceID
			}

			return ctx.Response().Json(http.StatusBadRequest, responseData)
		}
		// 如果不是业务错误，返回通用错误
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, "balance_update_success", http.Json{})
}
