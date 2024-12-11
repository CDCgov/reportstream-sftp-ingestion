package orchestration

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/sftp"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
)

type PollingMessageHandler struct {
}

func (receiver PollingMessageHandler) HandleMessageContents(message azqueue.DequeuedMessage) error {
	slog.Info("Handling polling message", slog.String("message text", *message.MessageText))

	// In future, we will use the message contents to figure out stuff about config and files
	// SFTP handler currently has hard-coded details about where to retrieve files from
	credentialGetter, err := secrets.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any(utils.ErrorKey, err))
		return err
	}

	sftpHandler, err := sftp.NewSftpHandler(credentialGetter)
	if err != nil {
		slog.Error("failed to create sftp handler", slog.Any(utils.ErrorKey, err))
		return err
	}
	defer sftpHandler.Close()

	// We don't collect errors from `CopyFiles`, so the queue message will be deleted if we reach this step
	// regardless of whether it succeeds.
	// Any files that didn't get copied will be picked up on the next scheduled polling event
	// Once we see any real errors, we may revisit this
	slog.Info("about to call CopyFiles")
	sftpHandler.CopyFiles()
	slog.Info("called CopyFiles")

	return nil
}
