package utils

import (
	"bytes"
	"log/slog"
)

// SetupLogger can be used when asserting slog output during testing
func SetupLogger() (*bytes.Buffer, *slog.Logger) {
	defaultLogger := slog.Default()

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	return buffer, defaultLogger
}

// CONSTANTS

const SuccessSourceUrl = "http://localhost/sftp/customer/success/order_message.hl7"
const SourceUrl = "http://localhost/sftp/customer/import/order_message.hl7"
const FailureSourceUrl = "http://localhost/sftp/customer/failure/order_message.hl7"
