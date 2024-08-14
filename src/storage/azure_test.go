package storage

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

const connectionString = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://sftp-Azurite:10000/devstoreaccount1;QueueEndpoint=http://sftp-Azurite:10001/devstoreaccount1"

func Test_NewAzureBlobHandler_returnsAzureBlobHandler(t *testing.T) {
	os.Setenv("AZURE_STORAGE_CONNECTION_STRING", connectionString)
	defer os.Unsetenv("AZURE_STORAGE_CONNECTION_STRING")

	azureBlobHandler, err := NewAzureBlobHandler()

	assert.NotNil(t, azureBlobHandler)
	assert.NoError(t, err)
}

func Test_NewAzureBlobHandler_UnableToGetConnectionString_returnsError(t *testing.T) {
	os.Setenv("AZURE_STORAGE_CONNECTION_STRING", "")
	defer os.Unsetenv("AZURE_STORAGE_CONNECTION_STRING")

	azureBlobHandler, err := NewAzureBlobHandler()
	expectedAzureBlobHandler := AzureBlobHandler{}

	assert.Equal(t, expectedAzureBlobHandler, azureBlobHandler)
	assert.Error(t, err)
}

//func Test_FetchFile_UnableToGetConnectionString_returnsError(t *testing.T) {
//	mockAzureBlobClient := new(MockAzureBlobClient)
//	azureBlobHandler := AzureBlobHandler{blobClient: mockAzureBlobClient}
//
//	actualByte, err := azureBlobHandler.FetchFile(utils.SuccessSourceUrl)
//
//	assert.Nil(t, actualByte)
//	assert.Error(t, err)
//}

// 	mockAzureBlobHandler.On("FetchFile", mock.Anything).Return([]byte("kys"), nil)

type MockAzureBlobClient struct {
	*mock.Mock
}

func (receiver *MockAzureBlobClient) DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
	args := receiver.Called(ctx, containerName, blobName, o)
	return azblob.DownloadStreamResponse{}, args.Error(1)
}

func (receiver *MockAzureBlobClient) UploadBuffer(ctx context.Context, containerName string, blobName string, buffer []byte, o *azblob.UploadBufferOptions) (azblob.UploadBufferResponse, error) {
	args := receiver.Called(ctx, containerName, blobName, buffer, o)
	return args.Get(0).(azblob.UploadBufferResponse), args.Error(1)
}

func (receiver *MockAzureBlobClient) DeleteBlob(ctx context.Context, containerName string, blobName string, o *azblob.DeleteBlobOptions) (azblob.DeleteBlobResponse, error) {
	args := receiver.Called(ctx, containerName, blobName, o)
	return args.Get(0).(azblob.DeleteBlobResponse), args.Error(1)
}
