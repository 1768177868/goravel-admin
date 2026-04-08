package kafka

import (
	"context"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	frameworkerrors "github.com/goravel/framework/errors"
)

const (
	BindingQueue = "goravel.kafka.queue"
	Name         = "kafka"
)

var App foundation.Application

type ServiceProvider struct{}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			BindingQueue,
		},
		Dependencies: []string{
			binding.Config,
		},
		ProvideFor: []string{
			binding.Queue,
		},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.BindWith(BindingQueue, func(app foundation.Application, parameters map[string]any) (any, error) {
		config := app.MakeConfig()
		if config == nil {
			return nil, frameworkerrors.ConfigFacadeNotSet.SetModule(frameworkerrors.ModuleQueue)
		}

		queue := app.MakeQueue()
		if queue == nil {
			return nil, frameworkerrors.QueueFacadeNotSet.SetModule(frameworkerrors.ModuleQueue)
		}

		connection, ok := parameters["connection"].(string)
		if !ok || connection == "" {
			return nil, ErrQueueConnectionIsRequired
		}

		return NewQueue(context.Background(), config, queue, app.GetJson(), connection)
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {}
