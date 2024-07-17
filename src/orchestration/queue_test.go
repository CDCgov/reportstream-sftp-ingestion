package orchestration

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"os"
	"testing"
	"time"
)

func Test_getUrlFromMessage_DataIsValid_ReturnsUrl(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedUrl := "https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7"
	assert.Nil(t, err)
	assert.Equal(t, expectedUrl, actualUrl)
}

func Test_getUrlFromMessage_MessageCannotBeDecoded_ReturnsError(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"

	actualUrl, err := getUrlFromMessage(messageText)
	expectedError := "illegal base64 data at input byte 0"

	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_getUrlFromMessage_DataIsNotAMap_ReturnsError(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":\"the data\",\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedError := "could not assert event data to a map"

	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_getUrlFromMessage_UrlIsMissing_ReturnsError(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedError := "could not assert event data url to a string"

	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_getUrlFromMessage_MessageCannotUnmarshal_ReturnsError(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedError := "unmarshalling type *azeventgrid.Event"
	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_deleteMessage_MessageCanBeDeleted_DeletesMessage(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	queueHandler := QueueHandler{queueClient: &mockQueueClient}

	message := createGoodMessage()

	err := queueHandler.deleteMessage(message)

	assert.NoError(t, err)
}

func Test_deleteMessage_MessageCannotBeDeleted_ReturnsError(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("unable to delete message"))
	queueHandler := QueueHandler{queueClient: &mockQueueClient}

	message := createGoodMessage()

	err := queueHandler.deleteMessage(message)

	assert.Error(t, err)
}

func Test_handleMessage_MessageHandledSuccessfully_ReturnNil(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}

	message := createGoodMessage()

	err := queueHandler.handleMessage(message)

	assert.NoError(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockQueueClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_handleMessage_FailedToGetFileUrl_DoesNotCallReadAndSendOrDeleteMessage(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}

	message := createBadMessage()

	err := queueHandler.handleMessage(message)

	assert.NoError(t, err)
	mockReadAndSendUsecase.AssertNotCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockQueueClient.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_handleMessage_FailureWithDeleteMessage_ReturnsError(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("unable to delete message"))

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}

	message := createGoodMessage()

	err := queueHandler.handleMessage(message)

	assert.Error(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
}

func Test_handleMessage_FailureWithReadAndSend_ReturnsError(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(errors.New("failed to read and send"))
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}

	message := createGoodMessage()

	err := queueHandler.handleMessage(message)

	assert.NoError(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockQueueClient.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_handleMessage_OverDequeueThreshold_ReturnsError(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler, deadLetterQueueClient: &mockDeadLetterQueueClient}

	message := createMessageOverDequeueThreshold()

	err := queueHandler.handleMessage(message)

	assert.Error(t, err)
	mockReadAndSendUsecase.AssertNotCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockDeadLetterQueueClient.AssertCalled(t, "EnqueueMessage", mock.Anything, mock.Anything, mock.Anything)
	mockQueueClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ReceiveQueue_OnSuccess_ReturnsNil(t *testing.T) {
	// Setup for DequeueMessage
	mockQueueClient := MockQueueClient{}
	message := createGoodMessage()
	dequeuedMessageResponse := azqueue.DequeueMessagesResponse{Messages: []*azqueue.DequeuedMessage{&message}}
	mockQueueClient.On("DequeueMessage", mock.Anything, mock.Anything).Return(dequeuedMessageResponse, nil)

	// Setup for handleMessage (to avoid adding otherwise unneeded interface for QueueHandler)
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)

	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}
	err := queueHandler.receiveQueue()

	mockQueueClient.AssertCalled(t, "DequeueMessage", mock.Anything, mock.Anything)
	assert.NoError(t, err)
}

func Test_ReceiveQueue_UnableToDequeueMessage_ReturnsError(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DequeueMessage", mock.Anything, mock.Anything).Return(azqueue.DequeueMessagesResponse{}, errors.New("dequeue message failed"))

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)

	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}
	err := queueHandler.receiveQueue()

	assert.Error(t, err)
}

func Test_ReceiveQueue_UnableToHandleMessage_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	// Setup for DequeueMessage
	mockQueueClient := MockQueueClient{}
	message := createGoodMessage()
	dequeuedMessageResponse := azqueue.DequeueMessagesResponse{Messages: []*azqueue.DequeuedMessage{&message}}
	mockQueueClient.On("DequeueMessage", mock.Anything, mock.Anything).Return(dequeuedMessageResponse, nil)

	// Setup for handleMessage (to avoid adding otherwise unneeded interface for QueueHandler)
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("failed to delete"))

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)

	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}
	err := queueHandler.receiveQueue()

	mockQueueClient.AssertCalled(t, "DequeueMessage", mock.Anything, mock.Anything)
	assert.NoError(t, err)

	// The `slog.Error` call we're checking for happens in a GoRoutine, which completes immediately after the
	// receiveQueue function. Since no production code is called after this GoRoutine is done, there is no race
	// condition to worry about, and we can just wait a short time in this test to ensure all calls are completed
	time.Sleep(1 * time.Second)
	assert.Contains(t, buffer.String(), "Unable to handle message")
}

