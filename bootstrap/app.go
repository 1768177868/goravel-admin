package bootstrap

import (
	"goravel/config"
	"goravel/routes"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithSeeders(Seeders).
		WithRouting(func() {
			routes.Web()
			routes.Api()
			routes.Admin()
			routes.Pprof()
		}).
		WithProviders(Providers).
		WithRunners(func() []contractsfoundation.Runner {
			return QueueRunners()
		}).
		WithCommandsFilter(func() []string {
			if facades.Config().GetString("app.env", "production") != "production" {
				return nil
			}
			return []string{
				"up", "down", "about", "key:generate",
				"schedule:*", "queue:*", "migrate*",
				"db:*", "cache:*", "config:*", "lang:*",
				"app:*", "order:*", "payment:*", "es:*", "token:*",
			}
		}).
		WithConfig(config.Boot).
		Create()
}
