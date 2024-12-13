package mocks

import "github.com/stretchr/testify/mock"

type MockBlobHandler struct {
	mock.Mock
}

func (receiver *MockBlobHandler) FetchFileByUrl(sourceUrl string) ([]byte, error) {
	args := receiver.Called(sourceUrl)
	return args.Get(0).([]byte), args.Error(1)
}

func (receiver *MockBlobHandler) MoveFile(sourceUrl string, destinationUrl string) error {
	args := receiver.Called(sourceUrl, destinationUrl)
	return args.Error(0)
}

func (receiver *MockBlobHandler) UploadFile(fileBytes []byte, blobPath string) error {
	args := receiver.Called(fileBytes, blobPath)
	return args.Error(0)
}
