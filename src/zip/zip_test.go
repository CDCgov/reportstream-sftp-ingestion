package zip

import (
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/mocks"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yeka/zip"
	"log/slog"
	"path/filepath"
	"testing"
)


var unzipSuccessPath = "unzip/success/cheeseburger.zip"
var unzipFailurePath = "unzip/failure/cheeseburger.zip"
var unzipStartingPath = "unzip/cheeseburger.zip"


func Test_Unzip_FileIsPasswordProtected_UnzipsSuccessfully(t *testing.T) {

	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test123", nil)

	zipPath := filepath.Join("..", "mocks", "test_data", "passworded.zip")
	zipReader, err := zip.OpenReader(zipPath)

	mockZipClient.On("OpenReader", mock.Anything).Return(zipReader, nil)

	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	err = zipHandler.Unzip(unzipStartingPath)

	assert.Contains(t, buffer.String(), "setting password for file")
	assert.Contains(t, buffer.String(), "Extracting file")
	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipSuccessPath)
	assert.NoError(t, err)
}

func Test_Unzip_FileIsNotProtected_UnzipsSuccessfully(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test123", nil)

	zipPath := filepath.Join("..", "mocks", "test_data", "unprotected.zip")
	zipReader, err := zip.OpenReader(zipPath)

	mockZipClient.On("OpenReader", mock.Anything).Return(zipReader, nil)

	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)


	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	err = zipHandler.Unzip(unzipStartingPath)

	assert.NotContains(t, buffer.String(), "setting password for file")
	assert.Contains(t, buffer.String(), "Extracting file")
	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipSuccessPath)
	assert.NoError(t, err)
}

func Test_Unzip_UnableToGetPassword_ReturnsError(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockCredentialGetter.On("GetSecret", mock.Anything).Return("", errors.New("error"))

	mockBlobHandler := new(mocks.MockBlobHandler)
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler: mockBlobHandler,
	}

	err := zipHandler.Unzip(unzipStartingPath)

	assert.NotContains(t, buffer.String(), "setting password for file")
	assert.NotContains(t, buffer.String(), "Extracting file")
	assert.Contains(t, buffer.String(), "Unable to get zip password")
	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipFailurePath)
	assert.Error(t, err)
}

func Test_Unzip_FailsToOpenReader_ReturnsError(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockZipClient := new(MockZipClient)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test123", nil)

	mockZipClient.On("OpenReader", mock.Anything).Return(&zip.ReadCloser{}, errors.New("error"))

	mockBlobHandler := new(mocks.MockBlobHandler)
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler {
		credentialGetter: mockCredentialGetter,
		zipClient:        mockZipClient,
		blobHandler:      mockBlobHandler,
	}

	err := zipHandler.Unzip(unzipStartingPath)

	assert.NotContains(t, buffer.String(), "setting password")
	assert.NotContains(t, buffer.String(), "preparing to process file")
	assert.Contains(t, buffer.String(), "Failed to open zip reader")
	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipFailurePath)
	assert.Error(t, err)
}

func Test_Unzip_FilePasswordIsWrong_UploadsErrorDocument(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockZipClient := new(MockZipClient)
	mockBlobHandler := new(mocks.MockBlobHandler)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test", nil)

	zipPath := filepath.Join("..", "mocks", "test_data", "passworded.zip")
	zipReader, err := zip.OpenReader(zipPath)

	mockZipClient.On("OpenReader", mock.Anything).Return(zipReader, nil)
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		zipClient:        mockZipClient,
		blobHandler:      mockBlobHandler,
	}

	err = zipHandler.Unzip(unzipStartingPath)

	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipFailurePath)
	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, unzipFailurePath+ ".txt")
	assert.Contains(t, buffer.String(), "setting password for file")
	assert.Contains(t, buffer.String(), "Extracting file")
	assert.Contains(t, buffer.String(), "Failed to read message file")
	assert.NoError(t, err)
}

func Test_Unzip_UnzippedFileCannotBeUploaded_ReturnsError(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test123", nil)

	zipPath := filepath.Join("..", "mocks", "test_data", "passworded.zip")
	zipReader, err := zip.OpenReader(zipPath)

	mockZipClient.On("OpenReader", mock.Anything).Return(zipReader, nil)

	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(errors.New("error"))
	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	err = zipHandler.Unzip(unzipStartingPath)

	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipFailurePath)
	assert.Contains(t, buffer.String(), "setting password")
	assert.Contains(t, buffer.String(), "Extracting file")
	assert.Contains(t, buffer.String(), "Failed to upload message file")
	assert.Error(t, err)
}

func Test_MoveZip_MoveZipSuccessful(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	zipHandler.MoveZip(unzipStartingPath, utils.FailureFolder)

	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipFailurePath)
	assert.Empty(t, buffer.String())
}

func Test_MoveZip_MoveZipUnsuccessful_LogsError(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockBlobHandler.On("MoveFile", mock.Anything, mock.Anything).Return(errors.New(""))

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	zipHandler.MoveZip(unzipStartingPath, utils.FailureFolder)

	mockBlobHandler.AssertCalled(t, "MoveFile", mock.Anything, unzipFailurePath)
	assert.Contains(t, buffer.String(), "Unable to move file")
}

type MockZipClient struct {
	mock.Mock
}

func (mockZipClient *MockZipClient) OpenReader(name string) (*zip.ReadCloser, error) {
	args := mockZipClient.Called(name)
	return args.Get(0).(*zip.ReadCloser), args.Error(1)
}
