package usecases

// The BlobHandler interface is about interacting with file data,
// e.g. in Azure Blob Storage or a local filesystem
type BlobHandler interface {
	FetchFile(sourceUrl string) ([]byte, error)
	MoveFile(sourceUrl string, destinationUrl string) error
	UploadFile(fileBytes []byte, blobPath string) error
}