func Test_ReceiveQueue_QueueContainsMultipleMessages_HandlesAllMessages(t *testing.T) {

	// Setup for DequeueMessage
	mockQueueClient := MockQueueClient{}
	message1 := createGoodMessage()
	message2 := createGoodMessage()
	message3 := createGoodMessage()
	messages := []*azqueue.DequeuedMessage{&message1, &message2, &message3}
	dequeuedMessageResponse := azqueue.DequeueMessagesResponse{Messages: messages}
	mockQueueClient.On("DequeueMessage", mock.Anything, mock.Anything).Return(dequeuedMessageResponse, nil)

	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)

	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}
	err := queueHandler.receiveQueue()

	assert.NoError(t, err)

	// The assertions below happens in a GoRoutine, which completes immediately after the
	// receiveQueue function. Since no production code is called after this GoRoutine is done, there is no race
	// condition to worry about, and we can just wait a short time in this test to ensure all calls are completed
	time.Sleep(1 * time.Second)
	mockReadAndSendUsecase.AssertNumberOfCalls(t, "ReadAndSend", len(messages))
}

func Test_overDeliveryThreshold_DeliveryCountParsedAndUnderDequeueThreshold_ReturnsFalse(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "5")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockQueueClient := MockQueueClient{}
	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}

	message := createGoodMessage()

	overThreshold := queueHandler.overDeliveryThreshold(message)

	assert.NotContains(t, buffer.String(), cannotParseMessage)
	assert.NotContains(t, buffer.String(), overDequeueMessage)
	assert.NotContains(t, buffer.String(), "Failed to move message to the DLQ")
	assert.Equal(t, false, overThreshold)
}

func Test_overDeliveryThreshold_DeliveryCountParsedAndOverDequeueThreshold_ReturnsTrue(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "5")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, deadLetterQueueClient: &mockDeadLetterQueueClient, messageContentHandler: importMessageHandler}

	message := createMessageOverDequeueThreshold()
	overThreshold := queueHandler.overDeliveryThreshold(message)

	assert.NotContains(t, buffer.String(), cannotParseMessage)
	assert.Contains(t, buffer.String(), overDequeueMessage)
	assert.NotContains(t, buffer.String(), "Failed to move message to the DLQ")
	assert.Contains(t, buffer.String(), "Successfully moved the message to the DLQ")
	assert.Equal(t, true, overThreshold)
}

func Test_overDeliveryThreshold_DequeueThresholdCannotBeParsedAndAttemptsUnderDefaultThreshold_ReturnsFalse(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "Five")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockQueueClient := MockQueueClient{}
	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, messageContentHandler: importMessageHandler}

	message := createGoodMessage()

	overThreshold := queueHandler.overDeliveryThreshold(message)

	assert.Contains(t, buffer.String(), cannotParseMessage)
	assert.NotContains(t, buffer.String(), overDequeueMessage)
	assert.NotContains(t, buffer.String(), "Failed to move message to the DLQ")
	assert.Equal(t, false, overThreshold)
}

func Test_overDeliveryThreshold_DequeueThresholdCannotBeParsedAndAttemptsOverDefaultThreshold(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "Five")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}
	queueHandler := QueueHandler{queueClient: &mockQueueClient, deadLetterQueueClient: &mockDeadLetterQueueClient, messageContentHandler: importMessageHandler}

	message := createMessageOverDequeueThreshold()
	overThreshold := queueHandler.overDeliveryThreshold(message)

	assert.Contains(t, buffer.String(), cannotParseMessage)
	assert.Contains(t, buffer.String(), overDequeueMessage)
	assert.NotContains(t, buffer.String(), "Failed to move message to the DLQ")
	assert.Contains(t, buffer.String(), "Successfully moved the message to the DLQ")
	assert.Equal(t, true, overThreshold)
}

func Test_overDeliveryThreshold_OverThresholdAndUnableToDeadLetter_ReturnsTrue(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "5")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, errors.New("DLQ failed"))

	queueHandler := QueueHandler{queueClient: &mockQueueClient, deadLetterQueueClient: &mockDeadLetterQueueClient}

	message := createMessageOverDequeueThreshold()
	overThreshold := queueHandler.overDeliveryThreshold(message)

	assert.Contains(t, buffer.String(), "Failed to move message to the DLQ")
	assert.Equal(t, true, overThreshold)
}

