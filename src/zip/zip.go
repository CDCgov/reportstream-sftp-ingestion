package zip

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/yeka/zip"
	"io"
	"log/slog"
	"os"
)

type ZipHandler struct {
	credentialGetter secrets.CredentialGetter
	blobHandler      usecases.BlobHandler
	zipClient        ZipClient
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

// TODO - refactor for tests?
// TODO - tests
// TODO - move remaining items to future cards?
// TODO - check storage size/costs on the container
// TODO - update CA password after deploy per env
// TODO - check on visibility timeout for messages
func (zipHandler ZipHandler) Unzip(zipFilePath string) error {
	secretName := os.Getenv("CA_DPH_ZIP_PASSWORD_NAME")
	zipPassword, err := zipHandler.credentialGetter.GetSecret(secretName)

	if err != nil {
		slog.Error("Unable to get zip password", slog.Any(utils.ErrorKey, err))
		return err
	}

	zipReader, err := zipHandler.zipClient.OpenReader(zipFilePath)

	if err != nil {
		slog.Error("Failed to open zip reader", slog.Any(utils.ErrorKey, err))
		return err
	}
	defer zipReader.Close()

	// TODO - what if one file succeeds and another fails?
	for _, f := range zipReader.File {
		// TODO - should we warn or error if not encrypted? This would vary per customer
		if f.IsEncrypted() {
			slog.Info("setting password")
			f.SetPassword(zipPassword)
		}

		fileReader, err := f.Open()
		if err != nil {
			slog.Error("Failed to open file", slog.Any(utils.ErrorKey, err))
			return err
		}
		defer fileReader.Close()

		slog.Info("file opened", slog.Any("file", f))

		buf, err := io.ReadAll(fileReader)
		if err != nil {
			slog.Error("Failed to read file", slog.Any(utils.ErrorKey, err))
			return err
		}

		err = zipHandler.blobHandler.UploadFile(buf, utils.MessageStartingFolderPath+"/"+f.FileInfo().Name())

		if err != nil {
			slog.Error("Failed to upload file", slog.Any(utils.ErrorKey, err))
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
