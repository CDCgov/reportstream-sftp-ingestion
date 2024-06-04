package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"log"
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
	fmt.Printf("Updated number of messages in the queue=%d\n", len(resp.Messages))

}
