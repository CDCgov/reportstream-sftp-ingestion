package utils

import (
	"bytes"
	"log/slog"
)

// General setup for all tests
func SetupLogger() *bytes.Buffer {
	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))
	return buffer
}
