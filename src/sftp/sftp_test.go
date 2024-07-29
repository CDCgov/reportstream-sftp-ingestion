package sftp

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/mocks"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/CDCgov/reportstream-sftp-ingestion/zip"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	yekazip "github.com/yeka/zip"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func Test_getSshClientHostKeyCallback_ReturnsFixedHostKeyCallback(t *testing.T) {
	serverKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDg90HXaJnI1KtfJp8MWHxAwC00PvQCZKm4FRRdPGhEMepXIeLdjOtZV6LdePMT3WUmNkd6vaJ4EEmFUtH9lKLidALL9blOJF1iZKXK81JBJsds8axz5cqAau6aclgc9B1z2tAa+JtaSqN7uXvfPsrmsVss4jcOxX+thAhz7U6chN6ahabgIPqHBEjwvPlVNNbSqv0Q0eS4WaEEo/39tiXn5DYpPRC6DjuZ3m5s3VIgHznTv2Ufp3kcLcfEDZFwjm5XRWLNNvM5h3aW1vmr4lgBwuEzPV7CYIdIyDxe9V7YYcGfO+uu/VrDpY1wSmcD3lzHLLTbi5WWOurwiMsWIVRZfa/rmzuoTYknd5iJoiTyIWmR7L0FLfzPlDYJZmAWSdLZrZaUdD8SDIoKMSEV/5/ZzcI0wuoknis+zpyFqT0jfOy7E4GtG8pEQf7JGXaiExNd9TKxbRmaxp3Yv4WgPBThY39Va7EMUC/s0hX2Ah8pIWZG4Lze4x7Z4dElCOHDgnsl3Akc399jnIDfUY4bVn+rfBJntx9mBRaNnV1GqRodbSkHK5dTcZEmRslhuhsQVO2CxrlkPhFEe0XXpA3llO9YIkf4sCZDUbRFKPJiHyDhfrf2/HzkLndODdFaAnICYd51zOI1SgP3aFx60bZ2nPSoLs9DsR1LLIpz4uoiy5hCHw== sschuresko@flexion-mac-J40DPF4YQR"
	actualParsedKeyCallback, err := getSshClientHostKeyCallback(serverKey)

	assert.NotNil(t, actualParsedKeyCallback)
	assert.NoError(t, err)
}

func Test_getSshClientHostKeyCallback_UnableToParseServerKey_ReturnsError(t *testing.T) {
	serverKey := "AAAAB3NzaC1yc2EAAAADAQABAAACAQDg90HXaJnI1KtfJp8MWHxAwC00PvQCZKm4FRRdPGhEMepXIeLdjOtZV6LdePMT3WUmNkd6vaJ4EEmFUtH9lKLidALL9blOJF1iZKXK81JBJsds8axz5cqAau6aclgc9B1z2tAa+JtaSqN7uXvfPsrmsVss4jcOxX+thAhz7U6chN6ahabgIPqHBEjwvPlVNNbSqv0Q0eS4WaEEo/39tiXn5DYpPRC6DjuZ3m5s3VIgHznTv2Ufp3kcLcfEDZFwjm5XRWLNNvM5h3aW1vmr4lgBwuEzPV7CYIdIyDxe9V7YYcGfO+uu/VrDpY1wSmcD3lzHLLTbi5WWOurwiMsWIVRZfa/rmzuoTYknd5iJoiTyIWmR7L0FLfzPlDYJZmAWSdLZrZaUdD8SDIoKMSEV/5/ZzcI0wuoknis+zpyFqT0jfOy7E4GtG8pEQf7JGXaiExNd9TKxbRmaxp3Yv4WgPBThY39Va7EMUC/s0hX2Ah8pIWZG4Lze4x7Z4dElCOHDgnsl3Akc399jnIDfUY4bVn+rfBJntx9mBRaNnV1GqRodbSkHK5dTcZEmRslhuhsQVO2CxrlkPhFEe0XXpA3llO9YIkf4sCZDUbRFKPJiHyDhfrf2/HzkLndODdFaAnICYd51zOI1SgP3aFx60bZ2nPSoLs9DsR1LLIpz4uoiy5hCHw== sschuresko@flexion-mac-J40DPF4YQR"
	actualParsedKeyCallback, err := getSshClientHostKeyCallback(serverKey)

	assert.Nil(t, actualParsedKeyCallback)
	assert.Error(t, err)
}

