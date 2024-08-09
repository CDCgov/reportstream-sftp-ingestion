package sftp

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/CDCgov/reportstream-sftp-ingestion/zip"
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
	sftpClient       SftpWrapper
	blobHandler      usecases.BlobHandler
	credentialGetter secrets.CredentialGetter
	zipHandler       zip.ZipHandlerInterface
}

func NewSftpHandler() (*SftpHandler, error) {
	// TODO - pass in info about what customer we're using (and thus what URL/key/password to use)

	credentialGetter, err := secrets.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	// TODO uncomment code when partner is setup to receive key
	//pem, err := getPublicKeysForSshClient(credentialGetter)
	//if err != nil {
	//	return nil, err
	//}

	serverKeySecret := "sftp-server-public-key-" + utils.EnvironmentName() // pragma: allowlist secret
	serverKey, err := credentialGetter.GetSecret(serverKeySecret)
	if err != nil {
		slog.Error("Unable to get server key secret", slog.String("KeyName", serverKeySecret), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	hostKeyCallback, err := getSshClientHostKeyCallback(serverKey)
	if err != nil {
		return nil, err
	}

	sftpUserNameSecret := "sftp-user-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpUser, err := credentialGetter.GetSecret(sftpUserNameSecret)
	if err != nil {
		slog.Error("Unable to get SFTP username secret", slog.String("KeyName", sftpUserNameSecret), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sftpPasswordSecret := "sftp-password-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpPassword, err := credentialGetter.GetSecret(sftpPasswordSecret)
	if err != nil {
		slog.Error("Unable to get SFTP password secret", slog.String("KeyName", sftpPasswordSecret), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	// TODO uncomment code when partner is setup to receive key
	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			//ssh.PublicKeys(pem),
			ssh.Password(sftpPassword),
		},
		HostKeyCallback: hostKeyCallback,
	}

	sftpServerAddressSecret := "sftp-server-address-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpServerAddress, err := credentialGetter.GetSecret(sftpServerAddressSecret)
	if err != nil {
		slog.Error("Unable to get SFTP server address secret", slog.String("KeyName", sftpServerAddressSecret), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sshClient, err := ssh.Dial("tcp", sftpServerAddress, config)
	if err != nil {
		slog.Error("Failed to make SSH client", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sftpClient, err := NewPkgSftpImplementation(sshClient)
	if err != nil {
		slog.Error("Failed to make SFTP client", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	blobHandler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	zipHandler, err := zip.NewZipHandler()

	if err != nil {
		slog.Error("Failed to init zip handler", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	return &SftpHandler{
		sshClient:        sshClient,
		sftpClient:       sftpClient,
		blobHandler:      blobHandler,
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

	userAuthenticationKeySecret := "sftp-key-" + utils.EnvironmentName() // pragma: allowlist secret

	key, err := credentialGetter.GetSecret(userAuthenticationKeySecret)
	if err != nil {
		slog.Error("Unable to retrieve user authentication key secret", slog.String("KeyName", userAuthenticationKeySecret), slog.Any(utils.ErrorKey, err))
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
	sftpStartingDirectorySecret := "sftp-starting-directory-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpStartingDirectory, err := receiver.credentialGetter.GetSecret(sftpStartingDirectorySecret)
	if err != nil {
		slog.Error("Unable to get SFTP starting directory secret", slog.String("KeyName", sftpStartingDirectorySecret), slog.Any(utils.ErrorKey, err))
		return
	}

	slog.Info("starting directory", slog.String("start dir", sftpStartingDirectory))

	fileInfos, err := receiver.sftpClient.ReadDir(sftpStartingDirectory)
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
	slog.Info("Considering file", slog.String(utils.FileNameKey, fileInfo.Name()), slog.Int("number", index))
	if fileInfo.IsDir() {
		slog.Info("Skipping directory", slog.String(utils.FileNameKey, fileInfo.Name()))
		return
	}

	fullFilePath := directory + "/" + fileInfo.Name()

	fileReadCloser, err := receiver.sftpClient.Open(fullFilePath)

	if err != nil {
		slog.Error("Failed to open file", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, fullFilePath))
		return
	}

	slog.Info("file opened", slog.String(utils.FileNameKey, fullFilePath))

	fileBytes, err := io.ReadAll(fileReadCloser)
	if err != nil {
		slog.Error("Failed to read file", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, fullFilePath))
		return
	}

	err = fileReadCloser.Close()
	if err != nil {
		slog.Error("Failed to close file after reading", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, fullFilePath))
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

	slog.Info("About to consider whether this is a zip", slog.String(utils.FileNameKey, fileInfo.Name()))

	// TODO - if non-CA customers want us to retrieve non-zip files, will need to update this `if`
	if !strings.Contains(fileInfo.Name(), ".zip") {
		slog.Info("Skipping file because it is not a zip file", slog.String(utils.FileNameKey, fileInfo.Name()))
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

	err = receiver.sftpClient.Remove(fullFilePath)
	if err != nil {
		slog.Error("Failed to remove file from SFTP server", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, fullFilePath))
		return
	}

	slog.Info("Successfully copied file and removed from SFTP server", slog.Any(utils.FileNameKey, fullFilePath))
}
