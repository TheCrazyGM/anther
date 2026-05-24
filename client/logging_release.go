//go:build !debug

package client

const enableLogging = false

func logDebug(format string, v ...any) {}
