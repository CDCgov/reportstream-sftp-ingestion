package utils

import (
	"encoding/json"
	"errors"
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

func (config *Config) populatePartnerSettings(input []byte) error {

	var partnerSettings PartnerSettings
	jsonData := input

	err := json.Unmarshal(jsonData, &partnerSettings)

	if err != nil {
		slog.Error("Unable unmarshall to partner settings", slog.Any(ErrorKey, err))
		return err
	}

	err = validateDefaultEncoding(partnerSettings.DefaultEncoding)
	if err != nil {
		slog.Error("Invalid encoding found", slog.Any(ErrorKey, err), slog.String("Partner ID", config.PartnerId), slog.String("Encoding", partnerSettings.DefaultEncoding))
		return err
	}

	config.partnerSettings = partnerSettings
	return nil

}

func validateDefaultEncoding(input string) error {
	if slices.Contains(allowedEncodingList, input) {
		return nil
	}
	return errors.New("Invalid encoding found: " + input)
}
