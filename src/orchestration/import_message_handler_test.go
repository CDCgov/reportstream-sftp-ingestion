package orchestration

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_HandleMessageContents_MessageHandledSuccessfully_ReturnNil(t *testing.T) {
	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}

	message := createGoodMessage()

	err := importMessageHandler.HandleMessageContents(message)

	assert.NoError(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
}

func Test_HandleMessageContents_FailedToGetFileUrl_DoesNotCallReadAndSend(t *testing.T) {
	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(nil)
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}

	message := createBadMessage()

	err := importMessageHandler.HandleMessageContents(message)

	assert.Error(t, err)
	mockReadAndSendUsecase.AssertNotCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
}

func Test_HandleMessageContents_FailureWithReadAndSend_ReturnsError(t *testing.T) {
	mockReadAndSendUsecase := MockReadAndSendUsecase{}

	mockReadAndSendUsecase.On("ReadAndSend", mock.AnythingOfType("string")).Return(errors.New("failed to read and send"))
	importMessageHandler := ImportMessageHandler{usecase: &mockReadAndSendUsecase}

	message := createGoodMessage()

	err := importMessageHandler.HandleMessageContents(message)

	assert.Error(t, err)
	mockReadAndSendUsecase.AssertCalled(t, "ReadAndSend", mock.AnythingOfType("string"))
}
