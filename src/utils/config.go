package utils

import (
	"encoding/json"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/storage"
	"log/slog"
	"slices"
	"time"
)

var allowedEncodingList = []string{"ISO-8859-1", "UTF-8"}

//https://www.notion.so/flexion-cdc-ti/Thinking-about-config-eb9424dafea14320be5cee1b8b03d890?pvs=4

// All the details each partner should have
// to null-equivalent, but we'll want to check for valid values
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

func (config *Config) retrievePartnerSettings(partnerId string) {
	// Create blob client
	handler, err := storage.NewAzureBlobHandler()
	if err != nil {
		slog.Error("Failed to create Azure Blob handler for config retrieval", slog.Any(ErrorKey, err), slog.String("partnerId", partnerId))
	}

	// TODO build correct filepath based on partnerId
	// FetchFileByUrl takes a blob URL. Do we want to build the URL or make
	// a new version of FetchFileByUrl that takes a container and path?
	fileContents, err := handler.FetchFile(partnerId)
	if err != nil {
		slog.Error("Failed to retrieve partner settings", slog.Any(ErrorKey, err), slog.String("partnerId", partnerId))
	}
	// Retrieve settings file(s)? from Azure
	// Parse file content by calling populate
	err = config.populatePartnerSettings(fileContents)
	if err != nil {
		// We log any errors in the called function
		return
	}

	// TODO confirm is this time setup is correct
	config.lastRetrieved = time.Now().UTC()

	// Somewhere else - call this function, and also validate things?

	// When we call this from main.go, do we want to call it for all
	// known partners (like have a hard coded list of IDs to look up)
	// or just specific ones, or loop through all files?

	// Should this return a config or populate the receiver?

}

func (config *Config) populatePartnerSettings(input []byte) error {

	var partnerSettings PartnerSettings
	err := json.Unmarshal(input, &partnerSettings)

	if err != nil {
		slog.Error("Unable unmarshall to partner settings", slog.Any(ErrorKey, err))
		return err
	}

	err = validateDefaultEncoding(partnerSettings.DefaultEncoding)
	if err != nil {
		slog.Error("Invalid encoding found", slog.Any(ErrorKey, err), slog.String("Partner ID", config.PartnerId), slog.String("Encoding", partnerSettings.DefaultEncoding))
		return err
	}

	// TODO - any other validation?

	config.partnerSettings = partnerSettings
	return nil

}

func validateDefaultEncoding(input string) error {
	if slices.Contains(allowedEncodingList, input) {
		return nil
	}
	return errors.New("Invalid encoding found: " + input)
}
