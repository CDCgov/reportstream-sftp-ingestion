package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Main(t *testing.T) {
	assert.Equal(t, "DogCow", "DogCow")
}

func Test_readAzureFile_successfullyReadsAndPrints(t *testing.T) {
	err := readAzureFile(TestBlobHandler{}, "dogcow.txt")
	assert.NoError(t, err)
}

func Test_readAzureFile_failsWithError(t *testing.T) {
	err := readAzureFile(TestBlobHandler{errors.New("it blew up")}, "dogcow.txt")
	assert.Error(t, err)
}

type TestBlobHandler struct {
	err error
}

func (receiver TestBlobHandler) FetchFile(blobPath string) ([]byte, error) {
	return []byte{}, receiver.err
}
