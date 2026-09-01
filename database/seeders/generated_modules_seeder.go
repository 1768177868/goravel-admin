package seeders

import (
	"context"

	"goravel/app/services"
)

type GeneratedModulesSeeder struct{}

func (s *GeneratedModulesSeeder) Signature() string {
	return "GeneratedModulesSeeder"
}

func (s *GeneratedModulesSeeder) Run() error {
	return services.InstallGeneratedModuleManifests(context.Background())
}
