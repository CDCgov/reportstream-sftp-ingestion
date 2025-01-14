package config

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
	"time"
)

type Config struct {
	// PartnerId is a unique name to identify a partner. It's put in queue message from polling function and used in blob paths
	PartnerId       string
	lastRetrieved   time.Time
	PartnerSettings PartnerSettings
}

// TODO confirm if these should stay here in config or move to constants
var allowedEncodingList = []string{"ISO-8859-1", "UTF-8"}
var KnownPartnerIds = []string{utils.CA_PHL, utils.FLEXION}
var Configs = make(map[string]*Config)

func init() {
	for _, partnerId := range KnownPartnerIds {
		partnerConfig, err := NewConfig(partnerId)
		if err != nil {
			// TODO - add an ADR talking about this. We're not crashing if a single config doesn't load in case only one partner is impacted
			slog.Error("Unable to load or parse config", slog.Any(utils.ErrorKey, err), slog.String("partner", partnerId))
		}

		Configs[partnerId] = partnerConfig
		slog.Info("config found", slog.String("Id", partnerId))

	}
}

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
	partnerSettings, err := populatePartnerS
	ettings(fileContents, partnerId)
	if err != nil {
		// We log any errors in the called function
		return nil, err
	}

	// Set up config object
	config := &Config{}
	config.lastRetrieved = time.Now().UTC()
	config.PartnerId = partnerId
	config.PartnerSettings = partnerSettings

	return config, nil
}
