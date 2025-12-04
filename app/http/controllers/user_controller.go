package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/requests"
	"goravel/app/http/response"
	"goravel/app/models"
)

type UserController struct {
	//Dependent services
}

func NewUserController() *UserController {
	return &UserController{
		//Inject services
	}
}

func (r *UserController) Index(ctx http.Context) http.Response {
	var users []models.User
	if err := facades.Orm().Query().Get(&users); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return ctx.Response().Success().Json(http.Json{
		"users": users,
	})
}

func (r *UserController) Show(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return response.Error(ctx, http.StatusNotFound, "user_not_found")
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Store(ctx http.Context) http.Response {
	var userCreate requests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 如果提供了密码，则加密密码
	hashedPassword := ""
	if userCreate.Password != "" {
		var err error
		hashedPassword, err = facades.Hash().Make(userCreate.Password)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
		}
	}

	// 设置默认状态
	status := userCreate.Status
	if status == 0 {
		status = 1 // 默认为启用
	}

	user := models.User{
		Username: userCreate.Username,
		Password: hashedPassword,
		Name:     userCreate.Name,
		Avatar:   userCreate.Avatar,
		Alias:    userCreate.Alias,
		Mail:     userCreate.Mail,
		Status:   status,
		Tags:     userCreate.Tags,
	}
	if err := facades.Orm().Query().Create(&user); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Update(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return response.Error(ctx, http.StatusNotFound, "user_not_found")
	}

	// 更新字段
	updateData := make(map[string]any)
	if name := ctx.Request().Input("name"); name != "" {
		updateData["name"] = name
	}
	if avatar := ctx.Request().Input("avatar"); avatar != "" {
		updateData["avatar"] = avatar
	}
	if alias := ctx.Request().Input("alias"); alias != "" {
		updateData["alias"] = alias
	}
	if mail := ctx.Request().Input("mail"); mail != "" {
		updateData["mail"] = mail
	}
	if status := ctx.Request().Input("status"); status != "" {
		updateData["status"] = status
	}
	if password := ctx.Request().Input("password"); password != "" {
		hashedPassword, err := facades.Hash().Make(password)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "password_encrypt_failed")
		}
		updateData["password"] = hashedPassword
	}

	if len(updateData) > 0 {
		if _, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Update(&models.User{}, updateData); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "update_failed")
		}
	}

	// 重新查询用户
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return response.Error(ctx, http.StatusNotFound, "user_not_found")
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Destroy(ctx http.Context) http.Response {
	result, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Delete(&models.User{})
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return ctx.Response().Success().Json(http.Json{
		"rows_affected": result.RowsAffected,
	})
}
