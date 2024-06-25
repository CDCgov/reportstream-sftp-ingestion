package sftp

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/usecases"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"log/slog"
	"os"
)

type SftpHandler struct {
	sshClient   *ssh.Client
	sftpClient  *sftp.Client
	blobHandler usecases.BlobHandler
}

func NewSftpHandler() (*SftpHandler, error) {
	// TODO - pass in info about what customer we're using (and thus what URL/key/password to use)
	secretName := os.Getenv("SFTP_KEY_NAME")

	credentialGetter, err := utils.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.String("error", err.Error()))
		return nil, err
	}

	key, err := credentialGetter.GetSecret(secretName)
	if err != nil {
		slog.Error("Unable to retrieve SFTP Key", slog.String("KeyName", secretName), slog.String("Error", err.Error()))
		return nil, err
	}

	pem, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		slog.Error("Unable to parse private key", slog.String("Error", err.Error()))
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: os.Getenv("SFTP_USER"),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pem),
			ssh.Password(os.Getenv("SFTP_PASSWORD")),
		},
		// TODO - InsecureIgnoreHostKey should not be used in prod code. Need public key for SFTP server?
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClient, err := ssh.Dial("tcp", os.Getenv("SFTP_SERVER_ADDRESS"), config)

	if err != nil {
		slog.Error("Failed to make SSH client", slog.Any("error", err))
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		slog.Error("Failed to make SFTP client ", slog.Any("error", err))
		return nil, err
	}

	blobHandler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to init Azure blob client", slog.String("BlobOpenError", err.Error()))
		return nil, err
	}

	return &SftpHandler{
		sshClient:   sshClient,
		sftpClient:  sftpClient,
		blobHandler: blobHandler,
	}, nil
}

func (receiver *SftpHandler) Close() {
	// TODO - error handling on closes?
	receiver.sftpClient.Close()
	receiver.sshClient.Close()
	slog.Info("SFTP handler closed")
}

func (receiver *SftpHandler) CopyFiles() {
	// TODO - use "files" for readDir for now, but maybe replace with an env var for whatever directory we should start in
	//readDir using sftp client
	fileInfos, err := receiver.sftpClient.ReadDir("files")
	if err != nil {
		log.Fatal("Failed to read directory ", err)
	}

	//loop through files
	for index, fileInfo := range fileInfos {
		go func() {
			slog.Info("Considering file", slog.String("name", fileInfo.Name()), slog.Int("number", index))
			if fileInfo.IsDir() {
				slog.Info("Skipping directory", slog.String("file name", fileInfo.Name()))
				return
			}
			// TODO - create path some better way than this - should match path used in `ReadDir` above
			file, err := receiver.sftpClient.Open("files/" + fileInfo.Name())

			if err != nil {
				slog.Error("Failed to open file", slog.String("FileOpenError", err.Error()))
				return
			}
			fileBytes, err := io.ReadAll(file)

			// TODO - build a better path (unzip? import? how do we know?)
			err = receiver.blobHandler.UploadFile(fileBytes, fileInfo.Name())

			if err != nil {
				slog.Error("Failed to upload file", slog.String("BlobUploadError", err.Error()))
			}
		}()
	}

	/*
		Eventually:
		- have per-customer config, which contains things like how to connect to external servers (if any) and when, plus blob storage folder name
		- pass customer info to SFTP client, so we know whose files these are/what creds to use
		- since we have customer info, can use that to build destination path for upload
		- have a type or enum or something for allowed destination subfolders? E.g. import, unzip, failure, success, etc.
	*/

}
