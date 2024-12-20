package config

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

const partnerId = "test"

func Test_populatePartnerSettings_populates(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Test Name",
	"isActive": true,
	"isExternalSftpConnection": true,
	"hasZipPassword": true,
	"defaultEncoding": "ISO-8859-1"
}`)

	partnerSettings, _ := populatePartnerSettings(jsonInput, partnerId)

	assert.Contains(t, partnerSettings.DisplayName, "Test Name")
	assert.Equal(t, true, partnerSettings.IsActive)
	assert.Equal(t, true, partnerSettings.IsExternalSftpConnection)
	assert.Equal(t, true, partnerSettings.HasZipPassword)
}

func Test_populatePartnerSettings_errors_whenJsonInvalid(t *testing.T) {
	jsonInput := []byte(`bad json`)

	_, err := populatePartnerSettings(jsonInput, partnerId)

	assert.Error(t, err)

}

func Test_populatePartnerSettings_errors_whenBooleanInvalid(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Test Name",
	"isActive": "dogcow",
	"isExternalSftpConnection": true,
	"hasZipPassword": true,
	"defaultEncoding": "ISO-8859-1"
}`)

	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	_, err := populatePartnerSettings(jsonInput, partnerId)

	assert.Error(t, err)
	assert.Contains(t, buffer.String(), "cannot unmarshal string into Go struct field PartnerSettings.isActive of type bool")
}

func Test_populatePartnerSettings_errors_whenEncodingInvalid(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Test Name",
	"isActive": true,
	"isExternalSftpConnection": true,
	"hasZipPassword": true,
	"defaultEncoding": "Something else"
}`)

	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	_, err := populatePartnerSettings(jsonInput, partnerId)

	assert.Error(t, err)
	assert.Contains(t, buffer.String(), "Invalid encoding found")
}

func Test_validateDefaultEncoding_errors(t *testing.T) {

	err := validateDefaultEncoding("bad encoding")

	assert.Error(t, err)
}
