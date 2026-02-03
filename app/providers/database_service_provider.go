package providers

import (
	"github.com/goravel/framework/contracts/foundation"
)

type DatabaseServiceProvider struct {
}

func (receiver *DatabaseServiceProvider) Register(app foundation.Application) {

}

func (receiver *DatabaseServiceProvider) Boot(app foundation.Application) {
	// Migrations and seeders are now registered in bootstrap/app.go via WithMigrations and WithSeeders
}
