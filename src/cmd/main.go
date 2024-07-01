package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/orchestration"
	"github.com/CDCgov/reportstream-sftp-ingestion/sftp"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func main() {

	setupLogging()

	slog.Info("Hello World")

	go setupHealthCheck()

	queueHandler, err := orchestration.NewQueueHandler()
	if err != nil {
		slog.Warn("Failed to create queueHandler", slog.Any("error", err))
	}

	// TODO - move calls to SFTP into whatever timer/trigger we set up later
	sftpHandler, err := sftp.NewSftpHandler()
	if err != nil {
		slog.Error("ope, failed to create sftp handler", slog.Any("error", err))
		// Don't return, we want to let things keep going for now
	}
	defer sftpHandler.Close()

	sftpHandler.CopyFiles()

	// ListenToQueue is not split into a separate Go Routine since it is the core driver of the application
	queueHandler.ListenToQueue()
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
