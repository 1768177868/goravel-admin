package utils

import (
	"strings"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// OrdersEnabled reports whether the orders example module is enabled.
func OrdersEnabled() bool {
	return facades.Config().GetBool("module.orders_enabled", true)
}

// PaymentsEnabled reports whether the payments example module is enabled.
func PaymentsEnabled() bool {
	return facades.Config().GetBool("module.payments_enabled", true)
}

// DevToolsEnabled reports whether general dev tools (e.g. form demo) are available.
// Enabled in local/development/test, or when APP_ENABLE_DEV_TOOL=true.
func DevToolsEnabled() bool {
	env := appEnv()
	if env == "local" || env == "development" || env == "test" {
		return true
	}
	return facades.Config().GetBool("app.enable_dev_tool", false)
}

// CodeGeneratorEnabled reports whether the code generator is available.
// Enabled in local/development only (not test), or when APP_ENABLE_DEV_TOOL=true.
func CodeGeneratorEnabled() bool {
	env := appEnv()
	if env == "local" || env == "development" {
		return true
	}
	return facades.Config().GetBool("app.enable_dev_tool", false)
}

func appEnv() string {
	return strings.ToLower(strings.TrimSpace(facades.Config().GetString("app.env", "production")))
}

// ElasticsearchEnabled reports whether Elasticsearch integration is enabled.
func ElasticsearchEnabled() bool {
	return facades.Config().GetBool("elasticsearch.enabled", false)
}

// OTELEnabled reports whether OpenTelemetry exporters are configured.
func OTELEnabled() bool {
	cfg := facades.Config()
	if strings.TrimSpace(cfg.GetString("OTEL_TRACES_EXPORTER", "")) != "" {
		return true
	}
	return strings.TrimSpace(cfg.GetString("OTEL_METRICS_EXPORTER", "")) != ""
}

// DisabledModuleMenuSlugs returns menu slugs that should be hidden when modules are off.
func DisabledModuleMenuSlugs() map[string]bool {
	disabled := make(map[string]bool)
	if !OrdersEnabled() {
		disabled["order"] = true
	}
	if !PaymentsEnabled() {
		disabled["payment"] = true
		disabled["payment-method"] = true
		disabled["payment-record"] = true
	}
	codeGenOn := CodeGeneratorEnabled()
	formDemoOn := DevToolsEnabled()
	if !codeGenOn {
		disabled["code-generator"] = true
	}
	if !formDemoOn {
		disabled["form_demo"] = true
	}
	if !codeGenOn && !formDemoOn {
		disabled["dev"] = true
	}
	return disabled
}

// FilterFlatMenusByModule removes menus whose slug is disabled by module switches.
func FilterFlatMenusByModule(menus []models.Menu) []models.Menu {
	disabled := DisabledModuleMenuSlugs()
	if len(disabled) == 0 {
		return menus
	}

	hiddenIDs := make(map[uint]bool)
	for _, menu := range menus {
		if disabled[strings.ToLower(strings.TrimSpace(menu.Slug))] {
			hiddenIDs[menu.ID] = true
		}
	}

	changed := true
	for changed {
		changed = false
		for _, menu := range menus {
			if hiddenIDs[menu.ParentID] && !hiddenIDs[menu.ID] {
				hiddenIDs[menu.ID] = true
				changed = true
			}
		}
	}

	filtered := make([]models.Menu, 0, len(menus))
	for _, menu := range menus {
		if !hiddenIDs[menu.ID] {
			filtered = append(filtered, menu)
		}
	}
	return filtered
}

// FilterTreeMenusByModule recursively filters tree menus by module switches.
func FilterTreeMenusByModule(menus []models.Menu) []models.Menu {
	disabled := DisabledModuleMenuSlugs()
	if len(disabled) == 0 {
		return menus
	}

	filtered := make([]models.Menu, 0, len(menus))
	for _, menu := range menus {
		if disabled[strings.ToLower(strings.TrimSpace(menu.Slug))] {
			continue
		}
		if len(menu.Children) > 0 {
			menu.Children = FilterTreeMenusByModule(menu.Children)
		}
		filtered = append(filtered, menu)
	}
	return filtered
}