func Test_getPublicKeysForSshClient_ReturnsPem(t *testing.T) {
	os.Setenv("SFTP_KEY_NAME", "sftp_server_user_id_rsa.pem")
	defer os.Unsetenv("SFTP_KEY_NAME")

	secretValue := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAACFwAAAAdzc2gtcn\nNhAAAAAwEAAQAAAgEAumdM026JYzIrA3aNXWY4o6SMcxRyIxmzU8ySo21iuT7NAbuPJXmJ\nyjw6WaMlIktUT1r/bV+/bOV41yNFiYUld7ZB6xIiBEESf7iNZYp3kboNvRI9gQiHtlYV+d\nawQwFb35w+0mlvjR2faSCdFPs6p6GiZdn9k1qG+CewSB9UbqG4kUV385vKke4zDe7EH8g9\nvLPWosYIqEkgHAjPwEArc9izuXTCR2Dsl0xLfwcNc8Xf/Su77Id/55yNIGr8gRBGPjtiwW\nBMN0PSyV109yyDBq6vjeDDZ9SHKSoErYnhFHnTkjprIlgR9/5jSVCBpr8eIdo1iRuLLzLh\nmQ8DZN+y7OsAwlJc1kEa5U4ubwmFxMqoCNRPBhqXdm+LDIx+7slEvHoOJPqMuiF+e3THpM\nk5vAwITlBVZtj3I/qkap2MR6lg+zkdW2cW8Ml+VxCmWT+sykoNR4DkNM7H9wi0wAwT74zG\nlQ72YSwvoaWMc3VSYPMpaVaJV+jhujBGUV3E2Ay9LfdR1oWZPJQs/RI0WuZOZkczv7sNR6\nNvHLl6VIsHcnvYY+prmFmEwJ+bHsysVsp7m/In46GLgZr73MSDznJntxRvPF+NVH58MtbP\n3i9IECBCH0BCG0waYQooKM1grdf3+da8ZA+tbakRcPjO89Gn+65jvBUM8+8VJt8jNA6tcm\nUAAAdIBj7vigY+74oAAAAHc3NoLXJzYQAAAgEAumdM026JYzIrA3aNXWY4o6SMcxRyIxmz\nU8ySo21iuT7NAbuPJXmJyjw6WaMlIktUT1r/bV+/bOV41yNFiYUld7ZB6xIiBEESf7iNZY\np3kboNvRI9gQiHtlYV+dawQwFb35w+0mlvjR2faSCdFPs6p6GiZdn9k1qG+CewSB9UbqG4\nkUV385vKke4zDe7EH8g9vLPWosYIqEkgHAjPwEArc9izuXTCR2Dsl0xLfwcNc8Xf/Su77I\nd/55yNIGr8gRBGPjtiwWBMN0PSyV109yyDBq6vjeDDZ9SHKSoErYnhFHnTkjprIlgR9/5j\nSVCBpr8eIdo1iRuLLzLhmQ8DZN+y7OsAwlJc1kEa5U4ubwmFxMqoCNRPBhqXdm+LDIx+7s\nlEvHoOJPqMuiF+e3THpMk5vAwITlBVZtj3I/qkap2MR6lg+zkdW2cW8Ml+VxCmWT+sykoN\nR4DkNM7H9wi0wAwT74zGlQ72YSwvoaWMc3VSYPMpaVaJV+jhujBGUV3E2Ay9LfdR1oWZPJ\nQs/RI0WuZOZkczv7sNR6NvHLl6VIsHcnvYY+prmFmEwJ+bHsysVsp7m/In46GLgZr73MSD\nznJntxRvPF+NVH58MtbP3i9IECBCH0BCG0waYQooKM1grdf3+da8ZA+tbakRcPjO89Gn+6\n5jvBUM8+8VJt8jNA6tcmUAAAADAQABAAACAQCAqZTJy95g7dvqxAXHlis6KPYY6N/vgmnZ\nSbddvr8KBmMS8xdXUpDdWr0b6hRTm5NSQwlTwWcsDyhdtybkSVcXTmIpk5aPQSs3pXdTw0\nPM/pNFEjYJvo2OOdVpYdrAJUv5CKwEKGqrCOtjcPN76/0Mf/DMRK9W6oGHAD4ZSibJRi9T\ndpPZPouQNs5eq5QMK/cRLUDVkcOgBPl44Ewl8yULDWTgecsv4aLsu+jQgVmzs71rzqgkF1\nMd111CJxarL0SM6Ai+WW3CJ7py62M0yTCXiDP8xkuae4Pf0fTwo98MdxqmMFSKnCeq+Zgm\nnr8fDYQK8cdKIAzuQzycnVRGaHHjEIQSAVv3qfxzb2lk8qCB2NTGvjfMFITJoKYyPWb7Jj\nb41EPk8NZGqOVch5a44vvrHYsuNwdk40+YtNodQ0DREDTtvplAUcmSwZrIACj7I/PsYRZx\nWCiSlJ6UxpdBbFJ7HpTDwlPQMkUzmxVQzg+abtqI7mZPomS/EZ2xtNpwm98p0pyell9wGw\nsiZBi6Mt6iPsKDdQTK6XbTZnYnLuIzXcpSJ/gTAavvyn3D5Up2LUU/NmTpUsuqoTz/VSjb\ntlVaDiz2nmem3zvC1t01PV+aSo39Wg4AMG2moEo/buZhfAqXUMz1XmYJ3js0fY0HUoq4RS\nqfd90aWhqmQmcTbpkScQAAAQEAuMHiHwAGi34hprW61Qxmu1b3XlicLW0kSP5qjn4l6kM9\nT44J1KSbU90Bs/FZq9GPfYazHEPf4j1BleHEOcsOTLf86rfkJHtuebU1Lelv3nGGytlfc/\ny6NVXTQXdG/RDIHec3LXX6D/zqfYQPbG4T8flWJ5c7/JtVScflhRp1SmjesoHkgq83eZI9\nY0j9W8CLA/LrMQnEq8SzL1p+Cj2n2aIwZhX9hS/VkQFDmvZ0w9Z4rxNRnsMxwabni1A84g\nP7qDZltTpJZLZ9BRlhP9hkqmO8tlDH2Lj1j8DaxlUlPNVzJTUY+SjctE/eLvYSWduUHJ9w\npgZvfwzVfoRd67T0ZwAAAQEA4hzssuT34awOuP6SCg6tshu9ORfDmSiZHdolnNcOpe9GZm\ncg/aR4RcPjrpeQxEIEjlEBbvyXu+G5A3rr+SCnBduD0szzpAkVkAAy3+Tat9iNhrPxD5bU\nTc1VSaSiAln533cdgqBRAXp7zU5vXhD3DA1cWmhjoLnkggfp96kX9z13zw66n7IiQF9BDW\ns1AuUGhjFxtxXvkdncS4EjijwSSCSMu3ttEwpXrQXJjmbER5GkxEIX1jJTLgCukzEsAFG8\nwDVTBxB3QNi+luucoKRyzZlf2fc+m529M+QnVCxWu4ElQsssexDEX/mGdYU9IIDhP9KaRA\nRQ/OZX9/8tAPCHqwAAAQEA0wq11SyeNXx67U63Go2iQnTkKWIqjdVIuQd4vgdmXiHglmBE\nxTmd7VFNBZ7Waje4y7WmMVYdoCAlyOYpKGdwGX5HjE3r4D60HN7+zOYxSdUBUCJWykER1Y\nVjQxSwnSkh4Xdil3QK7Ql1nYRfNSgOwMHd5RyBglSC88eh2vtH5FU8OafzBYmfDkSAdyy2\n5vX83kv5oMUoliJuyFSz7b/AF3b+OAxVxwQfy1J+2ufErRbxNIePfc/EhoSD0MxZD8SebR\nZG0RV/SBTxh5UMmFKqx5OsXJuG7WRmuqqY8+LHDy0JtcKYeEYkSuX2u4JeY1xrcyVU9jM/\nx02R0p/Ln1ueLwAAAA1tZUBoYWxwcmluLmlvAQIDBA==\n-----END OPENSSH PRIVATE KEY-----\n" //pragma: allowlist secret

	mockCredentialGetter := new(MockCredentialGetter)
	mockCredentialGetter.On("GetSecret", mock.Anything).Return(secretValue, nil)

	pem, err := getPublicKeysForSshClient(mockCredentialGetter)

	mockCredentialGetter.AssertCalled(t, "GetSecret", mock.Anything)
	assert.NotNil(t, pem)
	assert.NoError(t, err)
}

