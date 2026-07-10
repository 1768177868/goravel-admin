package admin

import (
	"encoding/json"
	"fmt"
	appfacades "goravel/app/facades"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/jobs"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils"
)

type UserController struct {}


func NewUserController() *UserController {
	return &UserController{}
}

func (r *UserController) userService(ctx http.Context) services.UserService {
	return services.NewUserService(ctx)
}


// Index 用户列表
func (r *UserController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := services.UserFilters{
		Username: ctx.Request().Query("username", ""),
		Email:    ctx.Request().Query("email", ""),
		Phone:    ctx.Request().Query("phone", ""),
		Status:   ctx.Request().Query("status", ""),
	}

	users, total, err := r.userService(ctx).GetList(filters, page, pageSize)
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
	id := helpers.GetUintRoute(ctx, "id")
	user, err := r.userService(ctx).GetByID(id)
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
	user, err := r.userService(ctx).CreateWithValidation(
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
	id := helpers.GetUintRoute(ctx, "id")

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
	if err := r.userService(ctx).ValidateUserExists("", userUpdate.Email, userUpdate.Phone, id); err != nil {
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

	if err := r.userService(ctx).Update(id, &user); err != nil {
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
	id := helpers.GetUintRoute(ctx, "id")
	if err := r.userService(ctx).Delete(id); err != nil {
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
	userID := helpers.GetUintRoute(ctx, "id")
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

	if err := r.userService(ctx).UpdateBalance(userID, amount, logType, source, sourceID, description, operatorID, remark); err != nil {
		// response.Error 会自动检测 BusinessError 并处理占位符替换
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	return response.Success(ctx, "balance_update_success", http.Json{})
}

// ResetPassword 重置用户密码（管理员操作）
func (r *UserController) ResetPassword(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")

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
	if err := r.userService(ctx).ResetPassword(id, resetPasswordRequest.Password); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.ErrorWithLog(ctx, "password", err, map[string]any{
			"user_id": id,
		})
	}

	return response.Success(ctx, "password_reset_success", http.Json{})
}

// Export 导出用户列表
func (r *UserController) Export(ctx http.Context) http.Response {
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	// 防重复点击
	lockKey := fmt.Sprintf("export:users:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)

	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "already_queued")
	}

	// 获取存储驱动配置
	disk := utils.GetConfigValue(ctx, "storage", "file_disk", "")
	if disk == "" {
		disk = utils.GetConfigValue(ctx, "storage", "export_disk", "")
	}
	if disk == "" {
		disk = "local"
	}

	exportRecord := models.Export{
		AdminID: adminID,
		Type:    models.ExportTypeUsers,
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "",
	}
	if err := appfacades.OrmQuery(ctx).Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	// 构建筛选条件（POST 请求，从 body 获取参数）
	filtersMap := map[string]any{
		"username": ctx.Request().Input("username", ""),
		"nickname": ctx.Request().Input("nickname", ""),
		"email":    ctx.Request().Input("email", ""),
		"phone":    ctx.Request().Input("phone", ""),
		"order_by": ctx.Request().Input("order_by", "id:desc"),
	}
	if statusStr := ctx.Request().Input("status", ""); statusStr != "" {
		filtersMap["status"] = cast.ToUint(statusStr)
	}

	lang := utils.GetCurrentLanguage(ctx)
	timezone := helpers.GetCurrentTimezone(ctx)

	exportArgsStruct := jobs.ExportUsersArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "users",
		Language: lang,
		Timezone: timezone,
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		facades.Log().Errorf("序列化导出参数失败: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	facades.Log().Infof("提交用户导出任务到队列: export_id=%d", exportRecord.ID)

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportUsers{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		facades.Log().Errorf("提交导出任务失败: export_id=%d, error=%v", exportRecord.ID, err)
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		appfacades.OrmQuery(ctx).Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   trans.Get(ctx, "queued"),
	})
}
