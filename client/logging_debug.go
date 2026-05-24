//go:build debug

package client

import "log"

const enableLogging = true

func logDebug(format string, v ...any) {
	log.Printf("[Anther DEBUG] "+format, v...)
}
