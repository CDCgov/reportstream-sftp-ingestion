package main

import (
	"fmt"
	"github.com/CDCgov/reportstream-sftp-ingestion/azure"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello World")

	//TODO: Extract the client string to allow multi-environment
	blobHandler, err := azure.NewBlobHandler("DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;")
	if err != nil {
		log.Fatalf("Failed to init Azure blob client: %v", err)
	}

	err = readAzureFile(blobHandler, "reportstream.txt")
	if err != nil {
		log.Fatalf("Failed to read the file: %v", err)
	}

	for {
		t := time.Now()
		fmt.Println(t.Format("2006-01-02T15:04:05Z07:00"))
		time.Sleep(10 * time.Second)
	}
}

type BlobHandler interface {
	FetchFile(blobPath string) ([]byte, error)
}

func readAzureFile(blobHandler BlobHandler, filePath string) error {
	content, err := blobHandler.FetchFile(filePath)
	if err != nil {
		return err
	}

	//TODO: Auth and call ReportStream
	log.Println(string(content))

	return nil
}