func Test_getPublicKeysForSshClient_UnableToRetrieveSFTPKey_ReturnsError(t *testing.T) {
	os.Setenv("SFTP_KEY_NAME", "sftp_server_user_id_rsa.pem")
	defer os.Unsetenv("SFTP_KEY_NAME")

	mockCredentialGetter := new(MockCredentialGetter)
	mockCredentialGetter.On("GetSecret", mock.Anything).Return("", errors.New(utils.ErrorKey))

	pem, err := getPublicKeysForSshClient(mockCredentialGetter)

	mockCredentialGetter.AssertCalled(t, "GetSecret", mock.Anything)
	assert.Nil(t, pem)
	assert.Error(t, err)
}

func Test_getPublicKeysForSshClient_UnableToParsePrivateKey_ReturnsError(t *testing.T) {
	os.Setenv("SFTP_KEY_NAME", "sftp_server_user_id_rsa.pem")
	defer os.Unsetenv("SFTP_KEY_NAME")

	secretValue := "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\n-----END OPENSSH PRIVATE KEY-----\n" //pragma: allowlist secret

	mockCredentialGetter := new(MockCredentialGetter)
	mockCredentialGetter.On("GetSecret", mock.Anything).Return(secretValue, nil)

	pem, err := getPublicKeysForSshClient(mockCredentialGetter)

	mockCredentialGetter.AssertCalled(t, "GetSecret", mock.Anything)
	assert.Nil(t, pem)
	assert.Error(t, err)
}

