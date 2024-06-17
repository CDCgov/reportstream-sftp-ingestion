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
	/*
			Four possible scenarios to handle:
			- Success from reportstream - move to 'success' folder and delete message
			- Transient error from reportstream (like a 404, 502 (bad gateway), or 503 (service unavailable) status)
				- this is something to retry, so leave message on queue and don't move file
			- Non-transient errors from reportstream (e.g. message body is wrong shape or credentials are wrong - need to
				look at their error responses). No point retrying, so move file to 'failure' folder and delete q message
			- Azure errors - these are probably transient, so leave message on queue and don't move file

		TODO - How do we decide if a reportstream error or azure error is transient?
	*/
	content, err := receiver.blobHandler.FetchFile(sourceUrl)
	if err != nil {
		slog.Error("Failed to read the file", slog.String("filepath", sourceUrl), slog.Any("error", err))
		return err
	}

	reportId, err := receiver.messageSender.SendMessage(content)
	if err != nil {
		slog.Error("Failed to send the file to ReportStream", slog.Any("error", err), slog.String("sourceUrl", sourceUrl))

		// Move file to the `failure` folder
		destinationUrl := strings.Replace(sourceUrl, "import", "failure", 1)
		if destinationUrl == sourceUrl {
			slog.Error("Unexpected source URL, did not move", slog.String("sourceUrl", sourceUrl))
		} else {
			// After successful message handling, move source file
			moveErr := receiver.blobHandler.MoveFile(sourceUrl, destinationUrl)
			if moveErr != nil {
				slog.Error("Failed to move file after processing", slog.Any("error", moveErr))
			}
		}

		// Return an error so that we'll leave the message on the queue and keep retrying
		// TODO - differentiate between transient and non-transient RS errors - here or in queue.go?
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
			// Log an error, but return nil. We've successfully sent the file to ReportStream,
			// so we should delete the queue message
			return nil
		}
	}

	return nil
}
