package config

import (
	"encoding/json"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
	"slices"
)

/*
The below struct is the struct for the values of partner configs. If adding new configs add to this struct
*/
type PartnerSettings struct {
	DisplayName              string `json:"displayName"` // full name if we need pretty names
	IsActive                 bool   `json:"isActive"`
	IsExternalSftpConnection bool   `json:"isExternalSftpConnection"`
	HasZipPassword           bool   `json:"hasZipPassword"`
	DefaultEncoding          string `json:"defaultEncoding"`
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
