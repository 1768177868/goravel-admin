package bootstrap

import (
	"fmt"
	"path"

	"goravel/app/facades"
)

// runnerDisabled reports whether a runner signature matches app.disabled_runners glob patterns.
func runnerDisabled(signature string) bool {
	for _, pattern := range facades.Config().GetStringSlice("app.disabled_runners", []string{}) {
		if pattern == "" {
			continue
		}
		if pattern == "*" {
			return true
		}
		matched, err := path.Match(pattern, signature)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func queueConnectionReady() bool {
	connection := facades.Config().GetString("queue.default")
	if connection == "" {
		return false
	}
	driver := facades.Config().GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
	return driver != "" && driver != "sync"
}

func shouldRunQueueRunner(signature string) bool {
	if !queueConnectionReady() {
		return false
	}
	if runnerDisabled("*") || runnerDisabled(signature) || runnerDisabled("queue-*") {
		return false
	}
	return true
}