func Test_CopyFiles_SuccessfullyCopiesFiles(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	var files []os.FileInfo
	filePath := filepath.Join("..", "..", "mock_data", "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)
	files = append(files, fileInfo)
	fileBytes, _ := os.ReadFile(filePath)

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("ReadDir", mock.Anything).Return(files, nil)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)
	mockSftpClient.On("Remove", mock.Anything).Return(nil)

	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockCredentialGetter.On("GetSecret", mock.Anything).Return("dogcow", nil)

	mockZipHandler := &MockZipHandler{}
	mockZipHandler.On("Unzip", mock.Anything).Return(nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, credentialGetter: mockCredentialGetter, zipHandler: mockZipHandler}

	sftpHandler.CopyFiles()

	mockSftpClient.AssertCalled(t, "ReadDir", mock.Anything)
	assert.NotContains(t, buffer.String(), "Failed to read directory ")
}

func Test_CopyFiles_FailsToReadDirectory_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	var files []os.FileInfo
	filePath := filepath.Join("..", "..", "mock_data", "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)
	files = append(files, fileInfo)

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("ReadDir", mock.Anything).Return(files, errors.New(utils.ErrorKey))

	mockCredentialGetter := new(mocks.MockCredentialGetter)
	mockCredentialGetter.On("GetSecret", mock.Anything).Return("dogcow", nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, credentialGetter: mockCredentialGetter}

	sftpHandler.CopyFiles()

	mockSftpClient.AssertCalled(t, "ReadDir", mock.Anything)
	assert.Contains(t, buffer.String(), "Failed to read directory")
}

