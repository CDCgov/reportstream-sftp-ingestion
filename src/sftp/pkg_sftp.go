package sftp

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
)

// TODO - Either Add coverage or exclude and cover with Integration testing.
type PkgSftpImplementation struct {
	client *sftp.Client
}

func NewPkgSftpImplementation(sshClient *ssh.Client) (PkgSftpImplementation, error) {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return PkgSftpImplementation{}, err
	}

	return PkgSftpImplementation{client: sftpClient}, nil
}

func (p PkgSftpImplementation) ReadDir(path string) ([]os.FileInfo, error) {
	return p.client.ReadDir(path)
}

func (p PkgSftpImplementation) Open(path string) (io.ReadCloser, error) {
	file, err := p.client.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (p PkgSftpImplementation) Close() error {
	return p.client.Close()
}

func (p PkgSftpImplementation) Remove(path string) error {
	return p.client.Remove(path)
}