func Test_deadLetter_MessageAddedToDLQAndSuccessfullyDeletedMessageFromOriginalQueue_ReturnsNil(t *testing.T) {
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, nil)

	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	queueHandler := QueueHandler{queueClient: &mockQueueClient, deadLetterQueueClient: &mockDeadLetterQueueClient}

	message := createMessageOverDequeueThreshold()
	err := queueHandler.deadLetter(message)

	assert.NoError(t, err)
	mockDeadLetterQueueClient.AssertCalled(t, "EnqueueMessage", mock.Anything, mock.Anything, mock.Anything)
	mockQueueClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_deadLetter_MessageCannotBeAddedToDLQ_ReturnsError(t *testing.T) {
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, errors.New("couldn't enqueue message"))

	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	queueHandler := QueueHandler{queueClient: &mockQueueClient, deadLetterQueueClient: &mockDeadLetterQueueClient}

	message := createMessageOverDequeueThreshold()
	err := queueHandler.deadLetter(message)

	assert.Error(t, err)
	mockDeadLetterQueueClient.AssertCalled(t, "EnqueueMessage", mock.Anything, mock.Anything, mock.Anything)
	mockQueueClient.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_deadLetter_MessageCannotBeDeletedFromOriginalQueue_ReturnsError(t *testing.T) {
	mockDeadLetterQueueClient := MockQueueClient{}
	mockDeadLetterQueueClient.On("EnqueueMessage", mock.Anything, mock.Anything, mock.Anything).Return(azqueue.EnqueueMessagesResponse{}, nil)

	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("couldn't delete message from original queue"))

	queueHandler := QueueHandler{queueClient: &mockQueueClient, deadLetterQueueClient: &mockDeadLetterQueueClient}

	message := createMessageOverDequeueThreshold()
	err := queueHandler.deadLetter(message)

	assert.Error(t, err)
	mockDeadLetterQueueClient.AssertCalled(t, "EnqueueMessage", mock.Anything, mock.Anything, mock.Anything)
	mockQueueClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func createMessageOverDequeueThreshold() azqueue.DequeuedMessage {
	messageId := "1234"
	popReceipt := "abcd"
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))
	var dequeueCount int64 = 6
	message := azqueue.DequeuedMessage{MessageID: &messageId, PopReceipt: &popReceipt, MessageText: &messageBody, DequeueCount: &dequeueCount}
	return message
}

// Helper functions for tests
func createGoodMessage() azqueue.DequeuedMessage {
	messageId := "1234"
	popReceipt := "abcd"
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))
	var dequeueCount int64 = 4
	message := azqueue.DequeuedMessage{MessageID: &messageId, PopReceipt: &popReceipt, MessageText: &messageBody, DequeueCount: &dequeueCount}
	return message
}

func createBadMessage() azqueue.DequeuedMessage {
	messageId := "1234"
	popReceipt := "abcd"
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))
	var dequeueCount int64 = 4
	message := azqueue.DequeuedMessage{MessageID: &messageId, PopReceipt: &popReceipt, MessageText: &messageBody, DequeueCount: &dequeueCount}
	return message
}

// Mocks for test

type MockQueueClient struct {
	mock.Mock
}

func (receiver *MockQueueClient) DeleteMessage(ctx context.Context, messageID string, popReceipt string, o *azqueue.DeleteMessageOptions) (azqueue.DeleteMessageResponse, error) {
	args := receiver.Called(ctx, messageID, popReceipt, o)
	return args.Get(0).(azqueue.DeleteMessageResponse), args.Error(1)
}
func (receiver *MockQueueClient) DequeueMessage(ctx context.Context, o *azqueue.DequeueMessageOptions) (azqueue.DequeueMessagesResponse, error) {
	args := receiver.Called(ctx, o)
	return args.Get(0).(azqueue.DequeueMessagesResponse), args.Error(1)
}
func (receiver *MockQueueClient) EnqueueMessage(ctx context.Context, content string, o *azqueue.EnqueueMessageOptions) (azqueue.EnqueueMessagesResponse, error) {
	args := receiver.Called(ctx, content, o)
	return args.Get(0).(azqueue.EnqueueMessagesResponse), args.Error(1)
}

type MockReadAndSendUsecase struct {
	mock.Mock
}

func (receiver *MockReadAndSendUsecase) ReadAndSend(sourceUrl string) error {
	args := receiver.Called(sourceUrl)
	return args.Error(0)
}

// Constants for tests
const cannotParseMessage = "Failed to parse QUEUE_MAX_DELIVERY_ATTEMPTS, defaulting to 5"
const overDequeueMessage = "Message reached maximum number of delivery attempts"
