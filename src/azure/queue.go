package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"log"
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

func ListenToQueue() {

	azureQueueConnectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")

	//TODO: Revisit options to review queue settings
	client, err := azqueue.NewQueueClientFromConnectionString(azureQueueConnectionString, "flexion-local", nil)
	// TODO - bubble up error and do correct logging
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// The code below shows how a client or server can determine the approximate count of messages in the queue:
	resp, err := client.PeekMessages(ctx, nil)

	if err != nil {
		log.Fatal(err)
	}
	messageCount := len(resp.Messages)
	fmt.Printf("Updated number of messages in the queue=%d\n", messageCount)
	if messageCount > 0 {
		// TODO - dequeue multiple messages, loop over them and kick off go routine for each
		messageResponse, err := client.DequeueMessage(ctx, nil)
		if err != nil {
			log.Fatal(err)
		}

		message := messageResponse.Messages[0]
		messageText := *message.MessageText
		messageId := *message.MessageID
		popReceipt := *message.PopReceipt

		decoded, err := base64.StdEncoding.DecodeString(messageText)
		if err != nil {
			fmt.Println("decode error:", err)
			return
		}
		fmt.Println(string(decoded))

		slog.Info("message dequeued", slog.Any("message text", decoded), slog.Any("message response", messageResponse.Messages[0]))

		// TODO - delete message in the message handler instead of the queue reader
		deleteResponse, err := client.DeleteMessage(ctx, messageId, popReceipt, nil)

		if err != nil {
			log.Fatal(err)
		}

		slog.Info("message deleted", slog.Any("delete message response", deleteResponse))
	}

}
