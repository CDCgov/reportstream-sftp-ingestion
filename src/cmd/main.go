package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/orchestration"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {

	setupLogging()

	slog.Info("Hello World")

	go setupHealthCheck()

	setUpQueues()

	// This loop keeps the app alive. This lets the pre-live deployment slot remain healthy
	// even though it's configured without queues, which means we can quickly swap slots
	// if needed without having to redeploy
	for {
		t := time.Now()
		slog.Info(t.Format("2006-01-02T15:04:05Z07:00"))
		time.Sleep(10 * time.Minute)
	}
}

func setUpQueues() {
	// Set up the polling message handler and queue listener
	pollingMessageHandler := orchestration.PollingMessageHandler{}

	pollingQueueHandler, err := orchestration.NewQueueHandler(pollingMessageHandler, "polling-trigger")
	if err != nil {
		slog.Warn("Failed to create pollingQueueHandler", slog.Any(utils.ErrorKey, err))
		return
	}
	go func() {
		pollingQueueHandler.ListenToQueue()
	}()

	// Set up the import message handler and queue listener
	importMessageHandler, err := orchestration.NewImportMessageHandler()
	if err != nil {
		slog.Warn("Failed to create importMessageHandler", slog.Any(utils.ErrorKey, err))
		return
	}
	importQueueHandler, err := orchestration.NewQueueHandler(importMessageHandler, "message-import")
	if err != nil {
		slog.Warn("Failed to create importQueueHandler", slog.Any(utils.ErrorKey, err))
		return
	}

	go func() {
		importQueueHandler.ListenToQueue()
	}()
}

func setupLogging() {
	environment := os.Getenv("ENV")

	if environment == "" {
		//environment = "local"
	}

	if environment != "local" {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	}

	// add a unique ID to the currently running container so we can identify logs uniquely in a better way than Azure's built-in _ResourceId and Host field
	slog.SetDefault(slog.With(slog.String("containerId", uuid.NewString())))
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