func Test_copySingleFile_CopiesFile(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)
	mockSftpClient.On("Remove", mock.Anything).Return(nil)

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)
	fileBytes, _ := os.ReadFile(filePath)

	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	mockZipHandler := &MockZipHandler{}
	mockZipHandler.On("Unzip", mock.Anything).Return(nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, zipHandler: mockZipHandler}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, mock.Anything)
	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	assert.Contains(t, buffer.String(), "Considering file")
	assert.NotContains(t, buffer.String(), "Skipping directory")
	assert.NotContains(t, buffer.String(), "Failed to open file")
	assert.NotContains(t, buffer.String(), "Failed to read file")
	assert.NotContains(t, buffer.String(), "Failed to upload file")
	assert.Contains(t, buffer.String(), "Successfully copied file and removed from SFTP server")
}

func Test_copySingleFile_SkipsDirectory_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)

	fileDirectory := filepath.Join("..", "..", "mock_data")
	fileInfo, _ := os.Stat(fileDirectory)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	assert.Contains(t, buffer.String(), "Skipping directory")
}

func Test_copySingleFile_FailsToOpenFile_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, errors.New(utils.ErrorKey))

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	assert.Contains(t, buffer.String(), "Failed to open file")
}

func Test_copySingleFile_FailsToReadFile_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)

	var emptyBytes []byte
	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(emptyBytes, errors.New(utils.ErrorKey))

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, ioClient: mockIoWrapper}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	assert.Contains(t, buffer.String(), "Failed to read file")
}

func Test_copySingleFile_FileIsNotZipped_LogFileIsSkipped(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)
	mockSftpClient.On("Remove", mock.Anything).Return(nil)

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt")
	fileInfo, _ := os.Stat(filePath)
	fileBytes, _ := os.ReadFile(filePath)

	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	mockZipHandler := &MockZipHandler{}
	mockZipHandler.On("Unzip", mock.Anything).Return(nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, zipHandler: mockZipHandler}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, mock.Anything)
	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	assert.Contains(t, buffer.String(), "Considering file")
	assert.NotContains(t, buffer.String(), "Skipping directory")
	assert.NotContains(t, buffer.String(), "Failed to open file")
	assert.NotContains(t, buffer.String(), "Failed to read file")
	assert.NotContains(t, buffer.String(), "Failed to upload file")
	assert.Contains(t, buffer.String(), "Skipping file because it is not a zip file")
}

func Test_copySingleFile_FailsToUploadFile_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)
	fileBytes, _ := os.ReadFile(filePath)

	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(errors.New(utils.ErrorKey))

	mockZipHandler := &MockZipHandler{}
	mockZipHandler.On("Unzip", mock.Anything).Return(nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, zipHandler: mockZipHandler}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, mock.Anything)
	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	assert.Contains(t, buffer.String(), "Failed to upload file")
}

func Test_copySingleFile_FailsToUnzipFile_LogsError(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)
	mockSftpClient.On("Remove", mock.Anything).Return(nil)

	mockZipHandler := &MockZipHandler{}
	mockZipHandler.On("Unzip", mock.Anything).Return(errors.New("fails to unzip file"))

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)

	fileBytes, _ := os.ReadFile(filePath)
	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, zipHandler: mockZipHandler}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, mock.Anything)
	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	mockSftpClient.AssertCalled(t, "Remove", mock.Anything)
	assert.Contains(t, buffer.String(), "Failed to unzip file")
	assert.Contains(t, buffer.String(), "Successfully copied file and removed from SFTP server")
}

func Test_copySingleFile_FailsToDeleteFileFromSFTPServer_LogsErrorAndReturn(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)

	mockIoWrapper := new(MockIoWrapper)

	mockZipHandler := &MockZipHandler{}

	mockZipHandler.On("Unzip", mock.Anything).Return(nil)

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)

	fileBytes, _ := os.ReadFile(filePath)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}

	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	mockSftpClient.On("Remove", mock.Anything).Return(errors.New("failed to remove file from sftp server"))

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, zipHandler: mockZipHandler}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, mock.Anything)
	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	mockSftpClient.AssertCalled(t, "Remove", mock.Anything)
	assert.Contains(t, buffer.String(), "Failed to remove file from SFTP server")
	assert.NotContains(t, buffer.String(), "Successfully copied file and removed from SFTP server")
}

