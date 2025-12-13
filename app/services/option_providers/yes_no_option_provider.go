package option_providers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/trans"
)

type YesNoOptionProvider struct{}

func NewYesNoOptionProvider() *YesNoOptionProvider {
	return &YesNoOptionProvider{}
}

func (p *YesNoOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	options := []map[string]any{
		{"label": trans.Get(ctx, "common.yes"), "value": "1"},
		{"label": trans.Get(ctx, "common.no"), "value": "0"},
	}

	return map[string]any{
		"options": options,
	}, nil
}

