package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
)

type DictionaryController struct {
}

func NewDictionaryController() *DictionaryController {
	return &DictionaryController{}
}

// Index 字典列表
func (r *DictionaryController) Index(ctx http.Context) http.Response {
	page := cast.ToInt(ctx.Request().Query("page", "1"))
	pageSize := cast.ToInt(ctx.Request().Query("page_size", "10"))
	dictType := ctx.Request().Query("type", "")
	status := ctx.Request().Query("status", "")
	// 使用辅助函数自动转换时区
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

	total, err := query.Count()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	var dictionaries []models.Dictionary
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort asc, id desc").Get(&dictionaries); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Paginate(ctx, "get_success", dictionaries, total, page, pageSize)
}

// Show 字典详情
func (r *DictionaryController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var dictionary models.Dictionary
	if err := facades.Orm().Query().Where("id", id).First(&dictionary); err != nil {
		return response.Error(ctx, http.StatusNotFound, "dictionary_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"dictionary": dictionary,
	})
}

// Store 创建字典
func (r *DictionaryController) Store(ctx http.Context) http.Response {
	dictType := ctx.Request().Input("type")
	label := ctx.Request().Input("label")
	value := ctx.Request().Input("value")
	description := ctx.Request().Input("description")
	status := cast.ToUint8(ctx.Request().Input("status", "0"))
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))
	remark := ctx.Request().Input("remark")

	if dictType == "" || label == "" || value == "" {
		return response.Error(ctx, http.StatusBadRequest, "dictionary_type_label_value_required")
	}

	now := carbon.Now()
	dictionaryData := map[string]interface{}{
		"type":        dictType,
		"label":       label,
		"value":       value,
		"description": description,
		"status":      status, // 明确设置 status，即使是 0 也会被保存
		"sort":        sort,
		"remark":      remark,
		"created_at":  now,
		"updated_at":  now,
	}

	if err := facades.Orm().Query().Table("dictionaries").Create(dictionaryData); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	var dictionary models.Dictionary
	if err := facades.Orm().Query().Where("type", dictType).Where("value", value).First(&dictionary); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return response.Success(ctx, "create_success", http.Json{
		"dictionary": dictionary,
	})
}

// Update 更新字典
func (r *DictionaryController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var dictionary models.Dictionary
	if err := facades.Orm().Query().Where("id", id).First(&dictionary); err != nil {
		return response.Error(ctx, http.StatusNotFound, "dictionary_not_found")
	}

	dictType := ctx.Request().Input("type")
	label := ctx.Request().Input("label")
	value := ctx.Request().Input("value")
	description := ctx.Request().Input("description")
	status := ctx.Request().Input("status", "")
	sort := ctx.Request().Input("sort", "")
	remark := ctx.Request().Input("remark")

	if dictType != "" {
		dictionary.Type = dictType
	}
	if label != "" {
		dictionary.Label = label
	}
	if value != "" {
		dictionary.Value = value
	}
	if description != "" {
		dictionary.Description = description
	}
	if status != "" {
		dictionary.Status = cast.ToUint8(status)
	}
	if sort != "" {
		dictionary.Sort = cast.ToInt(sort)
	}
	if remark != "" {
		dictionary.Remark = remark
	}

	if err := facades.Orm().Query().Save(&dictionary); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"dictionary": dictionary,
	})
}

// Destroy 删除字典
func (r *DictionaryController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var dictionary models.Dictionary
	if err := facades.Orm().Query().Where("id", id).First(&dictionary); err != nil {
		return response.Error(ctx, http.StatusNotFound, "dictionary_not_found")
	}

	if _, err := facades.Orm().Query().Delete(&dictionary); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}

// GetByType 根据类型获取字典
func (r *DictionaryController) GetByType(ctx http.Context) http.Response {
	dictType := ctx.Request().Route("type")
	if dictType == "" {
		return response.Error(ctx, http.StatusBadRequest, "dictionary_type_required")
	}

	var dictionaries []models.Dictionary
	if err := facades.Orm().Query().Where("type", dictType).Where("status", 1).Order("sort asc, id asc").Get(&dictionaries); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, "get_success", http.Json{
		"dictionaries": dictionaries,
	})
}
