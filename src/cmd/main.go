package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
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

	usecase, err := NewReadAndSendUsecase()
	if err != nil {
		slog.Warn("Failed to init the usecase", slog.Any("error", err))
		slog.Info("Continuing for now while debugging")
	}

	err = usecase.ReadAndSend()
	if err != nil {
		slog.Warn("Usecase failed", slog.Any("error", err))
		slog.Info("Continuing for now while debugging")
	}

	for {
		azure.ListenToQueue()
		t := time.Now()
		slog.Info(t.Format("2006-01-02T15:04:05Z07:00"))
		time.Sleep(1 * time.Minute)
	}
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
			slog.Error("Failed to respond to health check", slog.Any("error", err))
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("Failed to start health check", slog.Any("error", err))
	}
}
