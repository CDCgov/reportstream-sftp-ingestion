package config

import (
	"encoding/json"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
	"slices"
	"time"
)

type PartnerSettings struct {
	DisplayName              string `json:"displayName"` // full name if we need pretty names
	IsActive                 bool   `json:"isActive"`
	IsExternalSftpConnection bool   `json:"isExternalSftpConnection"`
	HasZipPassword           bool   `json:"hasZipPassword"`
	DefaultEncoding          string `json:"defaultEncoding"`
}

// When did we last get the file and what did we parse out of it?
type Config struct {
	lastRetrieved   time.Time
	partnerSettings PartnerSettings
	PartnerId       string //unique name, put in queue message from polling function - keep this short so we can use it in TF resources? Currently TF and RS use `ca-phl`
}

// TODO confirm if these should stay here in config or move to constants
var allowedEncodingList = []string{"ISO-8859-1", "UTF-8"}
var KnownPartnerIds = []string{utils.CA_PHL, utils.FLEXION}
var Configs map[string]*Config

func NewConfig(partnerId string) (*Config, error) {
	// Create blob client
	handler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to create Azure Blob handler for config retrieval", slog.Any(utils.ErrorKey, err), slog.String("partnerId", partnerId))
	}

	// Retrieve settings file from Azure
	fileContents, err := handler.FetchFile("config", partnerId+".json")
	if err != nil {
		slog.Error("Failed to retrieve partner settings", slog.Any(utils.ErrorKey, err), slog.String("partnerId", partnerId))
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

func populatePartnerSettings(input []byte, partnerId string) (PartnerSettings, error) {

	var partnerSettings PartnerSettings
	err := json.Unmarshal(input, &partnerSettings)

	if err != nil {
		slog.Error("Unable unmarshall to partner settings", slog.Any(utils.ErrorKey, err))
		return PartnerSettings{}, err
	}

	err = validateDefaultEncoding(partnerSettings.DefaultEncoding)
	if err != nil {
		slog.Error("Invalid encoding found", slog.Any(utils.ErrorKey, err), slog.String("Partner ID", partnerId), slog.String("Encoding", partnerSettings.DefaultEncoding))
		return PartnerSettings{}, err
	}

	// TODO - any other validation?

	return partnerSettings, nil
}

func validateDefaultEncoding(input string) error {
	if slices.Contains(allowedEncodingList, input) {
		return nil
	}
	return errors.New("Invalid encoding found: " + input)
}
