package main

// The BlobHandler interface is about interacting with file data,
// e.g. in Azure Blob Storage or a local filesystem
type BlobHandler interface {
	FetchFile(blobPath string) ([]byte, error)
}
