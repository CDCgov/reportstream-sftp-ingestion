package usecases

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/senders"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"golang.org/x/text/encoding/charmap"
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
	blobHandler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any(utils.ErrorKey, err))
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
			slog.Warn("Failed to construct the ReportStream senders", slog.Any(utils.ErrorKey, err))
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
	content, err := receiver.blobHandler.FetchFileByUrl(sourceUrl)
	if err != nil {
		slog.Error("Failed to read the file", slog.String("filepath", sourceUrl), slog.Any(utils.ErrorKey, err))
		return err
	}

	encodedContent, err := receiver.ConvertToUtf8(content)
	if err != nil {
		slog.Error("Failed to encode content", slog.String("filepath", sourceUrl), slog.Any(utils.ErrorKey, err))
		return err
	}

	reportId, err := receiver.messageSender.SendMessage(encodedContent)
	if err != nil {
		slog.Error("Failed to send the file to ReportStream", slog.Any(utils.ErrorKey, err), slog.String("sourceUrl", sourceUrl))

		// As of August 2024, we trigger on any http status code >= 400 and < 500 and move to the `failure` folder.
		// Returning `nil` will let queue.go delete the queue message so that it will stop retrying
		// We're treating all other errors as unexpected (and possibly transient) for now
		if strings.Contains(err.Error(), utils.ReportStreamNonTransientFailure) {
			receiver.moveFile(sourceUrl, utils.FailureFolder)
			return nil
		}

		// For any other failures,  return an error so that we'll leave the message on the queue and keep retrying
		return err
	}

	slog.Info("File sent to ReportStream", slog.String("reportId", reportId))

	receiver.moveFile(sourceUrl, utils.SuccessFolder)

	return nil
}

// ConvertToUtf8 converts an HL7 file to UTF-8 encoding, which ReportStream expects
// CADPH files are ISO-8859-1, so for now we'll assume all files are this format
// TODO - make this conversion dynamic, possibly by file detection or partner config
func (receiver *ReadAndSendUsecase) ConvertToUtf8(content []byte) ([]byte, error) {
	encodedContent, err := charmap.ISO8859_1.NewDecoder().Bytes(content)
	if err != nil {
		return nil, err
	}
	return encodedContent, nil
}

func (receiver *ReadAndSendUsecase) moveFile(sourceUrl string, newFolderName string) {
	destinationUrl := strings.Replace(sourceUrl, utils.MessageStartingFolderPath, newFolderName, 1)

	if destinationUrl == sourceUrl {
		slog.Error("Unexpected source URL, did not move", slog.String("sourceUrl", sourceUrl))
		return
	}

	// After successful message handling, move source file
	err := receiver.blobHandler.MoveFile(sourceUrl, destinationUrl)
	if err != nil {
		slog.Error("Failed to move file after processing", slog.Any(utils.ErrorKey, err))
	}
}
