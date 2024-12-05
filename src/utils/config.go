package utils

import (
	"encoding/json"
	"log/slog"
	"time"
)

/*
Stuff we need to do in here:
  - constructor
  - constants for things like the file name, container name, look up the env where that matters, etc
  - get config from Azure storage
  - parse config into some kind of friendly object
  - validate config on parsing
  - return config to callers? And/or return specified config value to callers? (i.e. should they request CADPH's on/off
    status, or CADPH's whole config, or all configs for all partners)
  - get config every time we need it? If we get it on app startup, have to restart service to update values. Maybe
    add a retrieval time and have it expire? Getting it once per queue message is better than getting it every time
    we need a value, but will get slow pretty soon
*/

//https://www.notion.so/flexion-cdc-ti/Thinking-about-config-eb9424dafea14320be5cee1b8b03d890?pvs=4

// First stab at the data shape
// All the details each partner should have
// How do we make fields optional? I think in Go this doesn't matter because it'll default
// to null-equivalent, but we'll want to check for valid values
type partnerConfig struct {
	DisplayName        string `json:"displayName"` //unique name, put in queue message from polling function - keep this short so we can use it in TF resources? Currently TF and RS use `ca-phl`
	IsActive           bool   `json:"isActive"`
	SftpConnectionType string `json:"sftpConnectionType"` // either external or internal
	HasZipPassword     bool   `json:"hasZipPassword"`
	DefaultEncoding    string `json:"defaultEncoding"` // e.g. ISO 8859-1"
	CronExpression     string `json:"cronExpression"`  // might just go into the function app itself and not be config at all
	ContainerName      string `json:"containerName"`   // currently all in the same container - we'll need either partner specific subfolders or
	// partner-specific containers (so this should maybe be pathName or folderName)
}

// Map an ID/label to each set of details
type configEntry struct {
	PartnerId     string        `json:"partnerId"`
	PartnerConfig partnerConfig `json:"partnerSettings"`
}

// When did we last get the file and what did we parse out of it?
type config struct {
	lastRetrieved   time.Time
	partnerSettings []configEntry
	// TODO: Move this file so that we can import the blob handler
	//blobHandler     usecases.BlobHandler
}

/*
{
    "ca-phl": { //unique name, put in queue message from polling function - keep this short so we can use it in TF resources? Currently TF and RS use `ca-phl`
        "displayName": "California Department of Public Health",
        "isActive": true,
        "sftpConnectionType": "external",
        "hasZipPassword": true,
        "defaultEncoding": "ISO 8859-1",
        "cronExpression": "blah", // might just go into the function app itself and not be config at all
        "containerName": "blah" // currently all in the same container - we'll need either partner specific subfolders or partner-specific containers
    },
    "osdph": {
        "displayName": "Other State Department of Public Health",
        "isActive": false,
        "sftpConnectionType": "direct", // they sftp to us instead of us to them
        // assume `hasZipPassword` is false unless otherwise set? assume UTF-8 encoding unless otherwise set?
        // cronExpression only needed for external connections
        "containerName": "blah"
    }
}
*/

func (partnerConfig) populatePartnerConfig(input []byte) (partnerConfig, error) {

	jsonData := input

	var partnerConfig partnerConfig

	err := json.Unmarshal(jsonData, &partnerConfig)

	if err != nil {
		slog.Error("Unable unmarshall to partner config", slog.Any(ErrorKey, err))
		return partnerConfig, err
	}

	return partnerConfig, nil

}

func (partnerConfig) populateConfigEntry(input []byte) (configEntry, error) {

	jsonData := input

	var configEntry configEntry

	err := json.Unmarshal(jsonData, &configEntry)

	if err != nil {
		slog.Error("Unable unmarshall to config entry", slog.Any(ErrorKey, err))
		return configEntry, err
	}

	return configEntry, nil

}

//
//func populateConfig() error {
//
//}
