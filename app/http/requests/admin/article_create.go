package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type ArticleCreate struct {
	Title   string `form:"title" json:"title"`
	Content string `form:"content" json:"content"`
	Status  uint8  `form:"status" json:"status"`
	AdminId int    `form:"admin_id" json:"admin_id"`
}

func (r *ArticleCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ArticleCreate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{

		"title":    "required",
		"content":  "",
		"status":   "required",
		"admin_id": "required",
	}
	return rules
}

func (r *ArticleCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{

		"title.required":    trans.Get(ctx, "validation_title_required"),
		"content.required":  trans.Get(ctx, "validation_content_required"),
		"status.required":   trans.Get(ctx, "validation_status_required"),
		"admin_id.required": trans.Get(ctx, "validation_admin_id_required"),
	}
}

func (r *ArticleCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{

		"title":    trans.Get(ctx, "validation_title"),
		"content":  trans.Get(ctx, "validation_content"),
		"status":   trans.Get(ctx, "validation_status"),
		"admin_id": trans.Get(ctx, "validation_admin_id"),
	}
}
