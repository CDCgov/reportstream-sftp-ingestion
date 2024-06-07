package orchestration

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_getFilePathFromMessage_returnsFilePathWhenSubjectIsValid(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actual, err := getFilePathFromMessage(messageBody)
	expected := "customer/import/msg2.hl7"
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func Test_getFilePathFromMessage_returnsErrorWhenSubjectDoesNotContainBlobs(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actual, err := getFilePathFromMessage(messageBody)
	expected := ""
	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

func Test_getFilePathFromMessage_returnsErrorWhenSubjectContainsMoreThanOneBlobs(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))

	actual, err := getFilePathFromMessage(messageBody)
	expected := ""
	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

func Test_getFilePathFromMessage_returnsErrorWhenMessageNotEvent(t *testing.T) {
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06\"}"

	actual, err := getFilePathFromMessage(messageText)
	expected := ""
	assert.Error(t, err)
	assert.Equal(t, expected, actual)
}

type MockQueueClient struct {
	mock.Mock
}

// receiver.ctx, messageId, popReceipt, nil)
func (receiver MockQueueClient) DeleteMessage(ctx context.Context, messageID string, popReceipt string, o *azqueue.DeleteMessageOptions) (azqueue.DeleteMessageResponse, error) {
	args := receiver.Called(ctx, messageID, popReceipt, o)
	return args.Get(0).(azqueue.DeleteMessageResponse), args.Error(1)
}
func (receiver MockQueueClient) DequeueMessage(ctx context.Context, o *azqueue.DequeueMessageOptions) (azqueue.DequeueMessagesResponse, error) {
	return azqueue.DequeueMessagesResponse{}, nil
}

func Test_deleteMessage_returnNilWhenMessageCanBeDeleted(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)
	queueHandler := QueueHandler{queueClient: mockQueueClient, ctx: context.Background()}

	messageId := "1234"
	popReceipt := "abcd"
	message := azqueue.DequeuedMessage{MessageID: &messageId, PopReceipt: &popReceipt}

	err := queueHandler.deleteMessage(message)

	assert.NoError(t, err)
}

func Test_deleteMessage_returnErrorWhenMessageCannotBeDeleted(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, errors.New("unable to delete message"))
	queueHandler := QueueHandler{queueClient: mockQueueClient, ctx: context.Background()}

	messageId := "1234"
	popReceipt := "abcd"
	message := azqueue.DequeuedMessage{MessageID: &messageId, PopReceipt: &popReceipt}

	err := queueHandler.deleteMessage(message)

	assert.Error(t, err)
}

type MockReadAndSendUsecase struct {
	mock.Mock
}

func (receiver MockReadAndSendUsecase) NewReadAndSendUsecase() (MockReadAndSendUsecase, error) {
	return MockReadAndSendUsecase{}, nil
}
func (receiver MockReadAndSendUsecase) ReadAndSend(filepath string) error {
	return nil
}

func Test_handleMessage_returnNilWhenMessageHandledSuccessfully(t *testing.T) {
	mockQueueClient := MockQueueClient{}
	mockQueueClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(azqueue.DeleteMessageResponse{}, nil)

	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("NewReadAndSendUsecase").Return(MockReadAndSendUsecase{}, nil)
	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	queueHandler := QueueHandler{queueClient: mockQueueClient, ctx: context.Background(), usecase: mockReadAndSendUsecase}

	messageId := "1234"
	popReceipt := "abcd"
	messageText := "{\"topic\":\"/subscriptions/123/resourceGroups/resourceGroup/providers/Microsoft.Storage/storageAccounts/storageAccount\",\"subject\":\"/blobServices/default/containers/container/blobs/customer/import/msg2.hl7\",\"eventType\":\"Microsoft.Storage.BlobCreated\",\"id\":\"1234\",\"data\":{\"api\":\"PutBlob\",\"clientRequestId\":\"abcd\",\"requestId\":\"efghi\",\"eTag\":\"0x123\",\"contentType\":\"application/octet-stream\",\"contentLength\":1122,\"blobType\":\"BlockBlob\",\"url\":\"https://cdcrssftpinternal.blob.core.windows.net/container/customer/import/msg2.hl7\",\"sequencer\":\"000\",\"storageDiagnostics\":{\"batchId\":\"00000\"}},\"dataVersion\":\"\",\"metadataVersion\":\"1\",\"eventTime\":\"2024-06-06T19:57:35.6993902Z\"}"
	messageBody := base64.StdEncoding.EncodeToString([]byte(messageText))
	message := azqueue.DequeuedMessage{MessageID: &messageId, PopReceipt: &popReceipt, MessageText: &messageBody}

	err := queueHandler.handleMessage(message)

	assert.NoError(t, err)
}

// TODO - add handleMessage error scenario tests
