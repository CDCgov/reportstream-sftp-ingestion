package usecases

import (
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/mocks"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

const sourceUrl = "http://localhost/sftp/customer/import/order_message.hl7"
const successUrl = "http://localhost/sftp/customer/success/order_message.hl7"
const failureUrl = "http://localhost/sftp/customer/failure/order_message.hl7"

func Test_ReadAndSend_FailsToReadBlob_ReturnsError(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte{}, errors.New("it blew up"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}

	err := usecase.ReadAndSend(sourceUrl)

	assert.Error(t, err)
}

func Test_ReadAndSend_400FromReportStream_MovesFileToFailureFolder(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", sourceUrl, failureUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("", errors.New("400 Bad Request"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(sourceUrl)

	assert.NoError(t, err)
	mockBlobHandler.AssertCalled(t, "MoveFile", sourceUrl, failureUrl)
}

func Test_ReadAndSend_UnexpectedErrorFromReportStream_ReturnsErrorAndDoesNotMoveFile(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", sourceUrl, failureUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("", errors.New("401 Unauthorized"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(sourceUrl)

	assert.Error(t, err)
	mockBlobHandler.AssertNotCalled(t, "MoveFile", sourceUrl, failureUrl)
}

func Test_ReadAndSend_successfulReadAndSend(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", sourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", sourceUrl, successUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("epic report ID", nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(sourceUrl)

	assert.NoError(t, err)
	mockBlobHandler.AssertCalled(t, "MoveFile", sourceUrl, successUrl)
}

func Test_moveFile_UrlMatchesExpectedPattern_UpdatesUrlAndMovesFile(t *testing.T) {
	buffer := utils.SetupLogger()

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("MoveFile", sourceUrl, mock.Anything).Return(nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}
	usecase.moveFile(sourceUrl, "failed")

	assert.NotContains(t, buffer.String(), "Failed to move file after processing")
	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, mock.Anything)
}

func Test_moveFile_SourceUrlDoesNotContainStartingFolder_FileIsNotMoved(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}
	usecase.moveFile("https://example.com/this/that/another", "newFolder")

	mockBlobHandler.AssertNotCalled(t, "MoveFile", mock.Anything, mock.Anything)
}

func Test_moveFile_BlobHandlerFailsToMoveFile_LogsError(t *testing.T) {
	buffer := utils.SetupLogger()

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("MoveFile", sourceUrl, mock.Anything).Return(errors.New("failed to move the file"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}
	usecase.moveFile(sourceUrl, "newFolder")

	assert.Contains(t, buffer.String(), "Failed to move file after processing")
}

type MockMessageSender struct {
	mock.Mock
}

func (receiver *MockMessageSender) SendMessage(message []byte) (string, error) {
	args := receiver.Called(message)
	return args.Get(0).(string), args.Error(1)
}
