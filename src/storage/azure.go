package storage

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"io"
	"log/slog"
	"os"
)

type AzureBlobHandler struct {
	blobClient AzureBlobClient
}

type AzureBlobClient interface {
	DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
	UploadBuffer(ctx context.Context, containerName string, blobName string, buffer []byte, o *azblob.UploadBufferOptions) (azblob.UploadBufferResponse, error)
	DeleteBlob(ctx context.Context, containerName string, blobName string, o *azblob.DeleteBlobOptions) (azblob.DeleteBlobResponse, error)
}

func NewAzureBlobHandler() (AzureBlobHandler, error) {
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	blobClient, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return AzureBlobHandler{}, err
	}

	return AzureBlobHandler{blobClient: blobClient}, nil
}

func (receiver AzureBlobHandler) FetchFile(sourceUrl string) ([]byte, error) {
	sourceUrlParts, err := azblob.ParseURL(sourceUrl)
	if err != nil {
		slog.Error("Unable to parse source URL", slog.String("sourceUrl", sourceUrl), slog.Any(utils.ErrorKey, err))
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

func (receiver AzureBlobHandler) UploadFile(fileBytes []byte, blobPath string) error {
	uploadResponse, err := receiver.blobClient.UploadBuffer(context.Background(), utils.ContainerName, blobPath, fileBytes, nil)
	if err != nil {
		slog.Error("Unable to upload file", slog.String("destinationUrl", blobPath), slog.Any(utils.ErrorKey, err))
		return err
	}

	slog.Info("Successfully uploaded file", slog.String("destinationUrl", blobPath), slog.Any("uploadResponse", uploadResponse))

	return nil
}

func (receiver AzureBlobHandler) MoveFile(sourceUrl string, destinationUrl string) error {
	sourceUrlParts, err := azblob.ParseURL(sourceUrl)
	if err != nil {
		slog.Error("Unable to parse source URL", slog.String("sourceUrl", sourceUrl), slog.Any(utils.ErrorKey, err))
		return err
	}

	destinationUrlParts, err := azblob.ParseURL(destinationUrl)
	if err != nil {
		slog.Error("Unable to parse destination URL", slog.String("destinationUrl", destinationUrl), slog.Any(utils.ErrorKey, err))
		return err
	}

	fileBytes, err := receiver.FetchFile(sourceUrl)
	if err != nil {
		slog.Error("Unable to fetch file", slog.String("sourceUrl", sourceUrl), slog.Any(utils.ErrorKey, err))
		return err
	}

	err = receiver.UploadFile(fileBytes, destinationUrlParts.BlobName)
	if err != nil {
		return err
	}

	_, err = receiver.blobClient.DeleteBlob(context.Background(), sourceUrlParts.ContainerName, sourceUrlParts.BlobName, nil)
	if err != nil {
		slog.Error("Error deleting source file after copy", slog.String("source URL", sourceUrl), slog.Any(utils.ErrorKey, err))
		return err
	}

	return nil
}
