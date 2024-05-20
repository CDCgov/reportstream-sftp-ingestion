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
	blobHandler, err := azure.NewBlobHandler("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;")
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any("error", err))
		os.Exit(1)
	}

	content, err := readAzureFile(blobHandler, "reportstream.txt")
	if err != nil {
		slog.Error("Failed to read the file", slog.Any("error", err))
		os.Exit(1)
	}

	apiHandler := report_stream.ApiHandler{"http://localhost:7071"}
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