func Test_copySingleFile_DeletesFileFromSFTPServer(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	mockSftpClient := new(MockSftpClient)
	mockSftpClient.On("Open", mock.Anything).Return(&sftp.File{}, nil)
	mockSftpClient.On("Remove", mock.Anything).Return(nil)

	mockZipHandler := &MockZipHandler{}
	mockZipHandler.On("Unzip", mock.Anything).Return(nil)

	fileDirectory := filepath.Join("..", "..", "mock_data")
	filePath := filepath.Join(fileDirectory, "copy_file_test.txt.zip")
	fileInfo, _ := os.Stat(filePath)

	fileBytes, _ := os.ReadFile(filePath)
	mockIoWrapper := new(MockIoWrapper)
	mockIoWrapper.On("ReadBytesFromFile", mock.Anything).Return(fileBytes, nil)

	mockBlobHandler := &mocks.MockBlobHandler{}
	mockBlobHandler.On("UploadFile", mock.Anything, mock.Anything).Return(nil)

	sftpHandler := SftpHandler{sftpClient: mockSftpClient, blobHandler: mockBlobHandler, ioClient: mockIoWrapper, zipHandler: mockZipHandler}
	sftpHandler.copySingleFile(fileInfo, 1, fileDirectory)

	mockBlobHandler.AssertCalled(t, "UploadFile", mock.Anything, mock.Anything)
	mockIoWrapper.AssertCalled(t, "ReadBytesFromFile", mock.Anything)
	mockSftpClient.AssertCalled(t, "Open", mock.Anything)
	mockSftpClient.AssertCalled(t, "Remove", mock.Anything)
	assert.NotContains(t, buffer.String(), "Failed to remove file from SFTP server")
	assert.Contains(t, buffer.String(), "Successfully copied file and removed from SFTP server")
}

// Mocks for test
type MockCredentialGetter struct {
	mock.Mock
}

func (receiver *MockCredentialGetter) GetSecret(secretName string) (string, error) {
	args := receiver.Called(secretName)
	return args.Get(0).(string), args.Error(1)
}

func (receiver *MockCredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	args := receiver.Called(privateKeyName)
	return args.Get(0).(*rsa.PrivateKey), args.Error(1)
}

type MockSftpClient struct {
	mock.Mock
}

func (receiver *MockSftpClient) ReadDir(path string) ([]os.FileInfo, error) {
	args := receiver.Called(path)
	return args.Get(0).([]os.FileInfo), args.Error(1)
}

func (receiver *MockSftpClient) Open(path string) (*sftp.File, error) {
	args := receiver.Called(path)
	return args.Get(0).(*sftp.File), args.Error(1)
}

func (receiver *MockSftpClient) Close() error {
	args := receiver.Called()
	return args.Error(0)
}

func (receiver *MockSftpClient) Remove(path string) error {
	args := receiver.Called(path)
	return args.Error(0)
}

type MockIoWrapper struct {
	mock.Mock
}

func (receiver *MockIoWrapper) ReadBytesFromFile(file *sftp.File) ([]byte, error) {
	args := receiver.Called(file)
	return args.Get(0).([]byte), args.Error(1)
}

type MockZipHandler struct {
	mock.Mock
}

func (receiver *MockZipHandler) Unzip(zipFilePath string) error {
	args := receiver.Called(zipFilePath)
	return args.Error(0)
}

func (receiver *MockZipHandler) ExtractAndUploadSingleFile(f *yekazip.File, zipPassword string, errorList []zip.FileError) []zip.FileError {
	args := receiver.Called(f, zipPassword, errorList)
	return args.Get(0).([]zip.FileError)
}

func (receiver *MockZipHandler) UploadErrorList(zipFilePath string, errorList []zip.FileError, err error) error {
	args := receiver.Called(zipFilePath, errorList, err)
	return args.Error(0)
}
