package admin

import (
	"goravel/app/http/helpers"
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DictionaryUpdate struct {
	Type           *string `form:"type" json:"type"`
	Label          *string `form:"label" json:"label"`
	Value          *string `form:"value" json:"value"`
	TranslationKey *string `form:"translation_key" json:"translation_key"`
	Description    *string `form:"description" json:"description"`
	Status         *uint8  `form:"status" json:"status"`
	Sort           *int    `form:"sort" json:"sort"`
	Remark         *string `form:"remark" json:"remark"`
}

func (r *DictionaryUpdate) Authorize(ctx http.Context) error {
	return nil
}

func (r *DictionaryUpdate) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"type":            "max:50",
		"label":           "max:50",
		"value":           "max:100",
		"translation_key": "max:255",
		"description":     "max:255",
		"status":          "in:0,1",
		"remark":          "max:500",
	}
}

func (r *DictionaryUpdate) Attributes(ctx http.Context) map[string]any {
	return map[string]any{
		"type":            trans.Get(ctx, "validation.attributes.type"),
		"label":           trans.Get(ctx, "validation.attributes.label"),
		"value":           trans.Get(ctx, "validation.attributes.value"),
		"translation_key": trans.Get(ctx, "validation.attributes.translation_key"),
		"description":     trans.Get(ctx, "validation.attributes.description"),
		"status":          trans.Get(ctx, "validation.attributes.status"),
		"remark":          trans.Get(ctx, "validation.attributes.remark"),
	}
}

func (r *DictionaryUpdate) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return helpers.PrepareNumericFieldForValidation(data, "status")
}
