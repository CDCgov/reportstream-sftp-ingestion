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
	sshClient        *ssh.Client // sshAdapter
	sftpClient       SftpWrapper
	blobHandler      usecases.BlobHandler
	credentialGetter secrets.CredentialGetter
	zipHandler       zip.ZipHandlerInterface
	partnerId        string
}

func NewSftpHandler(credentialGetter secrets.CredentialGetter, partnerId string) (*SftpHandler, error) {
	// Use the partnerId to pull info about what customer we're using (and thus what URL/key/password to use)

	userCredentialPrivateKey, err := getUserCredentialPrivateKey(credentialGetter, partnerId)
	if err != nil {
		slog.Error("Unable to get user credential private key", slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	hostPublicKeyName := partnerId + "-sftp-host-public-key-" + utils.EnvironmentName() // pragma: allowlist secret
	hostPublicKey, err := credentialGetter.GetSecret(hostPublicKeyName)
	if err != nil {
		slog.Error("Unable to get host public key", slog.String("KeyName", hostPublicKeyName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	hostKeyCallback, err := getSshClientHostKeyCallback(hostPublicKey)
	if err != nil {
		slog.Error("Unable construct the host key callback", slog.Any("KeyName", hostPublicKeyName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	sftpUserName := partnerId + "-sftp-user-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpUser, err := credentialGetter.GetSecret(sftpUserName)
	if err != nil {
		slog.Error("Unable to get SFTP username secret", slog.String("KeyName", sftpUserName), slog.Any(utils.ErrorKey, err))
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(userCredentialPrivateKey),
		},
		HostKeyCallback: hostKeyCallback,
	}

	sftpServerAddressName := partnerId + "-sftp-server-address-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpServerAddress, err := credentialGetter.GetSecret(sftpServerAddressName)
	if err != nil {
		slog.Error("Unable to get SFTP server address secret", slog.String("KeyName", sftpServerAddressName), slog.Any(utils.ErrorKey, err))
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
		partnerId:        partnerId,
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

func getUserCredentialPrivateKey(credentialGetter secrets.CredentialGetter, partnerId string) (ssh.Signer, error) {

	userAuthenticationKeyName := partnerId + "-sftp-user-credential-private-key-" + utils.EnvironmentName() // pragma: allowlist secret

	key, err := credentialGetter.GetSecret(userAuthenticationKeyName)
	if err != nil {
		slog.Error("Unable to retrieve user authentication key secret", slog.String("KeyName", userAuthenticationKeyName), slog.Any(utils.ErrorKey, err))
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
	sftpStartingDirectoryName := receiver.partnerId + "-sftp-starting-directory-" + utils.EnvironmentName() // pragma: allowlist secret
	sftpStartingDirectory, err := receiver.credentialGetter.GetSecret(sftpStartingDirectoryName)
	if err != nil {
		slog.Error("Unable to get SFTP starting directory secret", slog.String("KeyName", sftpStartingDirectoryName), slog.Any(utils.ErrorKey, err))
		return
	}

	/*
		TODO -
			- make sure sftp client has partner-specific config set up
			- when we copy files, put those in a partner ID folder
			- after file trigger, parse partner ID out of path
			- then look up config and use that to send to RS
			- report_stream_sender.go and zip.go are the main places still using the utils.CA_PHL constant
			- do we want to check partner flag at that point? Like should the on/off flag apply if they're sending to us or only if we're retrieving from them?
			- do we want to update anything in the timer trigger yet or wait until we've got more than one external partner?
			- update app settings and sticky settings in app.tf? Currently only includes ca phl and not Flexion
			- update secrets.md based on new usage
			- add flexion secrets to key.tf and update values in envs?
	*/

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
	// Upload the retrieved file to either the `unzip` or `import` folder
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

	deleteZip := false
	isZip := strings.Contains(fileInfo.Name(), ".zip")
	if isZip {
		// write file to local filesystem
		zipFileName := fileInfo.Name()
		err = os.WriteFile(zipFileName, fileBytes, 0644) // permissions = owner read/write, group read, other read
		if err != nil {
			slog.Error("Failed to write file", slog.Any(utils.ErrorKey, err), slog.String("name", fileInfo.Name()))
			return
		}

		err = receiver.zipHandler.Unzip(zipFileName, blobPath)
		if err != nil {
			slog.Error("Failed to unzip file", slog.Any(utils.ErrorKey, err))
		} else {
			deleteZip = true
		}

		//delete file from local filesystem
		err = os.Remove(zipFileName)
		if err != nil {
			slog.Error("Failed to remove file from local server", slog.Any(utils.ErrorKey, err), slog.String("name", fileInfo.Name()))
		}
	}

	if !isZip || deleteZip {
		err = receiver.sftpClient.Remove(fullFilePath)
		if err != nil {
			slog.Error("Failed to remove file from SFTP server", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, fullFilePath))
			return
		}
		slog.Info("Successfully copied file and removed from SFTP server", slog.Any(utils.FileNameKey, fullFilePath))
	}
}
