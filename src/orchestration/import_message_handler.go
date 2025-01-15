package orchestration

import (
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
)

type ImportMessageHandler struct {
	usecase usecases.ReadAndSend
}

func NewImportMessageHandler() (ImportMessageHandler, error) {
	usecase, err := usecases.NewReadAndSendUsecase()

	if err != nil {
		slog.Error("Unable to create Usecase", slog.Any(utils.ErrorKey, err))
		return ImportMessageHandler{}, err
	}

	return ImportMessageHandler{usecase: &usecase}, nil
}

func (receiver ImportMessageHandler) HandleMessageContents(message azqueue.DequeuedMessage) error {
	sourceUrl, err := getUrlFromMessage(*message.MessageText)

	if err != nil {
		slog.Error("Failed to get the file URL", slog.Any(utils.ErrorKey, err))
		return err
	}

	// TODO - parse partner ID from sourceUrl either here or in ReadAndSend
	// URL looks like https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7
	// and we need to parse out `customer` (and probably see if it's a real one)
	return receiver.usecase.ReadAndSend(sourceUrl)
}

func getUrlFromMessage(messageText string) (string, error) {
	eventBytes, err := base64.StdEncoding.DecodeString(messageText)
	if err != nil {
		slog.Error("Failed to decode message text", slog.Any(utils.ErrorKey, err))
		return "", err
	}

	// Map bytes json to Event object format (shape)
	var event azeventgrid.Event
	err = event.UnmarshalJSON(eventBytes)

	if err != nil {
		slog.Error("Failed to unmarshal event", slog.Any(utils.ErrorKey, err))
		return "", err
	}

	// Data is an 'any' type. We need to tell Go that it's a map
	eventData, ok := event.Data.(map[string]any)

	if !ok {
		slog.Error("Could not assert event data to a map", slog.Any("event", event))
		return "", errors.New("could not assert event data to a map")
	}

	// Extract blob url from Event's data
	eventUrl, ok := eventData["url"].(string)

	if !ok {
		slog.Error("Could not assert event data url to a string", slog.Any("event", event))
		return "", errors.New("could not assert event data url to a string")
	}

	return eventUrl, nil
}
