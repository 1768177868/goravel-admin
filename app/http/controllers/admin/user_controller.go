package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
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
		return response.Error(ctx, http.StatusNotFound, err.Error())
	}

	return response.Success(ctx, http.Json{
		"user": user,
	})
}

// Store 创建用户
func (r *UserController) Store(ctx http.Context) http.Response {
	var user models.User
	if err := ctx.Request().Bind(&user); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}

	// 密码加密
	if user.Password != "" {
		hashedPassword, err := facades.Hash().Make(user.Password)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "密码加密失败")
		}
		user.Password = hashedPassword
	}

	if err := r.userService.Create(&user); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"user": user,
	})
}

// Update 更新用户
func (r *UserController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var user models.User
	if err := ctx.Request().Bind(&user); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}

	// 如果提供了密码，则加密
	if user.Password != "" {
		hashedPassword, err := facades.Hash().Make(user.Password)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "密码加密失败")
		}
		user.Password = hashedPassword
	} else {
		// 不更新密码
		user.Password = ""
	}

	if err := r.userService.Update(id, &user); err != nil {
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
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"message": "删除成功",
	})
}

// UpdateBalance 更新用户余额
func (r *UserController) UpdateBalance(ctx http.Context) http.Response {
	userID := cast.ToUint(ctx.Request().Input("user_id", "0"))
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
		return response.Error(ctx, http.StatusBadRequest, "user_id 不能为空")
	}
	if amount == 0 {
		return response.Error(ctx, http.StatusBadRequest, "amount 不能为0")
	}
	if logType == "" {
		return response.Error(ctx, http.StatusBadRequest, "type 不能为空")
	}

	if err := r.userService.UpdateBalance(userID, amount, logType, source, sourceID, description, operatorID, remark); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"message": "余额更新成功",
	})
}
