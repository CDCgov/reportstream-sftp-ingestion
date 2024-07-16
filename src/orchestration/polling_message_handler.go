package orchestration

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
)

type PollingMessageHandler struct {
	// TODO - add SFTP handler
}

func NewPollingMessageHandler() (PollingMessageHandler, error) {
	// TODO - add SFTP handler

	return PollingMessageHandler{}, nil
}

func (receiver PollingMessageHandler) HandleMessageContents(message azqueue.DequeuedMessage) error {
	// TODO - use SFTP handler
	return nil
}
