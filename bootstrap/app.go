package bootstrap

import (
	"goravel/config"
	"goravel/routes"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
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
		WithConfig(config.Boot).
		Create()
}
