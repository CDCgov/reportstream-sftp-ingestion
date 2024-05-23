package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"io"
)

type StorageHandler struct {
	blobClient *azblob.Client
}

func NewStorageHandler(conn string) (StorageHandler, error) {
	blobClient, err := azblob.NewClientFromConnectionString(conn, nil)
	if err != nil {
		return StorageHandler{}, err
	}

	return StorageHandler{blobClient: blobClient}, nil
}

// TODO - container should eventually be managed by Terraform

func (receiver StorageHandler) FetchFile(blobPath string) ([]byte, error) {
	// TODO - read containerName from env vars
	containerName := "sftp"

	streamResponse, err := receiver.blobClient.DownloadStream(context.Background(), containerName, blobPath, &azblob.DownloadStreamOptions{})
	if err != nil {
		return nil, err
	}

	retryReader := streamResponse.NewRetryReader(context.Background(), &azblob.RetryReaderOptions{})
	defer retryReader.Close()

	return io.ReadAll(retryReader)
}
