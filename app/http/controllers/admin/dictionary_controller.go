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

type DictionaryController struct {
	dictionaryService services.DictionaryService
}

func NewDictionaryController() *DictionaryController {
	return &DictionaryController{
		dictionaryService: services.NewDictionaryService(),
	}
}

// findDictionaryByID 根据ID查找字典，如果不存在则返回错误响应
func (r *DictionaryController) findDictionaryByID(ctx http.Context, id uint) (*models.Dictionary, http.Response) {
	dictionary, err := r.dictionaryService.GetByID(id)
	if err != nil {
		return nil, response.Error(ctx, http.StatusNotFound, apperrors.ErrDictionaryNotFound.Code)
	}
	return dictionary, nil
}

// buildFilters 构建查询过滤器
func (r *DictionaryController) buildFilters(ctx http.Context) services.DictionaryFilters {
	dictType := ctx.Request().Query("type", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")
	orderBy := ctx.Request().Query("order_by", "")

	return services.DictionaryFilters{
		Type:      dictType,
		Status:    status,
		StartTime: startTime,
		EndTime:   endTime,
		OrderBy:   orderBy,
	}
}

// Index 字典列表
func (r *DictionaryController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "20"))

	filters := r.buildFilters(ctx)

	dictionaries, total, err := r.dictionaryService.GetList(filters, page, pageSize)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      dictionaries,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show 字典详情
func (r *DictionaryController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	dictionary, resp := r.findDictionaryByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"dictionary": *dictionary,
	})
}

// Store 创建字典
func (r *DictionaryController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var dictionaryCreate adminrequests.DictionaryCreate
	errors, err := ctx.Request().ValidateRequest(&dictionaryCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	now := carbon.Now()
	dictionaryData := map[string]any{
		"type":        dictionaryCreate.Type,
		"label":       dictionaryCreate.Label,
		"value":       dictionaryCreate.Value,
		"description": dictionaryCreate.Description,
		"status":      dictionaryCreate.Status,
		"sort":        dictionaryCreate.Sort,
		"remark":      dictionaryCreate.Remark,
		"created_at":  now,
		"updated_at":  now,
	}

	if err := facades.Orm().Query().Table("dictionaries").Create(dictionaryData); err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"type":  dictionaryCreate.Type,
			"label": dictionaryCreate.Label,
		})
	}

	var dictionary models.Dictionary
	if err := facades.Orm().Query().Where("type", dictionaryCreate.Type).Where("value", dictionaryCreate.Value).First(&dictionary); err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"type":  dictionaryCreate.Type,
			"value": dictionaryCreate.Value,
		})
	}

	return response.Success(ctx, http.Json{
		"dictionary": dictionary,
	})
}

func (r *DictionaryController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	dictionary, resp := r.findDictionaryByID(ctx, id)
	if resp != nil {
		return resp
	}

	// 使用请求验证
	var dictionaryUpdate adminrequests.DictionaryUpdate
	errors, err := ctx.Request().ValidateRequest(&dictionaryUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用 All() 方法检查字段是否存在
	allInputs := ctx.Request().All()

	if _, exists := allInputs["type"]; exists {
		dictionary.Type = dictionaryUpdate.Type
	}
	if _, exists := allInputs["label"]; exists {
		dictionary.Label = dictionaryUpdate.Label
	}
	if _, exists := allInputs["value"]; exists {
		dictionary.Value = dictionaryUpdate.Value
	}
	if _, exists := allInputs["description"]; exists {
		dictionary.Description = dictionaryUpdate.Description
	}
	if _, exists := allInputs["status"]; exists {
		dictionary.Status = dictionaryUpdate.Status
	}
	if _, exists := allInputs["sort"]; exists {
		dictionary.Sort = dictionaryUpdate.Sort
	}
	if _, exists := allInputs["remark"]; exists {
		dictionary.Remark = dictionaryUpdate.Remark
	}

	if err := facades.Orm().Query().Save(dictionary); err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"dictionary_id": dictionary.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"dictionary": *dictionary,
	})
}

// Destroy 删除字典
func (r *DictionaryController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	dictionary, resp := r.findDictionaryByID(ctx, id)
	if resp != nil {
		return resp
	}

	if _, err := facades.Orm().Query().Delete(dictionary); err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"dictionary_id": dictionary.ID,
		})
	}

	return response.Success(ctx)
}

func (r *DictionaryController) GetByType(ctx http.Context) http.Response {
	dictType := ctx.Request().Route("type")
	if dictType == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrDictionaryTypeRequired.Code)
	}

	dictionaries, err := r.dictionaryService.GetByType(dictType)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	return response.Success(ctx, http.Json{
		"dictionaries": dictionaries,
	})
}
