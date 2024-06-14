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

func (receiver StorageHandler) MoveFile(sourceBlobUrl string, destinationBlobPath string) error {
	// the storage account-level azblob client doesn't have a `copy` function, so we have to use the blob-specific client instead
	blobClient, err := blob.NewClientFromConnectionString(os.Getenv("AZURE_STORAGE_CONNECTION_STRING"), containerName, destinationBlobPath, nil)
	if err != nil {
		slog.Error("Error creating blob client")
		return err
	}

	// TODO - how to get source blob URL from source blob path?
	src := "https://cdn2.auth0.com/docs/media/addons/azure_blob.svg"
	startCopy, err := blobClient.StartCopyFromURL(context.TODO(), sourceBlobUrl, nil)
	if err != nil {
		slog.Error("Error starting blob copy")
		return err
	}

	copyID := *startCopy.CopyID
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

	// TODO - what next?
	fmt.Printf("Copy from %s to %s: ID=%s, Status=%s\n", src, blobClient.URL(), copyID, copyStatus)

	/* Notes on above:
	- it feels icky to be creating a new client for each `move` action
	- but it probably feels even worse to download a file just to upload it again
	- not totally sure if `CopyFromUrl` would work for us - don't know if it needs to be a public URL?
		Might just need to update TF permissions
	- we can't use the `blob` client created in this method as a general client because it is blob-specific, so we
		can't just swap this new buddy into NewStorageHandler
	- we don't currently have env vars for account name and key. Could use `blob.NewClientFromConnectionString`, but
		that will also need a blob URL
	- if we have to use blob URLs (rather than paths), we're going to have to either build them from pieces, or get the
		URL out of the file create event (which means going back to parsing the `data` object)

	*/
	return nil
}
