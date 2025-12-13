package option_providers

import (
	"github.com/goravel/framework/contracts/http"
)

type MethodOptionProvider struct{}

func NewMethodOptionProvider() *MethodOptionProvider {
	return &MethodOptionProvider{}
}

func (p *MethodOptionProvider) GetOptions(ctx http.Context) (map[string]any, error) {
	options := []map[string]any{
		{"label": "GET", "value": "GET"},
		{"label": "POST", "value": "POST"},
		{"label": "PUT", "value": "PUT"},
		{"label": "DELETE", "value": "DELETE"},
		{"label": "PATCH", "value": "PATCH"},
	}

	return map[string]any{
		"options": options,
	}, nil
}

