package utils

import (
	"encoding/json"
	"log/slog"
	"time"
)

//https://www.notion.so/flexion-cdc-ti/Thinking-about-config-eb9424dafea14320be5cee1b8b03d890?pvs=4

// All the details each partner should have
// to null-equivalent, but we'll want to check for valid values
type PartnerConfig struct {
	DisplayName        string `json:"displayName"` //unique name, put in queue message from polling function - keep this short so we can use it in TF resources? Currently TF and RS use `ca-phl`
	IsActive           bool   `json:"isActive"`
	SftpConnectionType string `json:"sftpConnectionType"` // either external or internal
	HasZipPassword     bool   `json:"hasZipPassword"`
	ContainerName      string `json:"containerName"` // currently all in the same container - we'll need either partner specific subfolders or
}

// Map an ID/label to each set of details
type configEntry struct {
	PartnerId     string        `json:"partnerId"`
	PartnerConfig PartnerConfig `json:"partnerSettings"`
}

// When did we last get the file and what did we parse out of it?
type config struct {
	lastRetrieved   time.Time
	partnerSettings []configEntry
}

func (PartnerConfig) populatePartnerConfig(input []byte) (PartnerConfig, error) {

	jsonData := input

	var partnerConfig PartnerConfig

	err := json.Unmarshal(jsonData, &partnerConfig)

	if err != nil {
		slog.Error("Unable unmarshall to partner config", slog.Any(ErrorKey, err))
		return partnerConfig, err
	}

	return partnerConfig, nil

}

func (PartnerConfig) populateConfigEntry(input []byte) (configEntry, error) {

	jsonData := input

	var configEntry configEntry

	err := json.Unmarshal(jsonData, &configEntry)

	if err != nil {
		slog.Error("Unable unmarshall to config entry", slog.Any(ErrorKey, err))
		return configEntry, err
	}

	return configEntry, nil

}
