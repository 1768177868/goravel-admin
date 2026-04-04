package providers

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/binding"
	"goravel/app/clients"
)

type ElasticsearchServiceProvider struct{}

func (r *ElasticsearchServiceProvider) Register(app foundation.Application) {
	cfg := app.MakeConfig()
	if !cfg.GetBool("elasticsearch.enabled", false) {
		return
	}

	app.Singleton(binding.ElasticsearchClient, func(app foundation.Application) (any, error) {
		return clients.NewElasticsearchClient(app.MakeConfig(), "")
	})
}

func (r *ElasticsearchServiceProvider) Boot(app foundation.Application) {}
