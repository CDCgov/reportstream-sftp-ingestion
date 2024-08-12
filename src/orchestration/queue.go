package orchestration

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type QueueHandler struct {
	queueClient           QueueClient
	deadLetterQueueClient QueueClient
	messageContentHandler MessageContentHandler
}

type QueueClient interface {
	DeleteMessage(ctx context.Context, messageID string, popReceipt string, o *azqueue.DeleteMessageOptions) (azqueue.DeleteMessageResponse, error)
	DequeueMessage(ctx context.Context, o *azqueue.DequeueMessageOptions) (azqueue.DequeueMessagesResponse, error)
	EnqueueMessage(ctx context.Context, content string, o *azqueue.EnqueueMessageOptions) (azqueue.EnqueueMessagesResponse, error)
}

type MessageContentHandler interface {
	HandleMessageContents(message azqueue.DequeuedMessage) error
}

func NewQueueHandler(messageContentHandler MessageContentHandler, queueBaseName string) (QueueHandler, error) {
	azureQueueConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")

	client, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, queueBaseName+"-queue", nil)
	if err != nil {
		slog.Error("Unable to create Azure Queue Client for primary queue", slog.Any(utils.ErrorKey, err))
		return QueueHandler{}, err
	}

	dlqClient, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, queueBaseName+"-dead-letter-queue", nil)
	if err != nil {
		slog.Error("Unable to create Azure Queue Client for dead letter queue", slog.Any(utils.ErrorKey, err))
		return QueueHandler{}, err
	}

	return QueueHandler{queueClient: client, deadLetterQueueClient: dlqClient, messageContentHandler: messageContentHandler}, nil
}

func createQueueClient(connectionString, queueName string) (QueueClient, error) {
	client, err := azqueue.NewQueueClientFromConnectionString(connectionString, queueName, nil)
	if err != nil {
		slog.Error("Unable to create Azure Queue Client", slog.String("queueName", queueName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}
	return client, nil
}

func (receiver QueueHandler) deleteMessage(message azqueue.DequeuedMessage) error {
	messageId := *message.MessageID
	popReceipt := *message.PopReceipt

	deleteResponse, err := receiver.queueClient.DeleteMessage(context.Background(), messageId, popReceipt, nil)
	if err != nil {
		slog.Error("Unable to delete message", slog.Any(utils.ErrorKey, err))
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

	//Below function call is where the logical path splits depending on if the receiver is an import or polling queue
	err := receiver.messageContentHandler.HandleMessageContents(message)

	if err != nil {
		slog.Warn("Failed to handle message", slog.Any(utils.ErrorKey, err))
	} else {
		err = receiver.deleteMessage(message)
		if err != nil {
			slog.Warn("Failed to delete message", slog.Any(utils.ErrorKey, err))
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
		slog.Warn("Failed to parse QUEUE_MAX_DELIVERY_ATTEMPTS, defaulting to 5", slog.Any(utils.ErrorKey, err))
	}

	if *message.DequeueCount > maxDeliveryCount {
		slog.Error("Message reached maximum number of delivery attempts", slog.Any("message", message))
		err := receiver.deadLetter(message)
		if err != nil {
			slog.Error("Failed to move message to the DLQ", slog.Any("message", message), slog.Any(utils.ErrorKey, err))
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
		slog.Error("Failed to add the message to the DLQ", slog.Any(utils.ErrorKey, err))
		return err
	}

	err = receiver.deleteMessage(message)
	if err != nil {
		slog.Error("Failed to delete the message to the original queue after adding it to the DLQ", slog.Any(utils.ErrorKey, err))
		return err
	}

	slog.Info("Successfully moved the message to the DLQ")

	return nil
}

func (receiver QueueHandler) ListenToQueue() {
	for {
		err := receiver.receiveQueue()
		if err != nil {
			slog.Error("Failed to receive message", slog.Any(utils.ErrorKey, err))
		}
		time.Sleep(10 * time.Second)
	}
}

func (receiver QueueHandler) receiveQueue() error {

	slog.Info("Trying to dequeue")

	// 15 minutes in seconds
	var timeoutValue int32 = 900
	var options = azqueue.DequeueMessageOptions{
		VisibilityTimeout: &timeoutValue,
	}

	messageResponse, err := receiver.queueClient.DequeueMessage(context.Background(), &options)

	if err != nil {
		slog.Error("Unable to dequeue messages", slog.Any(utils.ErrorKey, err))
		return err
	}

	for _, dequeuedMessage := range messageResponse.Messages {
		message := *dequeuedMessage
		slog.Info("Dequeued message", slog.Any("next visible", message.TimeNextVisible), slog.Any("expiration", message.ExpirationTime), slog.Any("message", message.MessageText))
		go func() {
			err := receiver.handleMessage(message)
			if err != nil {
				slog.Error("Unable to handle message", slog.Any(utils.ErrorKey, err))
			}
		}()
	}

	return nil
}
