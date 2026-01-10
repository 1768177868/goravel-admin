package admin

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/services"
	"goravel/app/utils"
)

type UserBalanceLogController struct {
	balanceLogService services.UserBalanceLogService
}

func NewUserBalanceLogController() *UserBalanceLogController {
	return &UserBalanceLogController{
		balanceLogService: services.NewUserBalanceLogService(),
	}
}

// Index 余额变动记录列表
func (r *UserBalanceLogController) Index(ctx http.Context) http.Response {
	userID := cast.ToUint(ctx.Request().Query("user_id", "0"))
	if userID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "user_id_required_for_sharding")
	}

	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "20"))

	// 解析时间
	startTimeStr := ctx.Request().Query("start_time", "")
	endTimeStr := ctx.Request().Query("end_time", "")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, _ = utils.ParseDateTime(startTimeStr)
	}
	if endTimeStr != "" {
		endTime, _ = utils.ParseDateTime(endTimeStr)
	}

	var operatorID *uint
	if operatorIDStr := ctx.Request().Query("operator_id", ""); operatorIDStr != "" {
		id := cast.ToUint(operatorIDStr)
		operatorID = &id
	}

	filters := services.UserBalanceLogFilters{
		UserID:     userID,
		Type:       ctx.Request().Query("type", ""),
		Source:     ctx.Request().Query("source", ""),
		Status:     ctx.Request().Query("status", ""),
		StartTime:  startTime,
		EndTime:    endTime,
		OperatorID: operatorID,
	}

	logs, total, err := r.balanceLogService.GetLogs(filters, page, pageSize)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Statistics 用户余额统计
func (r *UserBalanceLogController) Statistics(ctx http.Context) http.Response {
	userID := cast.ToUint(ctx.Request().Query("user_id", "0"))
	if userID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "user_id_required")
	}

	startTimeStr := ctx.Request().Query("start_time", "")
	endTimeStr := ctx.Request().Query("end_time", "")

	var startTime, endTime time.Time
	if startTimeStr != "" {
		startTime, _ = utils.ParseDateTime(startTimeStr)
	}
	if endTimeStr != "" {
		endTime, _ = utils.ParseDateTime(endTimeStr)
	}

	stats, err := r.balanceLogService.GetUserStatistics(userID, startTime, endTime)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"statistics": stats,
	})
}

// Store 创建余额变动记录
func (r *UserBalanceLogController) Store(ctx http.Context) http.Response {
	userID := cast.ToUint(ctx.Request().Input("user_id", "0"))
	if userID == 0 {
		return response.Error(ctx, http.StatusBadRequest, "user_id_required_for_sharding")
	}

	logType := ctx.Request().Input("type", "")
	if logType == "" {
		return response.Error(ctx, http.StatusBadRequest, "balance_type_required")
	}

	amount := cast.ToFloat64(ctx.Request().Input("amount", "0"))
	if amount == 0 {
		return response.Error(ctx, http.StatusBadRequest, "amount_cannot_be_zero")
	}

	balance := cast.ToFloat64(ctx.Request().Input("balance", "0"))
	source := ctx.Request().Input("source", "manual")
	description := ctx.Request().Input("description", "")
	remark := ctx.Request().Input("remark", "")
	status := ctx.Request().Input("status", "success")

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

	// 如果未提供 balance，从用户表获取当前余额
	if balance == 0 {
		currentBalance, err := r.balanceLogService.GetUserBalance(userID)
		if err != nil {
			// 检查是否是业务错误
			if businessErr, ok := apperrors.GetBusinessError(err); ok {
				return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
			}
			return response.Error(ctx, http.StatusInternalServerError, "get_user_balance_failed")
		}
		balance = currentBalance
	}

	log, err := r.balanceLogService.CreateLog(userID, logType, amount, balance, source, sourceID, description, operatorID, status, remark)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, "balance_log_create_success", http.Json{
		"data": log,
	})
}
