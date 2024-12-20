package orchestration

import (
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

	isActive := checkIsActive(partnerId)
	if !isActive {
		// Return nil here so we'll delete the queue message and they won't pile up during an intentional downtime or misconfiguration
		return nil
	}

	credentialGetter, err := secrets.GetCredentialGetter()
	if err != nil {
		slog.Error("Unable to initialize credential getter", slog.Any(utils.ErrorKey, err))
		return err
	}

	sftpHandler, err := sftp.NewSftpHandler(credentialGetter, partnerId)
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

func checkIsActive(partnerId string) bool {
	isActive := false
	if val, ok := config.Configs[partnerId]; ok {
		isActive = val.PartnerSettings.IsActive
	} else {
		slog.Error("Partner not found in config", slog.String("partnerId", partnerId))
		return false
	}

	if !isActive {
		slog.Warn("Partner not active, skipping", slog.String("partnerId", partnerId))
		return false
	}
	return true
}
