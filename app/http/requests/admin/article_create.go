package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type ArticleCreate struct {
	Name    string `form:"name" json:"name"`
	Status  string `form:"status" json:"status"`
	AdminId string `form:"admin_id" json:"admin_id"`
}

func (r *ArticleCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *ArticleCreate) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{

		"name":     "required",
		"status":   "",
		"admin_id": "",
	}
	return rules
}

func (r *ArticleCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{

		"name.required":     trans.Get(ctx, "validation_name_required"),
		"status.required":   trans.Get(ctx, "validation_status_required"),
		"admin_id.required": trans.Get(ctx, "validation_admin_id_required"),
	}
}

func (r *ArticleCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{

		"name":     trans.Get(ctx, "validation_name"),
		"status":   trans.Get(ctx, "validation_status"),
		"admin_id": trans.Get(ctx, "validation_admin_id"),
	}
}
