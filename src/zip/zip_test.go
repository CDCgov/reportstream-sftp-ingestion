package zip

import (
	"bytes"
	"github.com/CDCgov/reportstream-sftp-ingestion/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yeka/zip"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func Test_Unzip_FileIsPasswordProtected_UnzipsSuccessfully(t *testing.T) {
	os.Setenv("CA_DPH_ZIP_PASSWORD_NAME", "Test")
	defer os.Unsetenv("CA_DPH_ZIP_PASSWORD_NAME")
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test123", nil)

	zipPath := filepath.Join("..", "mocks", "test_data", "passworded.zip")
	zipReader, err := zip.OpenReader(zipPath)

	mockZipClient.On("OpenReader", mock.Anything).Return(zipReader, nil)

	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	err = zipHandler.Unzip("cheezburger")

	assert.Contains(t, buffer.String(), "setting password")
	assert.Contains(t, buffer.String(), "file opened")
	assert.NoError(t, err)
}

func Test_Unzip_FileIsNotProtected_UnzipsSuccessfully(t *testing.T) {
	os.Setenv("CA_DPH_ZIP_PASSWORD_NAME", "Test")
	defer os.Unsetenv("CA_DPH_ZIP_PASSWORD_NAME")
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockBlobHandler := new(mocks.MockBlobHandler)
	mockZipClient := new(MockZipClient)

	mockCredentialGetter.On("GetSecret", mock.Anything).Return("test123", nil)

	zipPath := filepath.Join("..", "mocks", "test_data", "unprotected.zip")
	zipReader, err := zip.OpenReader(zipPath)

	mockZipClient.On("OpenReader", mock.Anything).Return(zipReader, nil)

	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	zipHandler := ZipHandler{
		credentialGetter: mockCredentialGetter,
		blobHandler:      mockBlobHandler,
		zipClient:        mockZipClient,
	}

	err = zipHandler.Unzip("cheezburger")

	assert.NotContains(t, buffer.String(), "setting password")
	assert.Contains(t, buffer.String(), "file opened")
	assert.NoError(t, err)
}

type MockZipClient struct {
	mock.Mock
}

func (mockZipClient MockZipClient) OpenReader(name string) (*zip.ReadCloser, error) {
	args := mockZipClient.Called(name)
	return args.Get(0).(*zip.ReadCloser), args.Error(1)
}
