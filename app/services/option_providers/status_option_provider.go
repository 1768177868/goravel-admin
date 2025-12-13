package option_providers

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/trans"
)

type StatusOptionProvider struct{}

func NewStatusOptionProvider() *StatusOptionProvider {
	return &StatusOptionProvider{}
}

func (p *StatusOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	options := []map[string]any{
		{"label": trans.Get(ctx, "common.enabled"), "value": "1"},
		{"label": trans.Get(ctx, "common.disabled"), "value": "0"},
	}

	return map[string]any{
		"options": options,
	}, nil
}

