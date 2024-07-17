package orchestration

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/sftp"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
)

type PollingMessageHandler struct {
}

func (receiver PollingMessageHandler) HandleMessageContents(message azqueue.DequeuedMessage) error {
	// TODO - use the message contents to figure out stuff about config and files
	sftpHandler, err := sftp.NewSftpHandler()
	if err != nil {
		slog.Error("ope, failed to create sftp handler", slog.Any(utils.ErrorKey, err))
		// Don't return - just because polling is broken for one partner doesn't mean we should take down imports too
	}
	defer sftpHandler.Close()

	sftpHandler.CopyFiles()
	// TODO - have CopyFiles return an error so we can do something smart with it, so we don't
	// 	keep pinging CA. May need to consider the kind of error, in case some situations result in
	// 	a call to CA and some don't
	return nil
}