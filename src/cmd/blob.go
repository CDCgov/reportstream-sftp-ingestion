package main

type BlobHandler interface {
	FetchFile(blobPath string) ([]byte, error)
}

/**
Future things to implement:
put file
move file
delete file
*/
