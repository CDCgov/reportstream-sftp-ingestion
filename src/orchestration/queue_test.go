package orchestration

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
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
