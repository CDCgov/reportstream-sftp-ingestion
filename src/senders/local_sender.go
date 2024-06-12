package senders

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

type FileSender struct {
}

func (receiver FileSender) SendMessage(message []byte) (string, error) {
	folder := "localdata"

	err := os.MkdirAll(folder, 0755)
	if err != nil {
		return "", err
	}

	randomUuid := uuid.NewString()

	filePath := filepath.Join(folder, fmt.Sprintf("%s.txt", randomUuid))
	err = os.WriteFile(filePath, message, 0644) // permissions = owner read/write, group read, other read
	if err != nil {
		return "", err
	}

	return randomUuid, nil
}
