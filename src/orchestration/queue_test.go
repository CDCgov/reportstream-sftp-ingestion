package orchestration

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func Test_getUrlFromMessage_returnsUrlWhenDataIsValid(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedUrl := "https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7"
	assert.Nil(t, err)
	assert.Equal(t, expectedUrl, actualUrl)
}

func Test_getUrlFromMessage_returnsErrorWhenMessageCannotBeDecoded(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"

	actualUrl, err := getUrlFromMessage(messageText)
	expectedError := "illegal base64 data at input byte 0"

	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_getUrlFromMessage_returnsErrorWhenDataIsNotAMap(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":\"the data\",\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedError := "could not assert event data to a map"

	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_getUrlFromMessage_returnsErrorWhenUrlIsMissing(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedError := "could not assert event data url to a string"

	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_getUrlFromMessage_returnsErrorWhenMessageCannotUnmarshal(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actualUrl, err := getUrlFromMessage(messageBody)
	expectedError := "unmarshalling type *azeventgrid.Event"
	assert.Error(t, err)
	assert.Empty(t, actualUrl)
	assert.Contains(t, err.Error(), expectedError)
}

func Test_deleteMessage_returnNilWhenMessageCanBeDeleted(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background()}

	message := createGoodMessage()

	err := queueHandler.deleteMessage(message)

	assert.NoError(t, err)
}

func Test_deleteMessage_returnErrorWhenMessageCannotBeDeleted(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("unable to delete message"))
	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background()}

	message := createGoodMessage()

	err := queueHandler.deleteMessage(message)

	assert.Error(t, err)
}

func Test_handleMessage_returnNilWhenMessageHandledSuccessfully(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background(), usecase: &mockReadAndSendUsecase}

	message := createGoodMessage()

	err := queueHandler.handleMessage(message)

	assert.NoError(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockQueueClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_handleMessage_returnErrorWhenFailedToGetFileUrl(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background(), usecase: &mockReadAndSendUsecase}

	message := createBadMessage()

	err := queueHandler.handleMessage(message)

	assert.Error(t, err)
	mockReadAndSendUsecase.AssertNotCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockQueueClient.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_handleMessage_returnErrorWhenFailureWithDeleteMessage(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("unable to delete message"))

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background(), usecase: &mockReadAndSendUsecase}

	message := createGoodMessage()

	err := queueHandler.handleMessage(message)

	assert.Error(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
}

func Test_handleMessage_returnErrorWhenFailureWithReadAndSend(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(errors.New("failed to read and send"))
	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background(), usecase: &mockReadAndSendUsecase}

	message := createGoodMessage()

	err := queueHandler.handleMessage(message)

	assert.NoError(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
	mockQueueClient.AssertNotCalled(t, "DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ReceiveQueue_HappyPath(t *testing.T) {
	// Setup for DequeueMessage
	mockQueueClient := MockQueueClient{}
	message := createGoodMessage()
	dequeuedMessageResponse := azqueue.DequeueMessagesResponse{Messages: []*azqueue.DequeuedMessage{&message}}
	mockQueueClient.On("DequeueMessage", mock.Anything, mock.Anything).Return(dequeuedMessageResponse, nil)

	// Setup for handleMessage (to avoid adding otherwise unneeded interface for QueueHandler)
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)

	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background(), usecase: &mockReadAndSendUsecase}
	err := queueHandler.receiveQueue()

	mockQueueClient.AssertCalled(t, "DequeueMessage", mock.Anything, mock.Anything)
	assert.NoError(t, err)
}

func Test_ReceiveQueue_UnableToDequeueMessage(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DequeueMessage", mock.Anything, mock.Anything).Return(azqueue.DequeueMessagesResponse{}, errors.New("dequeue message failed"))

	mockReadAndSendUsecase := MockReadAndSendUsecase{}
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)

	queueHandler := QueueHandler{queueClient: &mockQueueClient, ctx: context.Background(), usecase: &mockReadAndSendUsecase}
	err := queueHandler.receiveQueue()

	assert.Error(t, err)

}

func Test_checkDeliveryAttempts_returnNilWhenDeliveryCountParsedAndUnderDequeueThreshold(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "5")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	message := createGoodMessage()

	err := checkDeliveryAttempts(message)

	assert.NoError(t, err)
}

func Test_checkDeliveryAttempts_returnErrorWhenDeliveryCountParsedAndOverDequeueThreshold(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "5")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	message := createMessageOverDequeueThreshold()

	err := checkDeliveryAttempts(message)

	assert.Error(t, err)
	assert.NotContains(t, err.Error(), cannotParseMessage)
	assert.Contains(t, err.Error(), overDequeueMessage)
}

func Test_checkDeliveryAttempts_returnErrorWhenDeliveryCountCannotParseAndUnderDequeueThreshold(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "Five")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	message := createGoodMessage()

	err := checkDeliveryAttempts(message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), cannotParseMessage)
	assert.NotContains(t, err.Error(), overDequeueMessage)
}

func Test_checkDeliveryAttempts_returnErrorWhenDeliveryCountCannotParseAndOverDequeueThreshold(t *testing.T) {
	os.Setenv("QUEUE_MAX_DELIVERY_ATTEMPTS", "Five")
	defer os.Unsetenv("QUEUE_MAX_DELIVERY_ATTEMPTS")

	message := createMessageOverDequeueThreshold()

	err := checkDeliveryAttempts(message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), cannotParseMessage)
	assert.Contains(t, err.Error(), overDequeueMessage)
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

type MockReadAndSendUsecase struct {
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

func (receiver *MockReadAndSendUsecase) ReadAndSend(sourceUrl string) error {
	args := receiver.Called(sourceUrl)
	return args.Error(0)
}

// Constants for tests
const cannotParseMessage = "Failed to parse QUEUE_MAX_DELIVERY_ATTEMPTS"
const overDequeueMessage = "Message reached maximum number of delivery attempts"
