package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/orchestration"
	"github.com/CDCgov/reportstream-sftp-ingestion/sftp"
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

	queueHandler, err := orchestration.NewQueueHandler()
	if err != nil {
		slog.Warn("Failed to create queueHandler", slog.Any(utils.ErrorKey, err))
	}

	// TODO - move calls to SFTP into whatever timer/trigger we set up later
	sftpHandler, err := sftp.NewSftpHandler()
	if err != nil {
		slog.Error("ope, failed to create sftp handler", slog.Any(utils.ErrorKey, err))
		// Don't return, we want to let things keep going for now
	}

	defer sftpHandler.Close()

	// TODO - Refactor when we have an answer for the SFTP subfolders and timer trigger is properly set up
	// TODO - Consider moving unzipping from SFTP handler to a more appropriate place
	if err == nil {
		sftpHandler.CopyFiles()
	}

	// TODO - add another queue listener for the other queue? Or maybe one listener for all queues but different message handling?
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
			slog.Error("Failed to respond to health check", slog.Any(utils.ErrorKey, err))
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("Failed to start health check", slog.Any(utils.ErrorKey, err))
	}
}
