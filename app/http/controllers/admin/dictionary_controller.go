package admin

import (
	"context"

	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/apidoc"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type DictionaryController struct {
	dictionaryService services.DictionaryService
}

type DictionaryResponse struct {
	ID             uint   `json:"id" example:"1"`                              // 字典ID
	Type           string `json:"type" example:"order_status"`                 // 字典类型
	Label          string `json:"label" example:"已支付"`                         // 字典标签
	Value          string `json:"value" example:"paid"`                        // 字典值
	TranslationKey string `json:"translation_key" example:"order.status.paid"` // 多语言翻译Key
	Description    string `json:"description" example:"订单支付成功状态"`              // 字典描述
	Status         uint8  `json:"status" enums:"0,1" example:"1"`              // 状态（1-启用，0-禁用）
	Sort           int    `json:"sort" example:"10"`                           // 排序值
	Remark         string `json:"remark" example:"系统默认值"`                      // 备注
	CreatedAt      string `json:"created_at" example:"2024-01-01 00:00:00"`    // 创建时间
	UpdatedAt      string `json:"updated_at" example:"2024-01-01 00:00:00"`    // 更新时间
}

type DictionaryListData struct {
	List []DictionaryResponse `json:"list"`
	apidoc.Pagination
}

type DictionaryListResponse struct {
	apidoc.Success
	Data DictionaryListData `json:"data"`
}

type DictionaryDetailData struct {
	Dictionary DictionaryResponse `json:"dictionary"`
}

type DictionaryDetailResponse struct {
	apidoc.Success
	Data DictionaryDetailData `json:"data"`
}

type DictionaryByTypeData struct {
	Dictionaries []DictionaryResponse `json:"dictionaries"`
}

type DictionaryByTypeResponse struct {
	apidoc.Success
	Data DictionaryByTypeData `json:"data"`
}

type DictionaryTypesData struct {
	Types []string `json:"types"`
}

type DictionaryTypesResponse struct {
	apidoc.Success
	Data DictionaryTypesData `json:"data"`
}

func NewDictionaryController() *DictionaryController {
	return &DictionaryController{
		dictionaryService: services.NewDictionaryService(context.Background()),
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
	startTime := getTimeQueryUTC(ctx, "start_time")
	endTime := getTimeQueryUTC(ctx, "end_time")
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
// @Summary      获取字典列表
// @Description  分页获取字典列表，支持按类型、状态、时间范围筛选
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        page        query     int     false  "页码（从1开始）" default(1)
// @Param        page_size   query     int     false  "每页数量（建议 10-100）" default(10)
// @Param        type        query     string  false  "字典类型（精确匹配）"
// @Param        status      query     string  false  "状态（1-启用，0-禁用）" Enums(0,1)
// @Param        start_time  query     string  false  "开始时间（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        end_time    query     string  false  "结束时间（格式：YYYY-MM-DD HH:mm:ss）"
// @Param        order_by    query     string  false  "排序（格式：字段:asc/desc，例如：created_at:desc）"
// @Success      200         {object}  DictionaryListResponse
// @Failure      500         {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/dictionaries [get]
// @Security     BearerAuth
func (r *DictionaryController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

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
// @Summary      获取字典详情
// @Description  根据ID获取字典详情
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id      path      int  true  "字典ID"
// @Success      200     {object}  DictionaryDetailResponse
// @Failure      404     {object}  apidoc.Error "字典不存在"
// @Router       /api/admin/dictionaries/{id} [get]
// @Security     BearerAuth
func (r *DictionaryController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	dictionary, resp := r.findDictionaryByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"dictionary": *dictionary,
	})
}

// Store 创建字典
// @Summary      创建字典
// @Description  创建新的字典项
// @Description  字段说明：type-字典类型（必填）；label-字典标签（必填）；value-字典值（必填）；translation_key-多语言Key；description-描述；status-状态（1启用/0禁用）；sort-排序值；remark-备注
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        request  body      adminrequests.DictionaryCreate  true  "创建参数"
// @Success      200      {object}  DictionaryDetailResponse
// @Failure      400      {object}  apidoc.Error "参数错误"
// @Failure      500      {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/dictionaries [post]
// @Security     BearerAuth
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

	dictionary, err := r.dictionaryService.Create(
		dictionaryCreate.Type,
		dictionaryCreate.Label,
		dictionaryCreate.Value,
		dictionaryCreate.TranslationKey,
		dictionaryCreate.Description,
		dictionaryCreate.Remark,
		dictionaryCreate.Status,
		dictionaryCreate.Sort,
	)
	if err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"type":  dictionaryCreate.Type,
			"label": dictionaryCreate.Label,
		})
	}

	return response.Success(ctx, http.Json{
		"dictionary": dictionary,
	})
}

// @Summary      更新字典
// @Description  根据ID更新字典信息
// @Description  字段说明：type-字典类型；label-字典标签；value-字典值；translation_key-多语言Key；description-描述；status-状态（1启用/0禁用）；sort-排序值；remark-备注（均可选）
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                              true  "字典ID"
// @Param        request  body      adminrequests.DictionaryUpdate   true  "更新参数"
// @Success      200      {object}  DictionaryDetailResponse
// @Failure      400      {object}  apidoc.Error "参数错误"
// @Failure      404      {object}  apidoc.Error "字典不存在"
// @Failure      500      {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/dictionaries/{id} [put]
// @Security     BearerAuth
func (r *DictionaryController) Update(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
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
	if _, exists := allInputs["translation_key"]; exists {
		dictionary.TranslationKey = dictionaryUpdate.TranslationKey
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

	if err := r.dictionaryService.Update(dictionary); err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"dictionary_id": dictionary.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"dictionary": *dictionary,
	})
}

// Destroy 删除字典
// @Summary      删除字典
// @Description  根据ID删除字典项
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "字典ID"
// @Success      200   {object}  apidoc.Success
// @Failure      404   {object}  apidoc.Error "字典不存在"
// @Failure      500   {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/dictionaries/{id} [delete]
// @Security     BearerAuth
func (r *DictionaryController) Destroy(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	dictionary, resp := r.findDictionaryByID(ctx, id)
	if resp != nil {
		return resp
	}

	if err := r.dictionaryService.Delete(dictionary); err != nil {
		return response.ErrorWithLog(ctx, "dictionary", err, map[string]any{
			"dictionary_id": dictionary.ID,
		})
	}

	return response.Success(ctx)
}

// @Summary      按类型获取字典项
// @Description  根据字典类型获取启用状态的字典列表
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Param        type   path      string  true  "字典类型"
// @Success      200    {object}  DictionaryByTypeResponse
// @Failure      400    {object}  apidoc.Error "字典类型不能为空"
// @Failure      500    {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/dictionaries/type/{type} [get]
// @Security     BearerAuth
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

// @Summary      获取全部字典类型
// @Description  获取系统中已配置的全部字典类型列表
// @Tags         字典管理
// @Accept       json
// @Produce      json
// @Success      200    {object}  DictionaryTypesResponse
// @Failure      500    {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/dictionaries/types [get]
// @Security     BearerAuth
func (r *DictionaryController) GetAllTypes(ctx http.Context) http.Response {
	types, err := r.dictionaryService.GetAllTypes()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	return response.Success(ctx, http.Json{
		"types": types,
	})
}
