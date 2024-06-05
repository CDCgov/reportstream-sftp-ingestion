package main

import (
	"encoding/base64"
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

	azureBlobConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
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

func (receiver ReadAndSendUsecase) ReadAndSend(filepath string) error {
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

// Initial function that should be called from Main that will call the Queue function that checks for messages.
// Secondly it should call ReadAndSend if a filename comes back
// Finally it should call  storage.go to move and delete the file (Func needs created) and the queue.go function to delete the message if sent successfully
func (receiver ReadAndSendUsecase) CheckQueue() {

	queueHandler, err := azure.NewQueueHandler()
	if err != nil {
		slog.Warn("Failed to create queueHandler", slog.Any("error", err))
	}

	messageResponse, err := queueHandler.ListenToQueue()
	if err != nil {
		slog.Warn("ListenToQueue failed", slog.Any("error", err))
	}

	filepathBytes, err := base64.StdEncoding.DecodeString(*messageResponse.Messages[0].MessageText)

	filepath := string(filepathBytes)

	message := *messageResponse.Messages[0]

	usecase, err := NewReadAndSendUsecase()

	if err != nil {
		slog.Error("NewReadAndSendUsecase failed", slog.Any("error", err))
	}
	if filepath != "" {
		slog.Info("Calling read and send")
		err = usecase.ReadAndSend(filepath)
		if err != nil {
			slog.Error("ReadAndSend failed", slog.Any("error", err))
		}

		err = queueHandler.HandleMessage(message)

		if err != nil {
			slog.Error("HandleMessage failed", slog.Any("error", err))
		}
	} else {
		slog.Error("No queue file found", slog.Any("error", err))
	}

}
