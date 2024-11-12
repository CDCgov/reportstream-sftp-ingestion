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
	"strings"
)

type ZipHandler struct {
	credentialGetter secrets.CredentialGetter
	blobHandler      usecases.BlobHandler
	zipClient        ZipClient
}

type ZipHandlerInterface interface {
	Unzip(zipFilePath string, blobPath string) error
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

// Unzip opens a zip file (applying a password if necessary) and uploads each file within it to the `import` folder
// to begin processing. It collects any errors with individual subfiles and uploads that information as well. An error
// is only returned from the function when we cannot handle the main zip file for some reason or have failed to upload
// the error list about the contents
func (zipHandler ZipHandler) Unzip(zipFileName string, blobPath string) error {
	slog.Info("Preparing to unzip", slog.String("zipFileName", zipFileName))
	zipPasswordSecret := utils.CA_PHL + "-zip-password-" + utils.EnvironmentName() // pragma: allowlist secret
	zipPassword, err := zipHandler.credentialGetter.GetSecret(zipPasswordSecret)

	if err != nil {
		slog.Error("Unable to get zip password", slog.Any(utils.ErrorKey, err), slog.String("KeyName", zipPasswordSecret))

		// move zip file from unzip -> unzip/failure
		zipHandler.MoveZip(blobPath, utils.FailureFolder)
		return err
	}

	zipReader, err := zipHandler.zipClient.OpenReader(zipFileName)

	if err != nil {
		slog.Error("Failed to open zip reader", slog.Any(utils.ErrorKey, err))
		zipHandler.MoveZip(blobPath, utils.FailureFolder)
		return err
	}
	defer zipReader.Close()

	var errorList []FileError

	// loop over contents
	for _, f := range zipReader.File {
		errorList = zipHandler.ExtractAndUploadSingleFile(f, zipPassword, zipFileName, errorList)
	}

	// if errorList has contents -> move zip file from unzip -> unzip/failure
	if len(errorList) > 0 {
		slog.Info("Error list length over zero")
		zipHandler.MoveZip(blobPath, utils.FailureFolder)
	} else {
		// else -> move zip file from unzip -> unzip/success
		slog.Info("Error list length is zero")
		zipHandler.MoveZip(blobPath, utils.SuccessFolder)
	}


	// Upload error info if any
	err = zipHandler.UploadErrorList(blobPath, errorList, err)
	if err != nil {
		return err
	}

	return nil
}

// MoveZip moves a file from 'unzip' into the specified subfolder e.g. 'success', 'failure'
func (zipHandler ZipHandler) MoveZip(blobPath string, subfolder string) {
	slog.Info("About to move file", slog.String("blobPath", blobPath), slog.String("destination subfolder", subfolder))
	// url must include the container name while the blob path does not
	// e.g. when 'sftp' is the container name, the url is 'sftp/unzip/cheeseburger.zip' and the blob path is 'unzip/cheeseburger.zip'
	sourceUrl := filepath.Join(utils.ContainerName, blobPath)
	destinationUrl := strings.Replace(sourceUrl, utils.UnzipFolder, filepath.Join(utils.UnzipFolder, subfolder), 1)
	err := zipHandler.blobHandler.MoveFile(sourceUrl, destinationUrl)
	if err != nil {
		slog.Error("Unable to move file to "+destinationUrl, slog.Any(utils.ErrorKey, err))
	} else {
		slog.Info("Successfully moved file to "+destinationUrl)
	}
}

func (zipHandler ZipHandler) ExtractAndUploadSingleFile(f *zip.File, zipPassword string, zipFilePath string, errorList []FileError) []FileError {
	slog.Info("Extracting file", slog.String(utils.FileNameKey, f.Name), slog.String("zipFilePath", zipFilePath))

	// Apply the partner's Zip password if needed
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

// UploadErrorList takes a list of file-specific errors and uploads them to a single file named after the containing zip
func (zipHandler ZipHandler) UploadErrorList(zipFilePath string, errorList []FileError, err error) error {
	if len(errorList) > 0 {
		fileContents := ""
		for _, fileError := range errorList {

			fileContents += fileError.Filename + ": " + fileError.ErrorMessage + "\n"
		}

		errorDestinationPath := strings.Replace(zipFilePath, utils.UnzipFolder, filepath.Join(utils.UnzipFolder, utils.FailureFolder), 1) + ".txt"
		err = zipHandler.blobHandler.UploadFile([]byte(fileContents), errorDestinationPath)

		if err != nil {
			slog.Error("Failed to upload failure file", slog.Any(utils.ErrorKey, err), slog.String("errorDestinationPath", errorDestinationPath))
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
