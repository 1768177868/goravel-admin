package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("admin", map[string]any{
		// Protected Admin IDs
		//
		// These admin IDs cannot be deleted. Usually these are system administrators
		// or super administrators that are critical to the system operation.
		// You can add multiple IDs separated by commas in the .env file.
		"protected_ids": config.Env("ADMIN_PROTECTED_IDS", "1"), // Default: ID 1 (super admin)
	})
}
