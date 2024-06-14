package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type StorageHandler struct {
	blobClient *azblob.Client
}

// The container name for CA will be added as part of card 1077 and will be configurable in 1081
const containerName = "sftp"

func NewStorageHandler(conn string) (StorageHandler, error) {
	blobClient, err := azblob.NewClientFromConnectionString(conn, nil)
	if err != nil {
		return StorageHandler{}, err
	}

	return StorageHandler{blobClient: blobClient}, nil
}

// FetchFile retrieves the specified blob from Azure. The `blobPath` is everything after the container in the URL
func (receiver StorageHandler) FetchFile(blobPath string) ([]byte, error) {
	streamResponse, err := receiver.blobClient.DownloadStream(context.Background(), containerName, blobPath, &azblob.DownloadStreamOptions{})
	if err != nil {
		return nil, err
	}

	retryReader := streamResponse.NewRetryReader(context.Background(), &azblob.RetryReaderOptions{})
	defer retryReader.Close()

	resp, err := io.ReadAll(retryReader)

	//borrowed from https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore#ResponseError
	//TODO - return special stuff so we know when to retry or not
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		// Handle Error
		if respErr.StatusCode == http.StatusNotFound {
			fmt.Printf("Repository could not be found: %v", respErr)
		} else if respErr.StatusCode == http.StatusForbidden {
			fmt.Printf("You do not have permission to access this repository: %v", respErr)
		} else {
			// ...
		}
	}

	return resp, err
}

func (receiver StorageHandler) MoveFile(sourceBlobUrl string, sourceBlobPath string, destinationBlobPath string) error {
	// the storage account-level azblob client doesn't have a `copy` function, so we have to use the blob-specific client instead
	blobClient, err := blob.NewClientFromConnectionString(os.Getenv("AZURE_STORAGE_CONNECTION_STRING"), containerName, destinationBlobPath, nil)
	if err != nil {
		slog.Error("Error creating blob client")
		return err
	}

	startCopy, err := blobClient.StartCopyFromURL(context.TODO(), sourceBlobUrl, nil)
	if err != nil {
		slog.Error("Error starting blob copy")
		return err
	}

	copyStatus := *startCopy.CopyStatus
	for copyStatus == blob.CopyStatusTypePending {
		time.Sleep(time.Second * 2)
		getMetadata, err := blobClient.GetProperties(context.TODO(), nil)
		if err != nil {
			slog.Error("Error during blob copy")
			return err
		}
		copyStatus = *getMetadata.CopyStatus
	}
	slog.Info("Copied blob", slog.String("source URL", sourceBlobUrl), slog.String("destination path", destinationBlobPath))

	_, err = receiver.blobClient.DeleteBlob(context.Background(), containerName, sourceBlobPath, &azblob.DeleteBlobOptions{})
	if err != nil {
		slog.Error("Error deleting source file after copy", slog.String("source URL", sourceBlobUrl))
		return err
	}
	return nil
}
