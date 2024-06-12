package storage

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"io"
	"log/slog"
	"os"
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

// FetchFile retrieves the specified blob from Azure. The `blobPath` is everything after the container in the URL
func (receiver StorageHandler) FetchFile(blobPath string) ([]byte, error) {
	// The container name for CA will be added as part of card 1077 and will be configurable in 1081
	containerName := "sftp"

	streamResponse, err := receiver.blobClient.DownloadStream(context.Background(), containerName, blobPath, &azblob.DownloadStreamOptions{})
	if err != nil {
		return nil, err
	}

	retryReader := streamResponse.NewRetryReader(context.Background(), &azblob.RetryReaderOptions{})
	defer retryReader.Close()

	return io.ReadAll(retryReader)
}

func (receiver StorageHandler) MoveFile(sourceBlobPath string, destinationBlobPath string) error {

	// Borrowed from https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/storage/azblob/blob/examples_test.go
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"), os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	// Create a containerClient object to a container where we'll create a blob and its snapshot.
	// Create a blockBlobClient object to a blob in the container.
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/mycontainer/CopiedBlob.bin", accountName)
	credential, err := blob.NewSharedKeyCredential(accountName, accountKey)

	blobClient, err := blob.NewClientWithSharedKeyCredential(blobURL, credential, nil)
	if err != nil {
		slog.Error("Error")
		return err
	}
	src := "https://cdn2.auth0.com/docs/media/addons/azure_blob.svg"
	startCopy, err := blobClient.StartCopyFromURL(context.TODO(), src, nil)

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
