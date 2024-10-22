package usecases

import (
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/mocks"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func Test_ReadAndSend_FailsToReadBlob_ReturnsError(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", utils.SourceUrl).Return([]byte{}, errors.New("it blew up"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}

	err := usecase.ReadAndSend(utils.SourceUrl)

	assert.Error(t, err)
}

func Test_ReadAndSend_NonTransientFailureFromReportStream_MovesFileToFailureFolder(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", utils.SourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", utils.SourceUrl, utils.FailureSourceUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("", errors.New(utils.ReportStreamNonTransientFailure))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(utils.SourceUrl)

	assert.NoError(t, err)
	mockBlobHandler.AssertCalled(t, "MoveFile", utils.SourceUrl, utils.FailureSourceUrl)
}

func Test_ReadAndSend_UnexpectedErrorFromReportStream_ReturnsErrorAndDoesNotMoveFile(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", utils.SourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", utils.SourceUrl, utils.FailureSourceUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("", errors.New("401 Unauthorized"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(utils.SourceUrl)

	assert.Error(t, err)
	mockBlobHandler.AssertNotCalled(t, "MoveFile", utils.SourceUrl, utils.FailureSourceUrl)
}

func Test_ReadAndSend_successfulReadAndSend(t *testing.T) {
	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("FetchFile", utils.SourceUrl).Return([]byte("The DogCow went Moof!"), nil)
	mockBlobHandler.On("MoveFile", utils.SourceUrl, utils.SuccessSourceUrl).Return(nil)

	mockMessageSender := &MockMessageSender{}
	mockMessageSender.On("SendMessage", mock.Anything).Return("epic report ID", nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler, messageSender: mockMessageSender}

	err := usecase.ReadAndSend(utils.SourceUrl)

	assert.NoError(t, err)
	mockBlobHandler.AssertCalled(t, "MoveFile", utils.SourceUrl, utils.SuccessSourceUrl)
}

func Test_ConvertToUtf8_ConvertsSuccessfully_ReturnsEncodedContent(t *testing.T) {
	usecase := ReadAndSendUsecase{}
	originalContent, _ := os.ReadFile(filepath.Join("..", "..", "mock_data", "ISO-8859-1.hl7"))
	// The mu character is a single byte (0xb5 in hex or 181 in decimal) in the western ISO 8859-1 encoding
	// In UTF-8, it's two bytes (0xc2 0xb5 in hex or 194 181 in decimal)
	utfMu := string([]byte{194, 181})
	westernMu := string([]byte{181})

	encodedContent, err := usecase.ConvertToUtf8(originalContent)

	assert.NoError(t, err)
	assert.NotEqual(t, originalContent, encodedContent)

	// Since the byte form of the UTF mu contains the byte form of the western mu and
	// Go checks `Contains` using bytes, we can't assert that the encoded content
	// doesn't contain the western mu (because byte-wise, it does)
	assert.Contains(t, string(encodedContent), utfMu)
	assert.Contains(t, string(originalContent), westernMu)
	assert.NotContains(t, string(originalContent), utfMu)
}

func Test_ConvertToUtf8_SourceDataIsNotWesternEncoded_GarblesContent(t *testing.T) {
	usecase := ReadAndSendUsecase{}
	originalContent := "µmol/L"
	doubleEncoded := "Âµmol/L"

	encodedContent, err := usecase.ConvertToUtf8([]byte(originalContent))

	assert.NoError(t, err)
	assert.NotEqual(t, originalContent, encodedContent)
	assert.Equal(t, doubleEncoded, string(encodedContent))
}

func Test_moveFile_UrlMatchesExpectedPattern_UpdatesUrlAndMovesFile(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("MoveFile", utils.SourceUrl, mock.Anything).Return(nil)

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}
	usecase.moveFile(utils.SourceUrl, "failed")

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
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("MoveFile", utils.SourceUrl, mock.Anything).Return(errors.New("failed to move the file"))

	usecase := ReadAndSendUsecase{blobHandler: mockBlobHandler}
	usecase.moveFile(utils.SourceUrl, "newFolder")

	assert.Contains(t, buffer.String(), "Failed to move file after processing")
}

type MockMessageSender struct {
	mock.Mock
}

func (receiver *MockMessageSender) SendMessage(message []byte) (string, error) {
	args := receiver.Called(message)
	return args.Get(0).(string), args.Error(1)
}
