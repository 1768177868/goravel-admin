package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DictionaryUpdate struct {
	Type        string `form:"type" json:"type"`
	Label       string `form:"label" json:"label"`
	Value       string `form:"value" json:"value"`
	Description string `form:"description" json:"description"`
	Status      uint8  `form:"status" json:"status"`
	Sort        int    `form:"sort" json:"sort"`
	Remark      string `form:"remark" json:"remark"`
}

func (r *DictionaryUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DictionaryUpdate) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        "max_len:50",
		"label":       "max_len:50",
		"value":       "max_len:100",
		"description": "max_len:255",
		"status":      "in:0,1",
		"remark":      "max_len:500",
	}
}

func (r *DictionaryUpdate) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"type.max_len":        trans.Get(ctx, "validation_type_max"),
		"label.max_len":       trans.Get(ctx, "validation_label_max"),
		"value.max_len":       trans.Get(ctx, "validation_value_max"),
		"description.max_len": trans.Get(ctx, "validation_description_max"),
		"status.in":           trans.Get(ctx, "validation_status_in"),
		"remark.max_len":      trans.Get(ctx, "validation_remark_max"),
	}
}

func (r *DictionaryUpdate) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"type":        trans.Get(ctx, "validation_type"),
		"label":       trans.Get(ctx, "validation_label"),
		"value":       trans.Get(ctx, "validation_value"),
		"description": trans.Get(ctx, "validation_description"),
		"status":      trans.Get(ctx, "validation_status"),
		"remark":      trans.Get(ctx, "validation_remark"),
	}
}

func (r *DictionaryUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}

