package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"io"
)

type BlobHandler struct {
	blobClient *azblob.Client
}

func NewBlobHandler(conn string) (BlobHandler, error) {
	blobClient, err := azblob.NewClientFromConnectionString(conn, nil)
	if err != nil {
		return BlobHandler{}, err
	}

	return BlobHandler{blobClient: blobClient}, nil
}

// TODO - container should eventually be managed by Terraform

func (receiver BlobHandler) FetchFile(blobPath string) ([]byte, error) {
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
