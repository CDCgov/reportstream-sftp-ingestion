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
	sftpClient  SftpClient
	blobHandler usecases.BlobHandler
}

type SftpClient interface {
	ReadDir(path string) ([]os.FileInfo, error)
	Open(path string) (*sftp.File, error)
	Close() error
}

func NewSftpHandler() (*SftpHandler, error) {
	// TODO - pass in info about what customer we're using (and thus what URL/key/password to use)

	credentialGetter, err := utils.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any("error", err))
		return nil, err
	}

	pem, err := getPublicKeysForSshClient(credentialGetter)
	if err != nil {
		return nil, err
	}

	serverKeyName := os.Getenv("SFTP_SERVER_PUBLIC_KEY_NAME")

	serverKey, err := credentialGetter.GetSecret(serverKeyName)

	if err != nil {
		slog.Error("Unable to get SFTP_SERVER_PUBLIC_KEY_NAME", slog.String("KeyName", serverKeyName), slog.Any("error", err))
		return nil, err
	}

	hostKeyCallback, err := getSshClientHostKeyCallback(serverKey)
	if err != nil {
		return nil, err
	}
	slog.Info("Creating SSH client")
	//TODO: Figure out if the ssh client config and the creation of the sftp client should go inside it's own function
	config := &ssh.ClientConfig{
		User: os.Getenv("SFTP_USER"),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pem),
			ssh.Password(os.Getenv("SFTP_PASSWORD")),
		},
		HostKeyCallback: hostKeyCallback,
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
		slog.Error("Failed to init Azure blob client", slog.Any("error", err))
		return nil, err
	}

	return &SftpHandler{
		sshClient:   sshClient,
		sftpClient:  sftpClient,
		blobHandler: blobHandler,
	}, nil
}

func getSshClientHostKeyCallback(serverKey string) (ssh.HostKeyCallback, error) {
	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(serverKey))
	if err != nil {
		slog.Error("Failed to parse authorized key", slog.Any("error", err))
		return nil, err
	}

	return ssh.FixedHostKey(pk), nil
}

func getPublicKeysForSshClient(credentialGetter utils.CredentialGetter) (ssh.Signer, error) {
	secretName := os.Getenv("SFTP_KEY_NAME")

	key, err := credentialGetter.GetSecret(secretName)
	if err != nil {
		slog.Error("Unable to retrieve SFTP Key", slog.String("KeyName", secretName), slog.Any("error", err))
		return nil, err
	}

	pem, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		slog.Error("Unable to parse private key", slog.Any("error", err))
		return nil, err
	}
	return pem, err
}

func (receiver *SftpHandler) Close() {
	err := receiver.sftpClient.Close()
	if err != nil {
		slog.Error("Failed to close SFTP client", slog.Any("error", err))
	}
	err = receiver.sshClient.Close()
	if err != nil {
		slog.Error("Failed to close SSH client", slog.Any("error", err))
	}
	slog.Info("SFTP handler closed")
}

func (receiver *SftpHandler) CopyFiles() {
	// TODO - use "files" for readDir for now, but maybe replace with an env var for whatever directory
	// 	we should start in - this should also be used in sftpClient.open below
	directory := "files"
	//readDir using sftp client
	fileInfos, err := receiver.sftpClient.ReadDir(directory)
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

			file, err := receiver.sftpClient.Open(directory + "/" + fileInfo.Name())

			if err != nil {
				slog.Error("Failed to open file", slog.Any("error", err))
				return
			}
			var read = fileIoWrapper{}
			fileBytes, err := read.ReadBytesFromFile(file)
			if err != nil {
				slog.Error("Failed to read file", slog.Any("error", err))
				return
			}

			// TODO - build a better path (unzip? import? how do we know?)
			err = receiver.blobHandler.UploadFile(fileBytes, fileInfo.Name())
			if err != nil {
				slog.Error("Failed to upload file", slog.Any("error", err))
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

type IoWrapper interface {
	ReadBytesFromFile(file *sftp.File) ([]byte, error)
}

func (receiver SftpHandler) ReadBytesFromFile(file *sftp.File) ([]byte, error) {
	return io.ReadAll(file)
}
