package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type ArticleCreate struct {

	Name string `form:"name" json:"name"`

	Status string `form:"status" json:"status"`

}

func (r *ArticleCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ArticleCreate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{

		"name": "required",

		"status": "",

	}
	return rules
}

func (r *ArticleCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{

		"name.required": trans.Get(ctx, "validation_name_required"),

		"status.required": trans.Get(ctx, "validation_status_required"),

	}
}

func (r *ArticleCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{

		"name": trans.Get(ctx, "validation_name"),

		"status": trans.Get(ctx, "validation_status"),

	}
}