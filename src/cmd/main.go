package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
	"github.com/CDCgov/reportstream-sftp-ingestion/report_stream"
	"log/slog"
	"os"
	"time"
)

func main() {

	setupLogging()

	slog.Info("Hello World")

	//TODO: Extract the client string to allow multi-environment
	azureBlobConnectionString := os.Getenv("AZURE_BLOB_CONNECTION_STRING")
	blobHandler, err := azure.NewBlobHandler(azureBlobConnectionString)
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any("error", err))
		os.Exit(1)
	}

	content, err := readAzureFile(blobHandler, "reportstream.txt")
	if err != nil {
		slog.Error("Failed to read the file", slog.Any("error", err))
		os.Exit(1)
	}

	apiHandler := report_stream.ApiHandler{BaseUrl: "http://localhost:7071"}
	reportId, err := apiHandler.SendReport(content)
	if err != nil {
		slog.Error("Failed to send the file to ReportStream", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("File sent to ReportStream", slog.String("reportId", reportId))

	for {
		t := time.Now()
		slog.Info(t.Format("2006-01-02T15:04:05Z07:00"))
		time.Sleep(10 * time.Second)
	}
}

func setupLogging() {
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(jsonLogger)
}

type BlobHandler interface {
	FetchFile(blobPath string) ([]byte, error)
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
