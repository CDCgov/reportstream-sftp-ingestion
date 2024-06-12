package usecases

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/senders"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"log/slog"
	"os"
)

type ReadAndSend interface {
	ReadAndSend(filepath string) error
}

type ReadAndSendUsecase struct {
	blobHandler   BlobHandler
	messageSender senders.MessageSender
}

func NewReadAndSendUsecase() (ReadAndSendUsecase, error) {

	azureBlobConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	blobHandler, err := storage.NewStorageHandler(azureBlobConnectionString)
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any("error", err))
		return ReadAndSendUsecase{}, err
	}

	reportStreamBaseUrl := os.Getenv("REPORT_STREAM_URL_PREFIX")
	var messageSender senders.MessageSender

	if reportStreamBaseUrl == "" {
		slog.Info("REPORT_STREAM_URL_PREFIX not set, using file senders instead")
		messageSender = senders.FileSender{}
	} else {
		slog.Info("Found REPORT_STREAM_URL_PREFIX, will send to ReportStream")
		messageSender, err = senders.NewSender()
		if err != nil {
			slog.Warn("Failed to construct the ReportStream senders", slog.Any("error", err))
			return ReadAndSendUsecase{}, err
		}
	}

	return ReadAndSendUsecase{
		blobHandler:   blobHandler,
		messageSender: messageSender,
	}, nil
}

func (receiver *ReadAndSendUsecase) ReadAndSend(filepath string) error {
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
