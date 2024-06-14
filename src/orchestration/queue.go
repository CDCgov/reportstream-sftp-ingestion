package orchestration

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type QueueHandler struct {
	queueClient QueueClient
	ctx         context.Context
	usecase     usecases.ReadAndSend
}

type QueueClient interface {
	DeleteMessage(ctx context.Context, messageID string, popReceipt string, o *azqueue.DeleteMessageOptions) (azqueue.DeleteMessageResponse, error)
	DequeueMessage(ctx context.Context, o *azqueue.DequeueMessageOptions) (azqueue.DequeueMessagesResponse, error)
}

func NewQueueHandler() (QueueHandler, error) {
	azureQueueConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")

	client, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, "blob-message-queue", nil)
	if err != nil {
		slog.Error("Unable to create Azure Queue Client", err)
		return QueueHandler{}, err
	}

	usecase, err := usecases.NewReadAndSendUsecase()

	if err != nil {
		slog.Error("Unable to create Usecase", err)
		return QueueHandler{}, err
	}

	return QueueHandler{queueClient: client, ctx: context.Background(), usecase: &usecase}, nil
}

func getFilePathFromMessage(messageText string) (string, error) {
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

	eventSubject := *event.Subject

	eventSubjectParts := strings.Split(eventSubject, "blobs/")

	// Determines whether a blob was given and split properly
	// EX: "subject":"/blobServices/default/containers/sftp/blobs/customer/import/msg2.hl7"
	// If more than 2 pieces result, there's something confusing about the file path
	// If fewer than 2 pieces result, this is probably not a blob
	if len(eventSubjectParts) != 2 {
		slog.Error("Failed to parse subject", slog.String("subject", eventSubject))
		return "", errors.New("failed to parse subject")
	}

	return eventSubjectParts[1], nil
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
	filePath, err := getFilePathFromMessage(*message.MessageText)

	if err != nil {
		slog.Error("Failed to get the file path", slog.Any("error", err))
		return err
	}

	err = receiver.usecase.ReadAndSend(filePath)
	/*
			Four possible scenarios to handle:
			- Success from reportstream - move to 'success' folder and delete message
			- Transient error from reportstream (like a 404, 502 (bad gateway), or 503 (service unavailable) status)
				- this is something to retry, so leave message on queue and don't move file
			- Non-transient errors from reportstream (e.g. message body is wrong shape or credentials are wrong - need to
				look at their error responses). No point retrying, so move file to 'failure' folder and delete q message
			- Azure errors - these are probably transient, so leave message on queue and don't move file

		How do we decide if a reportstream error is transient?

	*/

	if err != nil {
		slog.Warn("Failed to read/send file", slog.Any("error", err))
		err := checkDeliveryAttempts(message)
		if err != nil {
			slog.Error(err.Error())
		}
	} else {
		// Only delete message if file successfully sent to ReportStream
		err = receiver.deleteMessage(message)
		if err != nil {
			slog.Warn("Failed to delete message", slog.Any("error", err))
			return err
		}
	}

	return nil
}

// checkDeliveryAttempts checks whether the max delivery attempts for the message have been reached.
// If the threshold has been reached, the message should go to dead letter storage.
func checkDeliveryAttempts(message azqueue.DequeuedMessage) error {
	maxDeliveryCount, err := strconv.ParseInt(os.Getenv("QUEUE_MAX_DELIVERY_ATTEMPTS"), 10, 64)
	errorMessage := ""
	if err != nil {
		maxDeliveryCount = 5
		errorMessage = fmt.Sprintf("Failed to parse QUEUE_MAX_DELIVERY_ATTEMPTS, defaulting to %d. Error: %s. ", maxDeliveryCount, err.Error())
	}

	if *message.DequeueCount >= maxDeliveryCount {
		errorMessage = errorMessage + fmt.Sprintf("Message reached maximum number of delivery attempts %v", message)
	}

	if errorMessage != "" {
		return errors.New(errorMessage)
	}

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
	messageResponse, err := receiver.queueClient.DequeueMessage(receiver.ctx, nil)
	if err != nil {
		slog.Error("Unable to dequeue messages", err)
		return err
	} else {
		var messageCount int

		messageCount = len(messageResponse.Messages)
		slog.Info("", slog.Any("Number of messages in the queue", messageCount))

		if messageCount > 0 {
			message := *messageResponse.Messages[0]
			go func() {
				err := receiver.handleMessage(message)
				if err != nil {
					slog.Error("Unable to handle message", err)
				}
			}()
		}
	}
	return nil
}
