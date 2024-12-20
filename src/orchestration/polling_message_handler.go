package orchestration

import (
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/CDCgov/reportstream-sftp-ingestion/config"
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/sftp"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
)

type PollingMessageHandler struct {
}

func (receiver PollingMessageHandler) HandleMessageContents(message azqueue.DequeuedMessage) error {
	slog.Info("Handling polling message", slog.String("message text", *message.MessageText))
	partnerId := *message.MessageText

	isActive := false
	// TODO get config for partner ID in message
	if val, ok := config.Configs[partnerId]; ok {
		isActive = val.PartnerSettings.IsActive
	} else {
		// TODO - this will cause the queue message to retry. Is that what we want? Or should we log an error and return nil (deletes message)
		return errors.New("Partner not found in config: " + partnerId)
	}

	if !isActive {
		slog.Warn("Partner not active, skipping", slog.String("partnerId", partnerId))
		// Return nil here so we'll delete the queue message and they won't pile up during an intentional downtime
		return nil
	}

	credentialGetter, err := secrets.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any(utils.ErrorKey, err))
		return err
	}

	// TODO pass partner ID into sftp handler
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
