package config

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
	"time"
)

func NewConfig(partnerId string) (*Config, error) {
	// Create blob client
	handler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to create Azure Blob handler for config retrieval", slog.Any(utils.ErrorKey, err), slog.String("partnerId", partnerId))
		return nil, err
	}

	// Retrieve settings file from Azure
	fileContents, err := handler.FetchFile("config", partnerId+".json")
	if err != nil {
		slog.Error("Failed to retrieve partner settings", slog.Any(utils.ErrorKey, err), slog.String("partnerId", partnerId))
		return nil, err
	}

	// Parse file content by calling populate
	partnerSettings, err := populatePartnerSettings(fileContents, partnerId)
	if err != nil {
		// We log any errors in the called function
		return nil, err
	}

	// Set up config object
	config := &Config{}
	config.lastRetrieved = time.Now().UTC()
	config.PartnerId = partnerId
	config.partnerSettings = partnerSettings

	return config, nil
}
