package storage

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"io"
	"log/slog"
	"os"
)

type StorageHandler struct {
	blobClient *azblob.Client
}

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

	streamResponse, err := receiver.blobClient.DownloadStream(context.Background(), sourceUrlParts.ContainerName, sourceUrlParts.BlobName, nil)
	if err != nil {
		return nil, err
	}

	retryReader := streamResponse.NewRetryReader(context.Background(), nil)
	defer retryReader.Close()

	resp, err := io.ReadAll(retryReader)

	return resp, err
}

func (receiver StorageHandler) MoveFile(sourceUrl string, destinationUrl string) error {
	sourceUrlParts, err := azblob.ParseURL(sourceUrl)
	if err != nil {
		slog.Error("Unable to parse source URL", slog.String("sourceUrl", sourceUrl), slog.Any("error", err))
		return err
	}
	destinationUrlParts, err := azblob.ParseURL(destinationUrl)
	if err != nil {
		slog.Error("Unable to parse destination URL", slog.String("destinationUrl", destinationUrl), slog.Any("error", err))
		return err
	}

	fileBytes, err := receiver.FetchFile(sourceUrl)
	if err != nil {
		slog.Error("Unable to fetch file", slog.String("sourceUrl", sourceUrl), slog.Any("error", err))
		return err
	}

	uploadResponse, err := receiver.blobClient.UploadBuffer(context.Background(), destinationUrlParts.ContainerName, destinationUrlParts.BlobName, fileBytes, nil)
	if err != nil {
		slog.Error("Unable to upload file", slog.String("destinationUrl", destinationUrl), slog.Any("error", err))
		return err
	}

	slog.Info("Successfully uploaded file", slog.String("destinationUrl", destinationUrl), slog.Any("uploadResponse", uploadResponse))

	_, err = receiver.blobClient.DeleteBlob(context.Background(), sourceUrlParts.ContainerName, sourceUrlParts.BlobName, nil)
	if err != nil {
		slog.Error("Error deleting source file after copy", slog.String("source URL", sourceUrl), slog.Any("error", err))
		return err
	}
	return nil
}
