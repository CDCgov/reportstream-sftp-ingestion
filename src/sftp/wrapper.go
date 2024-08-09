package sftp

import (
	"io"
	"os"
)

type SftpWrapper interface {
	ReadDir(path string) ([]os.FileInfo, error)
	Open(path string) (io.ReadCloser, error)
	Close() error
	Remove(path string) error
}
