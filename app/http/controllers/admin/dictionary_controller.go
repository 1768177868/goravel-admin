package admin

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
)

type DictionaryController struct {
}

func NewDictionaryController() *DictionaryController {
	return &DictionaryController{}
}

// findDictionaryByID 根据ID查找字典，如果不存在则返回错误响应
func (r *DictionaryController) findDictionaryByID(ctx http.Context, id uint) (*models.Dictionary, http.Response) {
	return response.FindByID[models.Dictionary](ctx, id, &response.FindByIDOptions{
		NotFoundMessageKey: "dictionary_not_found",
	})
}

// buildQuery 构建字典查询
func (r *DictionaryController) buildQuery(ctx http.Context) orm.Query {
	dictType := ctx.Request().Query("type", "")
	status := ctx.Request().Query("status", "")
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	query := facades.Orm().Query().Model(&models.Dictionary{})

	if dictType != "" {
		query = query.Where("type LIKE ?", "%"+dictType+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 sort asc, id desc
	query = helpers.ApplySort(query, orderBy, "sort:asc,id:desc")

	return query
}

// Index 字典列表
func (r *DictionaryController) Index(ctx http.Context) http.Response {
	query := r.buildQuery(ctx)
	var dictionaries []models.Dictionary
	return response.PaginateQuery(ctx, query, &dictionaries, nil)
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
		return response.Error(ctx, http.StatusBadRequest, "dictionary_type_required")
	}

	var dictionaries []models.Dictionary
	if err := facades.Orm().Query().Where("type", dictType).Where("status", 1).Order("sort asc, id asc").Get(&dictionaries); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, http.Json{
		"dictionaries": dictionaries,
	})
}
