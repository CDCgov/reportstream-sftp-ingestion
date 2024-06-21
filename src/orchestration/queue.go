package orchestration

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type QueueHandler struct {
	queueClient           QueueClient
	deadLetterQueueClient QueueClient
	ctx                   context.Context
	usecase               usecases.ReadAndSend
}

type QueueClient interface {
	DeleteMessage(ctx context.Context, messageID string, popReceipt string, o *azqueue.DeleteMessageOptions) (azqueue.DeleteMessageResponse, error)
	DequeueMessage(ctx context.Context, o *azqueue.DequeueMessageOptions) (azqueue.DequeueMessagesResponse, error)
	EnqueueMessage(ctx context.Context, content string, o *azqueue.EnqueueMessageOptions) (azqueue.EnqueueMessagesResponse, error)
}

func NewQueueHandler() (QueueHandler, error) {
	azureQueueConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")

	client, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, "blob-message-queue", nil)
	if err != nil {
		slog.Error("Unable to create Azure Queue Client for primary queue", err)
		return QueueHandler{}, err
	}

	dlqClient, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, "blob-message-dead-letter-queue", nil)
	if err != nil {
		slog.Error("Unable to create Azure Queue Client for dead letter queue", err)
		return QueueHandler{}, err
	}

	usecase, err := usecases.NewReadAndSendUsecase()

	if err != nil {
		slog.Error("Unable to create Usecase", err)
		return QueueHandler{}, err
	}

	return QueueHandler{queueClient: client, deadLetterQueueClient: dlqClient, ctx: context.Background(), usecase: &usecase}, nil
}

func getUrlFromMessage(messageText string) (string, error) {
	eventBytes, err := base64.StdEncoding.DecodeString(messageText)
	if err != nil {
		slog.Error("Failed to decode message text", slog.Any("error", err))
		return "", err
	}

	// Map bytes json to Event object format (shape)
	var event azeventgrid.Event
	err = event.UnmarshalJSON(eventBytes)
	if err != nil {
		slog.Error("Failed to unmarshal event", slog.Any("error", err))
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

func (receiver QueueHandler) deleteMessage(message azqueue.DequeuedMessage) error {
	messageId := *message.MessageID
	popReceipt := *message.PopReceipt

	deleteResponse, err := receiver.queueClient.DeleteMessage(receiver.ctx, messageId, popReceipt, nil)
	if err != nil {
		slog.Error("Unable to delete message", slog.Any("error", err))
		return err
	}

	slog.Info("message deleted", slog.Any("delete message response", deleteResponse))

	return nil
}

func (receiver QueueHandler) handleMessage(message azqueue.DequeuedMessage) error {
	slog.Info("Handling message", slog.String("id", *message.MessageID))

	overThreshold := receiver.overDeliveryThreshold(message)
	if overThreshold {
		return errors.New("message delivery threshold exceeded")
	}

	sourceUrl, err := getUrlFromMessage(*message.MessageText)

	if err != nil {
		slog.Error("Failed to get the file URL", slog.Any("error", err))
		return err
	}

	err = receiver.usecase.ReadAndSend(sourceUrl)

	if err != nil {
		slog.Warn("Failed to read/send file", slog.Any("error", err))
	} else {
		// Only delete message if file successfully sent to ReportStream
		// (or if there's a known non-transient error and we've moved the file to `failure`)
		err = receiver.deleteMessage(message)
		if err != nil {
			slog.Warn("Failed to delete message", slog.Any("error", err))
			return err
		}
	}

	return nil
}

// overDeliveryThreshold checks whether the max delivery attempts for the message have been reached.
// If the threshold has been reached, the message should go to dead letter storage.
// Return true if we're over the threshold and should stop processing, else return false
func (receiver QueueHandler) overDeliveryThreshold(message azqueue.DequeuedMessage) bool {
	maxDeliveryCount, err := strconv.ParseInt(os.Getenv("QUEUE_MAX_DELIVERY_ATTEMPTS"), 10, 64)

	if err != nil {
		maxDeliveryCount = 5
		slog.Warn("Failed to parse QUEUE_MAX_DELIVERY_ATTEMPTS, defaulting to 5", slog.Any("error", err))
	}

	if *message.DequeueCount > maxDeliveryCount {
		slog.Error("Message reached maximum number of delivery attempts", slog.Any("message", message))
		err := receiver.deadLetter(message)
		if err != nil {
			slog.Error("Failed to move message to the DLQ", slog.Any("message", message))
		}
		return true
	}
	return false
}

func (receiver QueueHandler) deadLetter(message azqueue.DequeuedMessage) error {

	// a TimeToLive of -1 means the message will not expire
	opts := &azqueue.EnqueueMessageOptions{TimeToLive: to.Ptr(int32(-1))}
	_, err := receiver.deadLetterQueueClient.EnqueueMessage(context.Background(), *message.MessageText, opts)
	if err != nil {
		slog.Error("Failed to add the message to the DLQ", slog.Any("error", err))
		return err
	}

	err = receiver.deleteMessage(message)
	if err != nil {
		slog.Error("Failed to delete the message to the original queue after adding it to the DLQ", slog.Any("error", err))
		return err
	}

	slog.Info("Successfully moved the message to the DLQ")

	return nil
}

func (receiver QueueHandler) ListenToQueue() {
	for {
		err := receiver.receiveQueue()
		if err != nil {
			slog.Error("Failed to receive message", slog.Any("error", err))
		}
		time.Sleep(10 * time.Second)
	}
}

func (receiver QueueHandler) receiveQueue() error {

	slog.Info("Trying to dequeue")

	messageResponse, err := receiver.queueClient.DequeueMessage(receiver.ctx, nil)
	if err != nil {
		slog.Error("Unable to dequeue messages", err)
		return err
	}

	for _, dequeuedMessage := range messageResponse.Messages {
		message := *dequeuedMessage
		go func() {
			err := receiver.handleMessage(message)
			if err != nil {
				slog.Error("Unable to handle message", err)
			}
		}()
	}

	return nil
}
