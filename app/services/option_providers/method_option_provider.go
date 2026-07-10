package option_providers

import (
	"context"

	"github.com/goravel/framework/contracts/http"
)

type MethodOptionProvider struct {
	ctx context.Context
}

func NewMethodOptionProvider(ctx context.Context) *MethodOptionProvider {
	return &MethodOptionProvider{
		ctx: ctx}
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
