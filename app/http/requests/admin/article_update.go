package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type ArticleUpdate struct {
	AdminId *int64  `form:"admin_id" json:"admin_id"`
	Title   *string `form:"title" json:"title"`
	Content *string `form:"content" json:"content"`
	Status  *uint8  `form:"status" json:"status"`
}

func (r *ArticleUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ArticleUpdate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{
		"admin_id": "",
		"title":    "",
		"content":  "",
		"status":   "",
	}
	return rules
}

func (r *ArticleUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"admin_id.required": trans.Get(ctx, "validation_admin_id_required"),
		"title.required":    trans.Get(ctx, "validation_title_required"),
		"content.required":  trans.Get(ctx, "validation_content_required"),
		"status.required":   trans.Get(ctx, "validation_status_required"),
	}
}

func (r *ArticleUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"admin_id": trans.Get(ctx, "validation_admin_id"),
		"title":    trans.Get(ctx, "validation_title"),
		"content":  trans.Get(ctx, "validation_content"),
		"status":   trans.Get(ctx, "validation_status"),
	}
}
