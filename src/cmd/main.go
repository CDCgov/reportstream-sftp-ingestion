package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/orchestration"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func main() {

	setupLogging()

	slog.Info("Hello World")

	go setupHealthCheck()

	// Set up the polling message handler and queue listener
	pollingMessageHandler := orchestration.PollingMessageHandler{}

	pollingQueueHandler, err := orchestration.NewQueueHandler(pollingMessageHandler, "polling-trigger")
	if err != nil {
		slog.Warn("Failed to create pollingQueueHandler", slog.Any(utils.ErrorKey, err))
	}
	go func() {
		pollingQueueHandler.ListenToQueue()
	}()

	// Set up the import message handler and queue listener
	importMessageHandler, err := orchestration.NewImportMessageHandler()
	if err != nil {
		slog.Warn("Failed to create importMessageHandler", slog.Any(utils.ErrorKey, err))
	}
	importQueueHandler, err := orchestration.NewQueueHandler(importMessageHandler, "message-import")
	if err != nil {
		slog.Warn("Failed to create importQueueHandler", slog.Any(utils.ErrorKey, err))
	}
	// This ListenToQueue is not split into a separate Go Routine since it is the core driver of the application
	importQueueHandler.ListenToQueue()
}

func setupLogging() {
	environment := os.Getenv("ENV")

	if environment == "" {
		environment = "local"
	}

	if environment != "local" {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	}

}

func setupHealthCheck() {
	slog.Info("Bootstrapping health check")

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		slog.Info("Health check ping", slog.String("method", request.Method), slog.String("path", request.URL.String()))

		_, err := io.WriteString(response, "Operational")
		if err != nil {
			slog.Error("Failed to respond to health check", slog.Any(utils.ErrorKey, err))
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("Failed to start health check", slog.Any(utils.ErrorKey, err))
	}
}
