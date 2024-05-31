package main

import (
	"fmt"
	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
	"github.com/CDCgov/reportstream-sftp-ingestion/local"
	"github.com/CDCgov/reportstream-sftp-ingestion/report_stream"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	azlog.SetListener(func(event azlog.Event, s string) {
		fmt.Println(s)
	})

	setupLogging()

	slog.Info("Hello World")

	go setupHealthCheck()

	azureBlobConnectionString := os.Getenv("AZURE_BLOB_CONNECTION_STRING")
	blobHandler, err := azure.NewStorageHandler(azureBlobConnectionString)
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any("error", err))
		os.Exit(1)
	}

	filepath := "order_message.hl7"
	content, err := readAzureFile(blobHandler, filepath)
	if err != nil {
		slog.Error("Failed to read the file", slog.String("filepath", filepath), slog.Any("error", err))
		os.Exit(1)
	}

	reportStreamBaseUrl := os.Getenv("REPORT_STREAM_URL_PREFIX")
	var messageSender MessageSender

	if reportStreamBaseUrl == "" {
		slog.Info("No report stream url prefix set, using local sender instead")
		messageSender = local.FileSender{}
	} else {
		slog.Info("Found report stream url prefix, will send to ReportStream")
		messageSender = report_stream.NewSender()
	}

	reportId, err := messageSender.SendMessage(content)
	if err != nil {
		slog.Warn("Failed to send the file to ReportStream", slog.Any("error", err))
		slog.Info("actually continuing just for now while we debug")
	} else {
		slog.Info("File sent to ReportStream", slog.String("reportId", reportId))
	}

	for {
		t := time.Now()
		slog.Info(t.Format("2006-01-02T15:04:05Z07:00"))
		time.Sleep(10 * time.Second)
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

	http.HandleFunc("/health", func(response http.ResponseWriter, request *http.Request) {
		slog.Info("Health check ping")
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

func readAzureFile(blobHandler BlobHandler, filePath string) ([]byte, error) {
	content, err := blobHandler.FetchFile(filePath)
	if err != nil {
		return nil, err
	}

	//TODO: Auth and call ReportStream
	slog.Info(string(content))

	return content, nil
}
