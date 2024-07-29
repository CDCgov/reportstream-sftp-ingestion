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
	slog.Info("Handling polling message", slog.String("message text", *message.MessageText))
	// TODO - use the message contents to figure out stuff about config and files
	// SFTP handler has hard-coded details about where to retrieve files from
	sftpHandler, err := sftp.NewSftpHandler()
	if err != nil {
		slog.Error("failed to create sftp handler", slog.Any(utils.ErrorKey, err))
		return err
	}
	defer sftpHandler.Close()

	slog.Info("about to call CopyFiles")
	sftpHandler.CopyFiles()
	slog.Info("called CopyFiles")
	// TODO - have CopyFiles return an error so we can do something smart with it, so we don't
	// 	keep pinging CA. May need to consider the kind of error, in case some situations result in
	// 	a call to CA and some don't
	//  We are pushing this issue to the back board until we have defined error scenarios where we want to retry upon failure vs. waiting
	//  We may want to add a basic error handler to show any error scenarios that come up
	return nil
}
