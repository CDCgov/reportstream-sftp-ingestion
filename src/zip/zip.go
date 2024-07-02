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
}

type ZipClient interface {
}

func NewZipHandler() (ZipHandler, error) {
	// environment := os.Getenv("ENV")

	// TODO - Address local implementation vs test behavior

	//if environment == "local" {
	//	blobHandler, err := storage.NewAzureBlobHandler()
	//}

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
	}, nil
}

// TODO - refactor for tests?
// TODO - tests
// TODO - move remaining items to future cards?
// TODO - check storage size/costs on the container
// TODO - update CA password after deploy per env
func (zipHandler ZipHandler) Unzip(zipFilePath string) error {
	secretName := os.Getenv("CA_DPH_ZIP_PASSWORD_NAME")
	zipPassword, err := zipHandler.credentialGetter.GetSecret(secretName)

	if err != nil {
		slog.Error("Unable to get zip password", slog.Any(utils.ErrorKey, err))
		return err
	}

	slog.Info("Called unzip protected")
	zipReader, err := zip.OpenReader(zipFilePath)

	if err != nil {
		slog.Error("Failed to open zip reader", slog.Any(utils.ErrorKey, err))
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		slog.Info("inside of zip Reader loop")
		// TODO - should we warn or error if not encrypted? This would vary per customer
		if f.IsEncrypted() {
			f.SetPassword(zipPassword)
		}
		slog.Info("setting password")
		fileReader, err := f.Open()
		if err != nil {
			slog.Error("Failed to open file", slog.Any(utils.ErrorKey, err))
		}
		defer fileReader.Close()

		slog.Info("file opened", slog.Any("file", f))
		buf, err := io.ReadAll(fileReader)

		slog.Info(string(buf))

		if err != nil {
			slog.Error("Failed to read file", slog.Any(utils.ErrorKey, err))
		}

		err = zipHandler.blobHandler.UploadFile(buf, utils.MessageStartingFolderPath+"/"+f.FileInfo().Name())

		if err != nil {
			slog.Error("Failed to upload file", slog.Any(utils.ErrorKey, err))
		}

	}
	return nil
}
