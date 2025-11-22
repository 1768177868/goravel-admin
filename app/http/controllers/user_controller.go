package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/requests"
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
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"users": users,
	})
}

func (r *UserController) Show(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Store(ctx http.Context) http.Response {
	var userCreate requests.UserCreate
	errors, err := ctx.Request().ValidateRequest(&userCreate)
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": err.Error(),
		})
	}
	if errors != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"message": errors.All(),
		})
	}

	// 如果提供了密码，则加密密码
	hashedPassword := ""
	if userCreate.Password != "" {
		var err error
		hashedPassword, err = facades.Hash().Make(userCreate.Password)
		if err != nil {
			return ctx.Response().Json(http.StatusBadRequest, http.Json{
				"error": "密码加密失败",
			})
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
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Update(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": "用户不存在",
		})
	}

	// 更新字段
	updateData := make(map[string]interface{})
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
			return ctx.Response().Json(http.StatusBadRequest, http.Json{
				"error": "密码加密失败",
			})
		}
		updateData["password"] = hashedPassword
	}

	if len(updateData) > 0 {
		if _, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Update(&models.User{}, updateData); err != nil {
			return ctx.Response().Json(http.StatusBadRequest, http.Json{
				"error": err.Error(),
			})
		}
	}

	// 重新查询用户
	if err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).First(&user); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"user": user,
	})
}

func (r *UserController) Destroy(ctx http.Context) http.Response {
	result, err := facades.Orm().Query().Where("id", ctx.Request().Input("id")).Delete(&models.User{})
	if err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": err.Error(),
		})
	}

	return ctx.Response().Success().Json(http.Json{
		"rows_affected": result.RowsAffected,
	})
}
