package usecases

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/senders"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"log/slog"
	"os"
	"strings"
)

type ReadAndSend interface {
	ReadAndSend(sourceUrl string) error
}

type ReadAndSendUsecase struct {
	blobHandler   BlobHandler
	messageSender senders.MessageSender
}

func NewReadAndSendUsecase() (ReadAndSendUsecase, error) {
	blobHandler, err := storage.NewStorageHandler()
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

func (receiver *ReadAndSendUsecase) ReadAndSend(sourceUrl string) error {
	content, err := receiver.blobHandler.FetchFile(sourceUrl)
	if err != nil {
		slog.Error("Failed to read the file", slog.String("filepath", sourceUrl), slog.Any("error", err))
		return err
	}

	reportId, err := receiver.messageSender.SendMessage(content)
	if err != nil {
		// TODO - move file to failure folder
		slog.Error("Failed to send the file", slog.Any("error", err))
		return err
	}

	slog.Info("File sent to ReportStream", slog.String("reportId", reportId))

	destinationUrl := strings.Replace(sourceUrl, "import", "success", 1)
	if destinationUrl == sourceUrl {
		slog.Error("Unexpected source URL, did not move", slog.String("sourceUrl", sourceUrl))
	} else {
		// After successful message handling, move source file
		err = receiver.blobHandler.MoveFile(sourceUrl, destinationUrl)
		if err != nil {
			slog.Error("Failed to move file after processing", slog.Any("error", err))
			// TODO - should we not return an error here? or should we return one but handle it?
			return err
		}
	}

	return nil
}
