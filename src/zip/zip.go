package zip

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/yeka/zip"
	"io"
	"log/slog"
	"path/filepath"
)

type ZipHandler struct {
	credentialGetter secrets.CredentialGetter
	blobHandler      usecases.BlobHandler
	zipClient        ZipClient
}

type ZipHandlerInterface interface {
	Unzip(zipFilePath string) error
	ExtractAndUploadSingleFile(f *zip.File, zipPassword string, zipFilePath string, errorList []FileError) []FileError
	UploadErrorList(zipFilePath string, errorList []FileError, err error) error
}

type FileError struct {
	Filename     string
	ErrorMessage string
}

func NewZipHandler() (ZipHandler, error) {
	blobHandler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any(utils.ErrorKey, err))
		return ZipHandler{}, err
	}

	credentialGetter, err := secrets.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any(utils.ErrorKey, err))
		return ZipHandler{}, err
	}

	return ZipHandler{
		credentialGetter: credentialGetter,
		blobHandler:      blobHandler,
		zipClient:        ZipClientWrapper{},
	}, nil
}

// TODO - move remaining items to future cards?
// TODO - check storage size/costs on the container (unzipping into memory vs on the directory of the service)
// TODO - update CA password after deploy per env
// TODO - check on visibility timeout for messages (find out default and make sure we can d/l the zip within the timeframe)

// Unzip opens a zip file (applying a password if necessary) and uploads each file within it to the `import` folder
// to begin processing. It collects any errors with individual subfiles and uploads that information as well. An error
// is only returned from the function when we cannot handle the main zip file for some reason or have failed to upload
// the error list about the contents
func (zipHandler ZipHandler) Unzip(zipFilePath string) error {

	slog.Info("Preparing to unzip", slog.String("zipFilePath", zipFilePath))
	zipPasswordSecret := "ca-dph-zip-password-" + utils.EnvironmentName() // pragma: allowlist secret
	zipPassword, err := zipHandler.credentialGetter.GetSecret(zipPasswordSecret)
	if err != nil {
		slog.Error("Unable to get zip password", slog.Any(utils.ErrorKey, err), slog.String("KeyName", zipPasswordSecret))
		return err
	}

	zipReader, err := zipHandler.zipClient.OpenReader(zipFilePath)

	if err != nil {
		slog.Error("Failed to open zip reader", slog.Any(utils.ErrorKey, err))
		return err
	}
	defer zipReader.Close()

	var errorList []FileError

	// loop over contents
	for _, f := range zipReader.File {
		errorList = zipHandler.ExtractAndUploadSingleFile(f, zipPassword, zipFilePath, errorList)
	}
	// Upload error info if any
	err = zipHandler.UploadErrorList(zipFilePath, errorList, err)
	if err != nil {
		return err
	}

	return nil
}

func (zipHandler ZipHandler) ExtractAndUploadSingleFile(f *zip.File, zipPassword string, zipFilePath string, errorList []FileError) []FileError {
	slog.Info("Extracting file", slog.String(utils.FileNameKey, f.Name), slog.String("zipFilePath", zipFilePath))

	// TODO - should we warn or error if not encrypted? This would vary per customer
	if f.IsEncrypted() {
		slog.Info("setting password for file", slog.String(utils.FileNameKey, f.Name), slog.String("zipFilePath", zipFilePath))
		f.SetPassword(zipPassword)
	}

	fileReader, err := f.Open()
	if err != nil {
		slog.Error("Failed to open message file", slog.String(utils.FileNameKey, f.Name), slog.Any(utils.ErrorKey, err), slog.String("zipFilePath", zipFilePath))
		errorList = append(errorList, FileError{Filename: f.Name, ErrorMessage: err.Error()})
		return errorList
	}
	defer fileReader.Close()

	buf, err := io.ReadAll(fileReader)
	if err != nil {
		slog.Error("Failed to read message file", slog.String(utils.FileNameKey, f.Name), slog.Any(utils.ErrorKey, err), slog.String("zipFilePath", zipFilePath))
		errorList = append(errorList, FileError{Filename: f.Name, ErrorMessage: err.Error()})
		return errorList
	}

	err = zipHandler.blobHandler.UploadFile(buf, filepath.Join(utils.MessageStartingFolderPath, f.FileInfo().Name()))

	if err != nil {
		slog.Error("Failed to upload message file", slog.String(utils.FileNameKey, f.Name), slog.Any(utils.ErrorKey, err), slog.String("zipFilePath", zipFilePath))
		errorList = append(errorList, FileError{Filename: f.Name, ErrorMessage: err.Error()})
		return errorList
	}
	slog.Info("uploaded file to blob for import", slog.String(utils.FileNameKey, f.Name), slog.String("zipFilePath", zipFilePath))
	return errorList
}

// uploadErrorList takes a list of file-specific errors and uploads them to a single file named after the containing zip
func (zipHandler ZipHandler) UploadErrorList(zipFilePath string, errorList []FileError, err error) error {
	if len(errorList) > 0 {
		fileContents := ""
		for _, fileError := range errorList {
			fileContents += fileError.Filename + ": " + fileError.ErrorMessage + "\n"
		}

		err = zipHandler.blobHandler.UploadFile([]byte(fileContents), filepath.Join(utils.FailureFolder, zipFilePath+".txt"))
		if err != nil {
			slog.Error("Failed to upload failure file", slog.Any(utils.ErrorKey, err), slog.String("zipFilePath", zipFilePath))
			return err
		}
	}
	return nil
}

type ZipClient interface {
	OpenReader(name string) (*zip.ReadCloser, error)
}

type ZipClientWrapper struct {
}

func (zipClientWrapper ZipClientWrapper) OpenReader(zipFilePath string) (*zip.ReadCloser, error) {
	return zip.OpenReader(zipFilePath)
}
