package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type ArticleUpdate struct {
	Title   *string `form:"title" json:"title" example:"文章标题"`                    // 文章标题（可选）
	Content *string `form:"content" json:"content" example:"这里是文章内容"`              // 文章内容（可选）
	Status  *uint8  `form:"status" json:"status" enums:"0,1" example:"1"`          // 发布状态（1-发布，0-未发布）
	AdminId *int    `form:"admin_id" json:"admin_id" example:"1"`                   // 发布管理员ID（可选）
}

func (r *ArticleUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ArticleUpdate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{

		"title":    "",
		"content":  "",
		"status":   "",
		"admin_id": "",
	}
	return rules
}

func (r *ArticleUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{

		"title.required":    trans.Get(ctx, "validation_title_required"),
		"content.required":  trans.Get(ctx, "validation_content_required"),
		"status.required":   trans.Get(ctx, "validation_status_required"),
		"admin_id.required": trans.Get(ctx, "validation_admin_id_required"),
	}
}

func (r *ArticleUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{

		"title":    trans.Get(ctx, "validation_title"),
		"content":  trans.Get(ctx, "validation_content"),
		"status":   trans.Get(ctx, "validation_status"),
		"admin_id": trans.Get(ctx, "validation_admin_id"),
	}
}
