package sftp

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log/slog"
	"os"
)

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

func (p PkgSftpImplementation) Open(path string) (io.Reader, error) {
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
	cwd, err := p.client.Getwd()
	if err != nil {
		slog.Error("Failed to get current working directory", slog.Any(utils.ErrorKey, err))
	}

	slog.Info("current working directory", slog.Any("cwd", cwd))

	stat, err := p.client.Stat(path)
	if err != nil {
		slog.Error("Failed to stat file", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, path))
	}

	slog.Info("file name", slog.String(utils.FileNameKey, stat.Name()))

	realPath, err := p.client.RealPath(path)
	if err != nil {
		slog.Error("Failed to realpath the path", slog.Any(utils.ErrorKey, err), slog.String(utils.FileNameKey, path))
	}

	slog.Info("real path", slog.String(utils.FileNameKey, realPath))

	return p.client.Remove(path)
}
