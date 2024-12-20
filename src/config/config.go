package config

import (
	"encoding/json"
	"errors"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
	"slices"
)

type PartnerSettings struct {
	DisplayName              string `json:"displayName"` // full name if we need pretty names
	IsActive                 bool   `json:"isActive"`
	IsExternalSftpConnection bool   `json:"isExternalSftpConnection"`
	HasZipPassword           bool   `json:"hasZipPassword"`
	DefaultEncoding          string `json:"defaultEncoding"`
}

/*
TODO list as of Dec 20:
Current PR:
- In polling message handler, use queue message to:
	- decide whether to do retrieval ('no' for flexion probs) (done)
	- build key names for retrieving secrets (done)
	- add config to tests

Future PR:
- Set up config files in prod
- Set up another function trigger/CRON for Flexion
- In polling message handler, use queue message to:
	- build file paths for saving files (both zips and hl7s)
- In import message handler:
	- parse file path to get partner ID
	- use partner ID to build key names for retrieving secrets to call RS
	- add config to tests
- See if we need to do add'l TF to set up Flexion?
	- Probably need to set up all secrets in same pattern as the `ca-phl` ones
	- probably at least a cron expression and RS config. It would be nice to have an external Flexion SFTP site to hit for testing
	- Do we want to start making TF dynamic at this point or wait for add'l partners? I think maybe wait for 1-2 more partners?
*/

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
