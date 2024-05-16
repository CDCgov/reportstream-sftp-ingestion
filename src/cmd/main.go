package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"io"
	"log"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	fmt.Println("Hello World")
	ctx := context.Background()

	//Test commit
	conn := "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
	client, err := azblob.NewClientFromConnectionString(conn, nil)
	if err != nil {
		panic(err)
	}

	// Create the container
	containerName := "sftp"
	fmt.Printf("Creating a container named %s\n", containerName)
	_, err = client.CreateContainer(ctx, containerName, nil)

	data := []byte("\nHello, BLOB! This is a blob and we ARE BLOBBIN'. THE BLOB does UPDATE TO BLOB.  Blobbity bloobity\nDogCow!\n")
	blobName := "ca/nbs/results/sample-blob"

	// Upload to data to blob storage
	fmt.Printf("Uploading a blob named %s\n", blobName)
	_, err = client.UploadBuffer(ctx, containerName, blobName, data, &azblob.UploadBufferOptions{})

	if err != nil {
		panic(err)
	}

	//var output int64
	streamResponse, err := client.DownloadStream(ctx, containerName, blobName, &azblob.DownloadStreamOptions{})

	if err != nil {
		panic(err)
	}
	retryReader := streamResponse.NewRetryReader(ctx, &azblob.RetryReaderOptions{})
	defer retryReader.Close()

	readBytes, err := io.ReadAll(retryReader)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(readBytes))

	pager := client.NewListBlobsFlatPager(containerName, nil)
	fmt.Println("Listing blob items")
	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		handleError(err)
		for _, v := range resp.Segment.BlobItems {
			fmt.Println(*v.Name)
			//fmt.Println(*v.Properties)
		}
	}
}
