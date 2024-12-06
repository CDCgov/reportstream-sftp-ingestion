package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_populateConfig_populates(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Test Name",
	"isActive": true,
	"sftpConnectionType": "external",
	"hasZipPassword": true,
	"defaultEncoding": "ISO-8859-1",
	"containerName": "container-name"
}`)
	test := PartnerConfig{}

	output, _ := test.populatePartnerConfig(jsonInput)

	assert.Contains(t, output.DisplayName, "Test Name")
	assert.Equal(t, output.IsActive, true)
	assert.Contains(t, output.SftpConnectionType, "external")
	assert.Equal(t, output.HasZipPassword, true)
	assert.Contains(t, output.ContainerName, "container-name")

}

func Test_populateConfig_errors(t *testing.T) {
	jsonInput := []byte(`bad json`)
	test := PartnerConfig{}

	_, err := test.populatePartnerConfig(jsonInput)

	assert.Error(t, err)

}

func Test_populateConfigEntry_populates(t *testing.T) {
	jsonInput := []byte(`{
	"partnerId": "ca-phl",
	"partnerSettings": {
	"displayName": "Test Name",
	"isActive": true,
	"sftpConnectionType": "external",
	"hasZipPassword": true,
	"defaultEncoding": "ISO-8859-1",
	"containerName": "container-name"
}}`)
	test := PartnerConfig{}

	output, _ := test.populateConfigEntry(jsonInput)

	assert.Contains(t, output.PartnerId, "ca-phl")
	assert.Contains(t, output.PartnerConfig.DisplayName, "Test Name")
	assert.Equal(t, output.PartnerConfig.IsActive, true)
	assert.Contains(t, output.PartnerConfig.SftpConnectionType, "external")
	assert.Equal(t, output.PartnerConfig.HasZipPassword, true)
	assert.Contains(t, output.PartnerConfig.ContainerName, "container-name")
}

func Test_populateConfigEntry_errors(t *testing.T) {
	jsonInput := []byte(`bad json`)
	test := PartnerConfig{}

	_, err := test.populateConfigEntry(jsonInput)

	assert.Error(t, err)
}
