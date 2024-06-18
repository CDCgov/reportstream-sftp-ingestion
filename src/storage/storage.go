package storage

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"io"
	"log/slog"
	"os"
)

type StorageHandler struct {
	blobClient *azblob.Client
}

const containerName = "sftp"

func NewStorageHandler() (StorageHandler, error) {
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	blobClient, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return StorageHandler{}, err
	}

	return StorageHandler{blobClient: blobClient}, nil
}

func (receiver StorageHandler) FetchFile(sourceUrl string) ([]byte, error) {
	sourceUrlParts, err := azblob.ParseURL(sourceUrl)
	if err != nil {
		slog.Error("Unable to parse source URL", slog.String("sourceUrl", sourceUrl))
		return nil, err
	}

	streamResponse, err := receiver.blobClient.DownloadStream(context.Background(), containerName, sourceUrlParts.BlobName, &azblob.DownloadStreamOptions{})
	if err != nil {
		return nil, err
	}

	retryReader := streamResponse.NewRetryReader(context.Background(), &azblob.RetryReaderOptions{})
	defer retryReader.Close()

	resp, err := io.ReadAll(retryReader)

	// As of June 2024, we're treating all Azure errors as transient/retryable
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		slog.Error("Error fetching file", slog.Any("error", respErr), slog.String("sourceUrl", sourceUrl))
	}

	return resp, err
}

func (receiver StorageHandler) MoveFile(sourceUrl string, destinationUrl string) error {
	sourceUrlParts, err := azblob.ParseURL(sourceUrl)
	if err != nil {
		slog.Error("Unable to parse source URL", slog.String("sourceUrl", sourceUrl))
		return err
	}
	destinationUrlParts, err := azblob.ParseURL(destinationUrl)
	if err != nil {
		slog.Error("Unable to parse destination URL", slog.String("destinationUrl", destinationUrl))
		return err
	}

	// the storage account-level azblob client used on the StorageHandler struct doesn't have a `copy` function,
	// so we have to use a blob-file-specific client for copying
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	blobClient, err := blob.NewClientFromConnectionString(connectionString, containerName, destinationUrlParts.BlobName, nil)
	if err != nil {
		slog.Error("Error creating blob client")
		return err
	}
	copyResponse, err := blobClient.CopyFromURL(context.TODO(), sourceUrl, nil)
	if err != nil {
		slog.Error("Error copying file", slog.String("sourceUrl", sourceUrl), slog.String("destinationUrl", destinationUrl), slog.Any("error", err))
		return err
	}

	//startCopy, err := blobClient.StartCopyFromURL(context.TODO(), sourceUrl, nil)
	//if err != nil {
	//	slog.Error("Error starting blob copy")
	//	return err
	//}
	//
	//copyStatus := *startCopy.CopyStatus
	//slog.Info("copy status before copy", slog.Any("copyStatus", copyStatus))
	//for copyStatus == blob.CopyStatusTypePending {
	//	time.Sleep(time.Second * 2)
	//	getMetadata, err := blobClient.GetProperties(context.TODO(), nil)
	//	if err != nil {
	//		slog.Error("Error during blob copy")
	//		return err
	//	}
	//	copyStatus = *getMetadata.CopyStatus
	//	slog.Info("copy status during copy", slog.Any("copyStatus", copyStatus))
	//	slog.Info("metadata during copy", slog.Any("getMetadata", getMetadata))
	//}
	//slog.Info("copy status after copy", slog.Any("copyStatus", copyStatus))
	slog.Info("Copied blob", slog.String("source URL", sourceUrl), slog.String("destination URL", destinationUrl), slog.Any("copyReponse", copyResponse))
	getMetadata, err := blobClient.GetProperties(context.TODO(), nil)
	slog.Info("metadata after copy", slog.Any("getMetadata", getMetadata))

	_, err = receiver.blobClient.DeleteBlob(context.Background(), containerName, sourceUrlParts.BlobName, &azblob.DeleteBlobOptions{})
	if err != nil {
		slog.Error("Error deleting source file after copy", slog.String("source URL", sourceUrl))
		return err
	}
	return nil
}
