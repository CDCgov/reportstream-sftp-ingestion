package orchestration

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"log/slog"
	"os"
	"time"
)

/*
	func listenToQueue(sqsService *sqs.SQS, queueUrl *string) {
		log.Printf("Starting to listen to queue %s", *queueUrl)
		for {
			queueMessages, err := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl: queueUrl,
				MaxNumberOfMessages: aws.Int64(5),
				WaitTimeSeconds: aws.Int64(5),
			})
			if err != nil {
				log.Printf("AWS SQS queue messages weren't able to be retrieved; %+v", err)
			}

			for _, message := range queueMessages.Messages {
				go handleQueueMessage(message, sqsService, queueUrl)
			}
		}
	}
*/
type QueueHandler struct {
	queueClient *azqueue.QueueClient
	ctx         context.Context
}

func NewQueueHandler() (QueueHandler, error) {
	azureQueueConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")

	//TODO: Revisit options to review queue settings
	client, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, "flexion-local", nil)
	// TODO - bubble up error and do correct logging
	if err != nil {
		slog.Error("Unable to create Azure Queue Client", err)
		return QueueHandler{}, err
	}

	return QueueHandler{queueClient: client, ctx: context.Background()}, nil
}

func (receiver QueueHandler) handleMessage(message azqueue.DequeuedMessage) error {
	// TODO - use event schema: https://learn.microsoft.com/en-us/azure/event-grid/event-schema-blob-storage?toc=%2Fazure%2Fstorage%2Fblobs%2Ftoc.json&tabs=cloud-event-schema

	messageText := *message.MessageText

	usecase, err := usecases.NewReadAndSendUsecase()
	if err != nil {
		slog.Warn("Failed to init the usecase", slog.Any("error", err))
		slog.Info("Continuing for now while debugging")
	}
	// changing from a message body that looks like 'msg2.hl7' to a message body that's a json
	eventBytes, err := base64.StdEncoding.DecodeString(messageText)

	// Map bytes json to Event object format (shape)
	var event azeventgrid.Event
	err = event.UnmarshalJSON(eventBytes)
	if err != nil {
		slog.Error("Failed to unmarshal event", slog.Any("error", err))
		return err
	}

	slog.Info("Event", slog.Any("body", event))

	slog.Info("Event", slog.Any("data", event.Data))

	// Data is an 'any' type. We need to tell Go that it's a map
	eventData, ok := event.Data.(map[string]any)

	if !ok {
		slog.Error("Could not assert event data to a map", slog.Any("event", event))
		return errors.New("could not assert event data to a map")
	}

	// Extract blob url from Event's data
	eventUrl, ok := eventData["url"].(string)

	if !ok {
		slog.Error("Could not assert event data url to a string", slog.Any("event", event))
		return errors.New("could not assert event data url to a string")
	}

	// TODO - update readandsend to use whole blob url instead of just filepath
	err = usecase.ReadAndSend(eventUrl)

	// TODO - how do we decide when to move a file from import to failure/error?
	// If a queue message ends up on the poison queue, should the file still be in `import`
	// or should we know to move it to `error`? Does it matter if it's e.g. a non-success response from RS
	// vs an error calling them?
	// TODO - minimum option is to check the dequeue count, and if we're over the threshold, log an error so we at least know something failed
	if err != nil {
		slog.Warn("Failed to read/send file", slog.Any("error", err))
	} else {
		// Only delete message if file successfully sent to ReportStream
		messageId := *message.MessageID
		popReceipt := *message.PopReceipt

		deleteResponse, err := receiver.queueClient.DeleteMessage(receiver.ctx, messageId, popReceipt, nil)
		if err != nil {
			slog.Error("Unable to delete message", slog.Any("error", err))
			return err
		}

		slog.Info("message deleted", slog.Any("delete message response", deleteResponse))
	}

	return nil
}

func (receiver QueueHandler) ListenToQueue() error {
	for {
		messageResponse, err := receiver.queueClient.DequeueMessage(receiver.ctx, nil)
		if err != nil {
			slog.Error("Unable to dequeue messages", err)
			return err
		}
		// TODO - dequeue multiple messages, loop over them and kick off go routine for each
		var messageCount int

		messageCount = len(messageResponse.Messages)
		slog.Info("", slog.Any("Number of messages in the queue", messageCount))

		if messageCount > 0 {
			message := *messageResponse.Messages[0]
			go func() {
				err := receiver.handleMessage(message)
				if err != nil {
					slog.Error("Unable to handle message", err)
					// TODO - decide what to do with errored messages
				}
			}()
		}
		time.Sleep(10 * time.Second)
	}
}
