package main

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
	"github.com/CDCgov/reportstream-sftp-ingestion/local"
	"github.com/CDCgov/reportstream-sftp-ingestion/report_stream"
	"log/slog"
	"os"
)

type ReadAndSendUsecase struct {
	blobHandler   BlobHandler
	messageSender MessageSender
}

func NewReadAndSendUsecase() (ReadAndSendUsecase, error) {

	azureBlobConnectionString := os.Getenv("AZURE_BLOB_CONNECTION_STRING")
	blobHandler, err := azure.NewStorageHandler(azureBlobConnectionString)
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any("error", err))
		return ReadAndSendUsecase{}, err
	}

	reportStreamBaseUrl := os.Getenv("REPORT_STREAM_URL_PREFIX")
	var messageSender MessageSender

	if reportStreamBaseUrl == "" {
		slog.Info("REPORT_STREAM_URL_PREFIX not set, using file sender instead")
		messageSender = local.FileSender{}
	} else {
		slog.Info("Found REPORT_STREAM_URL_PREFIX, will send to ReportStream")
		messageSender, err = report_stream.NewSender()
		if err != nil {
			slog.Warn("Failed to construct the ReportStream sender", slog.Any("error", err))
			return ReadAndSendUsecase{}, err
		}
	}

	return ReadAndSendUsecase{
		blobHandler:   blobHandler,
		messageSender: messageSender,
	}, nil
}

func (receiver ReadAndSendUsecase) ReadAndSend() error {
	filepath := "order_message.hl7"
	content, err := receiver.blobHandler.FetchFile(filepath)
	if err != nil {
		slog.Error("Failed to read the file", slog.String("filepath", filepath), slog.Any("error", err))
		return err
	}

	reportId, err := receiver.messageSender.SendMessage(content)
	if err != nil {
		slog.Error("Failed to send the file", slog.Any("error", err))
		return err
	}

	slog.Info("File sent to ReportStream", slog.String("reportId", reportId))

	return nil
}
