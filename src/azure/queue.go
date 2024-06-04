package azure

import (
	"context"
	"encoding/base64"
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

func ListenToQueue() error {

	azureQueueConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")

	//TODO: Revisit options to review queue settings
	client, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, "flexion-local", nil)
	// TODO - bubble up error and do correct logging
	if err != nil {
		slog.Error("Unable to create Azure Queue Client", err)
		return err
	}
	ctx := context.Background()

	// The code below shows how a client or server can determine the approximate count of messages in the queue:
	resp, err := client.PeekMessages(ctx, nil)
	if err != nil {
		slog.Error("Unable to peek messages", err)
		return err
	}

	messageCount := len(resp.Messages)
	slog.Info("Updated number of messages in the queue=", messageCount)
	if messageCount > 0 {
		// TODO - dequeue multiple messages, loop over them and kick off go routine for each
		messageResponse, err := client.DequeueMessage(ctx, nil)
		if err != nil {
			slog.Error("Unable to dequeue messages", err)
			return err
		}

		message := messageResponse.Messages[0]
		messageText := *message.MessageText
		messageId := *message.MessageID
		popReceipt := *message.PopReceipt

		decoded, err := base64.StdEncoding.DecodeString(messageText)
		if err != nil {
			slog.Error("Unable to decode message text", err)
			return err
		}

		slog.Info("message dequeued", slog.Any("message text", decoded), slog.Any("message response", messageResponse.Messages[0]))

		// TODO - delete message in the message handler instead of the queue reader
		deleteResponse, err := client.DeleteMessage(ctx, messageId, popReceipt, nil)
		if err != nil {
			slog.Error("Unable to delete message", err)
			return err
		}

		slog.Info("message deleted", slog.Any("delete message response", deleteResponse))
	}
	return nil
}
