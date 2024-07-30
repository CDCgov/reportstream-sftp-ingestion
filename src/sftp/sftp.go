package sftp

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/CDCgov/reportstream-sftp-ingestion/zip"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type SftpHandler struct {
	sshClient        *ssh.Client
	sftpClient       SftpClient
	blobHandler      usecases.BlobHandler
	ioClient         IoClient
	credentialGetter secrets.CredentialGetter
	zipHandler       zip.ZipHandlerInterface
}

type SftpClient interface {
	ReadDir(path string) ([]os.FileInfo, error)
	Open(path string) (*sftp.File, error)
	Close() error
	Remove(path string) error
}

func NewSftpHandler() (*SftpHandler, error) {
	// TODO - pass in info about what customer we're using (and thus what URL/key/password to use)

	credentialGetter, err := secrets.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	pem, err := getPublicKeysForSshClient(credentialGetter)
	if err != nil {
		return nil, err
	}

	serverKeyName := os.Getenv("SFTP_SERVER_PUBLIC_KEY_NAME")

	serverKey, err := credentialGetter.GetSecret(serverKeyName)

	if err != nil {
		slog.Error("Unable to get SFTP_SERVER_PUBLIC_KEY_NAME", slog.String("KeyName", serverKeyName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	hostKeyCallback, err := getSshClientHostKeyCallback(serverKey)
	if err != nil {
		return nil, err
	}

	sftpUserName := os.Getenv("SFTP_USER_NAME")
	sftpUser, err := credentialGetter.GetSecret(sftpUserName)

	if err != nil {
		slog.Error("Unable to get SFTP_USER_NAME", slog.String("KeyName", sftpUserName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sftpPasswordName := os.Getenv("SFTP_PASSWORD_NAME")
	sftpPassword, err := credentialGetter.GetSecret(sftpPasswordName)

	if err != nil {
		slog.Error("Unable to get SFTP_PASSWORD_NAME", slog.String("KeyName", sftpPasswordName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pem),
			ssh.Password(sftpPassword),
		},
		HostKeyCallback: hostKeyCallback,
	}

	sftpServerAddressName := os.Getenv("SFTP_SERVER_ADDRESS_NAME")
	sftpServerAddress, err := credentialGetter.GetSecret(sftpServerAddressName)

	if err != nil {
		slog.Error("Unable to get SFTP_SERVER_ADDRESS_NAME", slog.String("KeyName", sftpServerAddress), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sshClient, err := ssh.Dial("tcp", sftpServerAddress, config)
	if err != nil {
		slog.Error("Failed to make SSH client", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		slog.Error("Failed to make SFTP client ", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	blobHandler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	ioWrapper := IoWrapper{}

	zipHandler, err := zip.NewZipHandler()

	if err != nil {
		slog.Error("Failed to init zip handler", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	return &SftpHandler{
		sshClient:        sshClient,
		sftpClient:       sftpClient,
		blobHandler:      blobHandler,
		ioClient:         &ioWrapper,
		credentialGetter: credentialGetter,
		zipHandler:       zipHandler,
	}, nil
}

func getSshClientHostKeyCallback(serverKey string) (ssh.HostKeyCallback, error) {
	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(serverKey))
	if err != nil {
		slog.Error("Failed to parse authorized key", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	return ssh.FixedHostKey(pk), nil
}

func getPublicKeysForSshClient(credentialGetter secrets.CredentialGetter) (ssh.Signer, error) {
	secretName := os.Getenv("SFTP_KEY_NAME")

	key, err := credentialGetter.GetSecret(secretName)
	if err != nil {
		slog.Error("Unable to retrieve SFTP Key", slog.String("KeyName", secretName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	pem, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		slog.Error("Unable to parse private key", slog.Any(utils.ErrorKey, err))
		return nil, err
	}
	return pem, err
}

func (receiver *SftpHandler) Close() {
	slog.Info("About to close SFTP handler")
	if receiver.sftpClient != nil {
		err := receiver.sftpClient.Close()
		if err != nil {
			slog.Error("Failed to close SFTP client", slog.Any(utils.ErrorKey, err))
		}
	}
	if receiver.sshClient != nil {
		err := receiver.sshClient.Close()
		if err != nil {
			slog.Error("Failed to close SSH client", slog.Any(utils.ErrorKey, err))
		}
	}
	slog.Info("SFTP handler closed")
}

func (receiver *SftpHandler) CopyFiles() {
	sftpStartingDirectoryName := os.Getenv("SFTP_STARTING_DIRECTORY_NAME")
	sftpStartingDirectory, err := receiver.credentialGetter.GetSecret(sftpStartingDirectoryName)

	if err != nil {
		slog.Error("Unable to get SFTP_STARTING_DIRECTORY_NAME", slog.String("KeyName", sftpStartingDirectory), slog.Any(utils.ErrorKey, err))
		return
	}

	fileInfos, err := receiver.sftpClient.ReadDir(sftpStartingDirectory)
	slog.Info("starting directory", slog.String("start dir", sftpStartingDirectory))
	if err != nil {
		slog.Error("Failed to read directory", slog.Any(utils.ErrorKey, err))
		return
	}

	var wg sync.WaitGroup
	//loop through files
	for index, fileInfo := range fileInfos {
		// Increment the wait group counter
		wg.Add(1)
		go func() {
			// Decrement the counter when the go routine completes
			defer wg.Done()
			receiver.copySingleFile(fileInfo, index, sftpStartingDirectory)
		}()
	}
	// Wait for all the wg elements to complete. Otherwise this function will return
	// before all the files are processed, and the SFTP client will close prematurely
	wg.Wait()

	/*
		Eventually:
		- have per-customer config, which contains things like how to connect to external servers (if any) and when,
			plus blob storage folder name
		- replace `files` hard-coded above with a per-customer value for e.g. `sftp_starting_folder` or similar (where
			we go on their external SFTP server to retrieve files)
		- pass customer info to SFTP client, so we know whose files these are/what creds to use
		- since we have customer info, can use that to build destination path for upload
		- have a type or enum or something for allowed destination subfolders? E.g. import, unzip, failure, success, etc.
	*/

}

// copySingleFile moves a single file from an external SFTP server to our blob storage. Zip files go to an `unzip`
// folder and then we call the zipHandler.Unzip. Other files go to `import` to begin processing
func (receiver *SftpHandler) copySingleFile(fileInfo os.FileInfo, index int, directory string) {
	slog.Info("Considering file", slog.String("name", fileInfo.Name()), slog.Int("number", index))
	if fileInfo.IsDir() {
		slog.Info("Skipping directory", slog.String("file name", fileInfo.Name()))
		return
	}

	file, err := receiver.sftpClient.Open(directory + "/" + fileInfo.Name())

	if err != nil {
		slog.Error("Failed to open file", slog.Any(utils.ErrorKey, err))
		return
	}

	slog.Info("file opened", slog.String("name", fileInfo.Name()), slog.Any("file", file))
	fileBytes, err := receiver.ioClient.ReadBytesFromFile(file)
	if err != nil {
		slog.Error("Failed to read file", slog.Any(utils.ErrorKey, err))
		return
	}

	var blobPath string
	if strings.Contains(fileInfo.Name(), ".zip") {
		blobPath = filepath.Join(utils.UnzipFolder, fileInfo.Name())
	} else {
		blobPath = filepath.Join(utils.MessageStartingFolderPath, fileInfo.Name())
	}
	err = receiver.blobHandler.UploadFile(fileBytes, blobPath)
	if err != nil {
		slog.Error("Failed to upload file", slog.Any(utils.ErrorKey, err))
		return
	}

	slog.Info("About to consider whether this is a zip", slog.String("file name", fileInfo.Name()))

	// TODO - if non-CA customers want us to retrieve non-zip files, will need to update this `if`
	if !strings.Contains(fileInfo.Name(), ".zip") {
		slog.Info("Skipping file because it is not a zip file", slog.String("file name", fileInfo.Name()))
		return
	}
	// write file to local filesystem
	err = os.WriteFile(fileInfo.Name(), fileBytes, 0644) // permissions = owner read/write, group read, other read
	if err != nil {
		slog.Error("Failed to write file", slog.Any(utils.ErrorKey, err), slog.String("name", fileInfo.Name()))
		return
	}

	err = receiver.zipHandler.Unzip(fileInfo.Name())
	if err != nil {
		slog.Error("Failed to unzip file", slog.Any(utils.ErrorKey, err))
	}
	// TODO - currently the zip file stays in the `unzip` folder regardless of success, failure, or partial failure.
	// 	Do we want to move the zip somewhere if done?

	//delete file from local filesystem
	err = os.Remove(fileInfo.Name())
	if err != nil {
		slog.Error("Failed to remove file from local server", slog.Any(utils.ErrorKey, err), slog.String("name", fileInfo.Name()))
	}

	err = receiver.sftpClient.Remove(directory + "/" + fileInfo.Name())
	if err != nil {
		slog.Error("Failed to remove file from SFTP server", slog.Any(utils.ErrorKey, err))
		return
	}

	slog.Info("Successfully copied file and removed from SFTP server", slog.Any("file name", fileInfo.Name()))

}

type IoClient interface {
	ReadBytesFromFile(file *sftp.File) ([]byte, error)
}

type IoWrapper struct {
}

func (receiver *IoWrapper) ReadBytesFromFile(file *sftp.File) ([]byte, error) {
	return io.ReadAll(file)
}
