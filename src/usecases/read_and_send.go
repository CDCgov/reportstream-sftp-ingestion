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

// ReadAndSend retrieves the specified blob from Azure and sends it to ReportStream. On a success response from ReportStream,
// we move the file to a `success` folder. On a non-transient error, we move the file to a `failure` folder and return
// `nil` so that we'll delete the queue message and not retry. On a transient error or an unknown error, we return
// an error, which will cause the queue message to retry later
func (receiver *ReadAndSendUsecase) ReadAndSend(sourceUrl string) error {
	content, err := receiver.blobHandler.FetchFile(sourceUrl)
	if err != nil {
		slog.Error("Failed to read the file", slog.String("filepath", sourceUrl), slog.Any("error", err))
		return err
	}

	reportId, err := receiver.messageSender.SendMessage(content)
	if err != nil {
		slog.Error("Failed to send the file to ReportStream", slog.Any("error", err), slog.String("sourceUrl", sourceUrl))

		// As of June 2024, only the 400 response triggers a move to the `failure` folder. Returning `nil` will let
		// queue.go delete the queue message so that it will stop retrying
		// We're treating all other errors as unexpected (and possibly transient) for now
		if strings.Contains(err.Error(), "400") {
			receiver.moveFile(sourceUrl, "failure")
			return nil
		}

		// For any other failures,  return an error so that we'll leave the message on the queue and keep retrying
		return err
	}

	slog.Info("File sent to ReportStream", slog.String("reportId", reportId))

	receiver.moveFile(sourceUrl, "success")

	return nil
}

func (receiver *ReadAndSendUsecase) moveFile(sourceUrl string, newFolderName string) {
	destinationUrl := strings.Replace(sourceUrl, "import", newFolderName, 1)

	if destinationUrl == sourceUrl {
		slog.Error("Unexpected source URL, did not move", slog.String("sourceUrl", sourceUrl))
		return
	}

	// After successful message handling, move source file
	err := receiver.blobHandler.MoveFile(sourceUrl, destinationUrl)
	if err != nil {
		slog.Error("Failed to move file after processing", slog.Any("error", err))
	}
}
