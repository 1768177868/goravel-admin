package admin

import (
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/apidoc"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type ArticleController struct {
	ArticleService services.ArticleService
}

type ArticleResponse struct {
	ID        uint   `json:"id" example:"1"`                                // 文章ID
	Title     string `json:"title" example:"示例标题"`                        // 文章标题
	Content   string `json:"content" example:"示例内容"`                      // 文章内容
	Status    uint8  `json:"status" enums:"0,1" example:"1"`                // 发布状态（1-发布，0-未发布）
	AdminID   uint   `json:"admin_id" example:"1"`                          // 发布管理员ID
	CreatedAt string `json:"created_at" example:"2024-01-01 00:00:00"`      // 创建时间
	UpdatedAt string `json:"updated_at" example:"2024-01-01 00:00:00"`      // 更新时间
}

type ArticleListData struct {
	List []ArticleResponse `json:"list"`
	apidoc.Pagination
}

type ArticleListResponse struct {
	apidoc.Success
	Data ArticleListData `json:"data"`
}

type ArticleDetailData struct {
	Article ArticleResponse `json:"article"`
}

type ArticleDetailResponse struct {
	apidoc.Success
	Data ArticleDetailData `json:"data"`
}

type ArticleDeleteData struct{}

type ArticleDeleteResponse struct {
	apidoc.Success
	Data ArticleDeleteData `json:"data"`
}

func NewArticleController() *ArticleController {
	return &ArticleController{
		ArticleService: services.NewArticleService(),
	}
}

// Index Article列表
// @Summary      获取文章列表
// @Description  分页获取文章列表，支持按标题、状态、管理员等筛选
// @Tags         内容管理
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "页码（从1开始）" default(1)
// @Param        page_size  query     int     false  "每页数量（建议 10-100）" default(10)
// @Param        title      query     string  false  "标题（模糊匹配）"
// @Param        content    query     string  false  "内容（精确匹配）"
// @Param        status     query     string  false  "状态（0-未发布，1-发布）" Enums(0,1)
// @Param        admin_id   query     string  false  "管理员ID（精确匹配）"
// @Param        created_at query     string  false  "创建时间（精确匹配）"
// @Param        updated_at query     string  false  "更新时间（精确匹配）"
// @Success      200        {object}  ArticleListResponse
// @Failure      500        {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/articles [get]
// @Security     BearerAuth
func (c *ArticleController) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	title := ctx.Request().Query("title", "")
	content := ctx.Request().Query("content", "")
	status := ctx.Request().Query("status", "")
	admin_id := ctx.Request().Query("admin_id", "")
	created_at := ctx.Request().Query("created_at", "")
	updated_at := ctx.Request().Query("updated_at", "")

	filters := services.ArticleFilters{

		Title:     title,
		Content:   content,
		Status:    status,
		AdminId:   admin_id,
		CreatedAt: created_at,
		UpdatedAt: updated_at,
	}

	list, total, err := c.ArticleService.GetList(filters, page, pageSize)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show Article详情
// @Summary      获取文章详情
// @Description  根据ID获取文章详情
// @Tags         内容管理
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "文章ID"
// @Success      200   {object}  ArticleDetailResponse
// @Failure      404   {object}  apidoc.Error "记录不存在"
// @Router       /api/admin/articles/{id} [get]
// @Security     BearerAuth
func (c *ArticleController) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	item, err := c.ArticleService.GetByID(id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})
}

// Store 创建Article
// @Summary      创建文章
// @Description  创建新的文章记录
// @Description  字段说明：title-文章标题（必填）；content-文章内容；status-发布状态（1发布/0未发布）；admin_id-发布管理员ID（必填）
// @Tags         内容管理
// @Accept       json
// @Produce      json
// @Param        request  body      adminrequests.ArticleCreate  true  "创建参数"
// @Success      200      {object}  ArticleDetailResponse
// @Failure      400      {object}  apidoc.Error "参数错误"
// @Failure      500      {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/articles [post]
// @Security     BearerAuth
func (c *ArticleController) Store(ctx http.Context) http.Response {

	var req adminrequests.ArticleCreate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	item, err := c.ArticleService.Create(&req)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})

}

// Update 更新Article
// @Summary      更新文章
// @Description  根据ID更新文章信息
// @Description  字段说明：title-文章标题；content-文章内容；status-发布状态（1发布/0未发布）；admin_id-发布管理员ID（均为可选）
// @Tags         内容管理
// @Accept       json
// @Produce      json
// @Param        id       path      int                         true  "文章ID"
// @Param        request  body      adminrequests.ArticleUpdate true  "更新参数"
// @Success      200      {object}  ArticleDetailResponse
// @Failure      400      {object}  apidoc.Error "参数错误"
// @Failure      500      {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/articles/{id} [put]
// @Security     BearerAuth
func (c *ArticleController) Update(ctx http.Context) http.Response {

	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.ArticleUpdate
	errors, err := ctx.Request().ValidateRequest(&req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	item, err := c.ArticleService.Update(id, &req)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"article": item,
	})

}

// Destroy 删除Article
// @Summary      删除文章
// @Description  根据ID删除文章
// @Tags         内容管理
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "文章ID"
// @Success      200   {object}  ArticleDeleteResponse
// @Failure      404   {object}  apidoc.Error "记录不存在"
// @Failure      500   {object}  apidoc.Error "服务器错误"
// @Router       /api/admin/articles/{id} [delete]
// @Security     BearerAuth
func (c *ArticleController) Destroy(ctx http.Context) http.Response {

	id := helpers.GetUintRoute(ctx, "id")
	if _, err := c.ArticleService.GetByID(id); err != nil {
		return response.Error(ctx, http.StatusNotFound, apperrors.ErrRecordNotFound.Code)
	}
	if err := c.ArticleService.Delete(id); err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, "delete_success", http.Json{})

}
