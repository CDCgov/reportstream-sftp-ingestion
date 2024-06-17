package usecases

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

const sourceUrl = "http://localhost/sftp/customer/import/order_message.hl7"
const destinationUrl = "http://localhost/sftp/customer/success/order_message.hl7"

func Test_ReadAndSendUsecase_failsToReadBlob(t *testing.T) {
	mockBlobHandler := &MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte{}, errors.New("it blew up"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}

	err := usecase.ReadAndSend(sourceUrl)

	assert.Error(t, err)
}

func Test_ReadAndSendUsecase_failsToSendMessage(t *testing.T) {
	mockBlobHandler := &MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte("The DogCow went Moof!"), nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("", errors.New("sending message failed"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(sourceUrl)

	assert.Error(t, err)
}

func Test_ReadAndSendUsecase_successfulReadAndSend(t *testing.T) {
	mockBlobHandler := &MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", sourceUrl, destinationUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("epic report ID", nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(sourceUrl)

	assert.NoError(t, err)
}

type MockBlobHandler struct {
	mock.Mock
}

func (receiver *MockBlobHandler) FetchFile(sourceUrl string) ([]byte, error) {
	args := receiver.Called(sourceUrl)
	return args.Get(0).([]byte), args.Error(1)
}

func (receiver *MockBlobHandler) MoveFile(sourceUrl string, destinationUrl string) error {
	args := receiver.Called(sourceUrl, destinationUrl)
	return args.Error(0)
}

type MockMessageSender struct {
	mock.Mock
}

func (receiver *MockMessageSender) SendMessage(message []byte) (string, error) {
	args := receiver.Called(message)
	return args.Get(0).(string), args.Error(1)
}
