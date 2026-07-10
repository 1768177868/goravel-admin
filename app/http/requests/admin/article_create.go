package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type ArticleCreate struct {
	AdminId int64  `form:"admin_id" json:"admin_id"`
	Title   string `form:"title" json:"title"`
	Content string `form:"content" json:"content"`
	Status  uint8  `form:"status" json:"status"`
}

func (r *ArticleCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ArticleCreate) Rules(ctx http.Context) map[string]any {
	rules := map[string]any{
		"admin_id": "required",
		"title":    "required",
		"content":  "",
		"status":   "required",
	}
	return rules
}

func (r *ArticleCreate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"admin_id": trans.Get(ctx, "validation.attributes.admin_id"),
		"title":    trans.Get(ctx, "validation.attributes.title"),
		"content":  trans.Get(ctx, "validation.attributes.content"),
		"status":   trans.Get(ctx, "validation.attributes.status"),
	}
}

func (r *ArticleCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	if err := helpers.PrepareRichTextFieldForValidation(data, "content"); err != nil {
		return err
	}
	return nil
}
