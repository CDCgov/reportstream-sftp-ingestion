package usecases

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_ReadAndSendUsecase_failsToReadBlob(t *testing.T) {
	mockBlobHandler := &MockBlobHandler{}
	mockBlobHandler.On("FetchFile", "order_message.hl7").Return([]byte{}, errors.New("it blew up"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}

	err := usecase.ReadAndSend("order_message.hl7")

	assert.Error(t, err)
}

func Test_ReadAndSendUsecase_failsToSendMessage(t *testing.T) {
	mockBlobHandler := &MockBlobHandler{}
	mockBlobHandler.On("FetchFile", "order_message.hl7").Return([]byte("The DogCow went Moof!"), nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("", errors.New("sending message failed"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend("order_message.hl7")

	assert.Error(t, err)
}

func Test_ReadAndSendUsecase_successfulReadAndSend(t *testing.T) {
	mockBlobHandler := &MockBlobHandler{}
	mockBlobHandler.On("FetchFile", "order_message.hl7").Return([]byte("The DogCow went Moof!"), nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("epic report ID", nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend("order_message.hl7")

	assert.NoError(t, err)
}

type MockBlobHandler struct {
	mock.Mock
}

func (receiver *MockBlobHandler) FetchFile(blobPath string) ([]byte, error) {
	args := receiver.Called(blobPath)
	return args.Get(0).([]byte), args.Error(1)
}

type MockMessageSender struct {
	mock.Mock
}

func (receiver *MockMessageSender) SendMessage(message []byte) (string, error) {
	args := receiver.Called(message)
	return args.Get(0).(string), args.Error(1)
}
