package utils

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func Test_populateConfig_populates(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Test Name",
	"isActive": true,
	"isExternalSftpConnection": true,
	"hasZipPassword": true,
	"defaultEncoding": "ISO-8859-1"
}`)
	test := Config{}

	_ = test.populatePartnerSettings(jsonInput)

	assert.Contains(t, test.partnerSettings.DisplayName, "Test Name")
	assert.Equal(t, test.partnerSettings.IsActive, true)
	assert.Equal(t, test.partnerSettings.IsExternalSftpConnection, true)
	assert.Equal(t, test.partnerSettings.HasZipPassword, true)
}

func Test_populateConfig_errors_whenJsonInvalid(t *testing.T) {
	jsonInput := []byte(`bad json`)
	test := Config{}

	err := test.populatePartnerSettings(jsonInput)

	assert.Error(t, err)

}

func Test_populateConfig_errors_whenEncodingInvalid(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Test Name",
	"isActive": true,
	"isExternalSftpConnection": true,
	"hasZipPassword": true,
	"defaultEncoding": "Something else"
}`)

	buffer, defaultLogger := SetupLogger()
	defer slog.SetDefault(defaultLogger)

	test := Config{}

	err := test.populatePartnerSettings(jsonInput)

	assert.Error(t, err)
	assert.Contains(t, buffer.String(), "Invalid encoding found")
}

func Test_validateDefaultEncoding_errors(t *testing.T) {

	err := validateDefaultEncoding("bad encoding")

	assert.Error(t, err)
}
