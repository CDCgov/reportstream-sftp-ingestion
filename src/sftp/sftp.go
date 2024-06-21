package sftp

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"log/slog"
	"os"
)

type SftpClient struct {
	secretGetter secrets.CredentialGetter
}

func (receiver SftpClient) DownloadFile() {
	config := &ssh.ClientConfig{
		User: os.Getenv("SFTP_USER"),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(receiver.secretGetter.GetPrivateKey("")),
			ssh.Password(os.Getenv("SFTP_PASSWORD")),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", os.Getenv("SFTP_SERVER"), config)

	if err != nil {
		slog.Error("Failed to make SSH client")
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal("Failed to make new client ", err)
	}

}

func getKeyFromString() (Signer, error) {

}
