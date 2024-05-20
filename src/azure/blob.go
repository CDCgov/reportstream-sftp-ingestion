package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"io"
)

//var blobClient *azblob.Client

type BlobHandler struct {
	blobClient *azblob.Client
}

func NewBlobHandler(conn string) (BlobHandler, error) {
	blobClient, err := azblob.NewClientFromConnectionString(conn, nil)
	if err != nil {
		return BlobHandler{}, err
	}

	return BlobHandler{blobClient}, nil
}

//// TODO - better error handling, don't kill the application
//func handleError(err error) {
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//}
//
//// TODO - only initialize client once?
//func InitClient(conn string) error {
//	var err error
//	blobClient, err = azblob.NewClientFromConnectionString(conn, nil)
//	return err
//}
//
//// TODO - container should eventually be managed by Terraform
//func initContainer(client *azblob.Client, ctx context.Context) error {
//	// Create the container
//	containerName := "sftp"
//	fmt.Printf("Creating a container named %s\n", containerName)
//	_, err := client.CreateContainer(ctx, containerName, nil)
//	return err
//}

//func (receiver BlobHandler) UploadFile() error {
//
//	fmt.Println("Hello World, preparing to upload file")
//
//	// Create the container
//	containerName := "sftp"
//	fmt.Printf("Creating a container named %s\n", containerName)
//
//	err := initContainer(receiver.blobClient, context.Background())
//	if err != nil {
//		return err
//	}
//
//	data := []byte("\nHello, BLOB! This is a blob and we ARE BLOBBIN'. THE BLOB does UPDATE TO BLOB.  Blobbity bloobity\nDogCow!\n")
//	blobName := "ca/nbs/results/sample-blob"
//
//	// Upload to data to blob storage
//	fmt.Printf("Uploading a blob named %s\n", blobName)
//	_, err = receiver.blobClient.UploadBuffer(context.Background(), containerName, blobName, data, &azblob.UploadBufferOptions{})
//
//	return err
//}

//func (receiver BlobHandler) listFiles() {
//	containerName := "sftp"
//
//	pager := receiver.blobClient.NewListBlobsFlatPager(containerName, nil)
//	fmt.Println("Listing blob items")
//	for pager.More() {
//		resp, err := pager.NextPage(context.TODO())
//		for _, v := range resp.Segment.BlobItems {
//			fmt.Println(*v.Name)
//			//fmt.Println(*v.Properties)
//		}
//	}
//}

func (receiver BlobHandler) FetchFile(blobPath string) ([]byte, error) {
	// TODO - read containerName from env vars
	containerName := "sftp"

	//var output int64
	streamResponse, err := receiver.blobClient.DownloadStream(context.Background(), containerName, blobPath, &azblob.DownloadStreamOptions{})
	if err != nil {
		return nil, err
	}

	retryReader := streamResponse.NewRetryReader(context.Background(), &azblob.RetryReaderOptions{})
	defer retryReader.Close()
	// GoLand prefers this second option to the line above, we're not yet sure why
	//defer func(retryReader *blob.RetryReader) {
	//	err := retryReader.Close()
	//	handleError(err)
	//}(retryReader)

	return io.ReadAll(retryReader)
}
