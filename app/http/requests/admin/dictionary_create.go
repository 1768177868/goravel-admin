package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DictionaryCreate struct {
	Type        string `form:"type" json:"type"`
	Label       string `form:"label" json:"label"`
	Value       string `form:"value" json:"value"`
	Description string `form:"description" json:"description"`
	Status      uint8  `form:"status" json:"status"`
	Sort        int    `form:"sort" json:"sort"`
	Remark      string `form:"remark" json:"remark"`
}

func (r *DictionaryCreate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DictionaryCreate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        "required|max_len:50",
		"label":       "required|max_len:50",
		"value":       "required|max_len:100",
		"description": "max_len:255",
		"status":      "in:0,1",
		"remark":      "max_len:500",
	}
}

func (r *DictionaryCreate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.required":        trans.Get(ctx, "dictionary_type_required"),
		"type.max_len":         trans.Get(ctx, "validation_type_max"),
		"label.required":       trans.Get(ctx, "validation_label_required"),
		"label.max_len":        trans.Get(ctx, "validation_label_max"),
		"value.required":       trans.Get(ctx, "validation_value_required"),
		"value.max_len":        trans.Get(ctx, "validation_value_max"),
		"description.max_len":   trans.Get(ctx, "validation_description_max"),
		"status.in":            trans.Get(ctx, "validation_status_in"),
		"remark.max_len":       trans.Get(ctx, "validation_remark_max"),
	}
}

func (r *DictionaryCreate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        trans.Get(ctx, "validation_type"),
		"label":       trans.Get(ctx, "validation_label"),
		"value":       trans.Get(ctx, "validation_value"),
		"description": trans.Get(ctx, "validation_description"),
		"status":      trans.Get(ctx, "validation_status"),
		"remark":      trans.Get(ctx, "validation_remark"),
	}
}

func (r *DictionaryCreate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}

