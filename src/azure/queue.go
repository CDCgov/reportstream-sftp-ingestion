package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"log/slog"
	"os"
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

func (receiver QueueHandler) HandleMessage(message azqueue.DequeuedMessage) error {
	// TODO - use event schema: https://learn.microsoft.com/en-us/azure/event-grid/event-schema-blob-storage?toc=%2Fazure%2Fstorage%2Fblobs%2Ftoc.json&tabs=cloud-event-schema
	messageId := *message.MessageID
	popReceipt := *message.PopReceipt

	// TODO - delete message in the message handler instead of the queue reader
	deleteResponse, err := receiver.queueClient.DeleteMessage(receiver.ctx, messageId, popReceipt, nil)
	if err != nil {
		slog.Error("Unable to delete message", slog.Any("error", err))
		return err
	}

	slog.Info("message deleted", slog.Any("delete message response", deleteResponse))

	return nil
}

func (receiver QueueHandler) ListenToQueue() (azqueue.DequeueMessagesResponse, error) {
	dequeueMessage, err := receiver.queueClient.DequeueMessage(receiver.ctx, nil)
	if err != nil {
		slog.Error("Unable to dequeue messages", err)
		return dequeueMessage, err
	}

	return dequeueMessage, nil
}

//func

//for {
//	// TODO - dequeue multiple messages, loop over them and kick off go routine for each
//
//
//	var messageCount int
//
//	messageCount = len(messageResponse.Messages)
//	slog.Info("", slog.Any("Number of messages in the queue", messageCount))
//
//	if messageCount > 0 {
//		message := *messageResponse.Messages[0]
//		go func() {
//			err := receiver.handleMessage(message)
//			if err != nil {
//				slog.Error("Unable to handle message", err)
//			}
//
//		}()
//	}
//	time.Sleep(1 * time.Minute)
//}
//}
